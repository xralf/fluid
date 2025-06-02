package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"slices"

	"github.com/xralf/fluid/pkg/model"
)

const URL = "http://localhost:50001/api/v1/"

var (
	logger *slog.Logger
)

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	logger.Info("Client says welcome!")
}

func main() {
	usage := fmt.Sprintf("usage: %v (start <example folder> | stop (all | <job ID>))", os.Args[0])

	if len(os.Args) < 3 {
		fmt.Print(usage)
		return
	}
	option := os.Args[1]

	allowedOptions := []string{"start", "stop"}
	if !slices.Contains(allowedOptions, option) {
		fmt.Print(usage)
		return
	}

	switch option {
	case "start":
		path := os.Args[2]
		demoStart(path)
	case "stop":
		s := os.Args[2]
		if s == "all" {
			demoStopAll()
		} else {
			jobId := s
			stopJob(jobId)
		}
	default:
		panic(fmt.Errorf("unknown option: %s", option))
	}
}

func demoStart(path string) {
	catalogId := addCatalog(path)
	queryId := addQuery(path)
	spoutId := addSpout(path)
	prepId := addPrep(path)
	jobId := addJob(catalogId, queryId, spoutId, prepId)
	fmt.Printf("JOB ID: %s\n", jobId)
	url := startJob(jobId)
	fmt.Printf("DASHBOARD URL: %s\n", url)
}

func demoStopAll() {
	jobIds := getJobIds()
	if len(jobIds) == 0 {
		panic(fmt.Errorf("cannot find any job"))
	}
	for _, id := range jobIds {
		stopJob(id)
	}
}

/*
func stopJob(id string) {
	url := URL + "job/stop"

	jobRequest := model.StopJobRequest{
		Id: id,
	}

	var data []byte
	var err error
	if data, err = json.Marshal(jobRequest); err != nil {
		log.Fatal(err)
	}
	var res *http.Response
	if res, err = http.Post(url, "application/json", bytes.NewReader(data)); err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	obj := model.StopResponse{}
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&obj)
	fmt.Printf("OBJ: %v", obj)
	fmt.Printf("ID: %v, message: %s\n", obj.Id, obj.Message)
	return
}
*/

func getJobIds() (ids []string) {
	url := URL + "job/list"

	var res *http.Response
	var err error

	if res, err = http.Get(url); err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	var jobsResponse model.ListJobsResponse
	if err = json.NewDecoder(res.Body).Decode(&jobsResponse); err != nil {
		log.Fatal(err)
	}

	for _, job := range jobsResponse.Jobs {
		ids = append(ids, job.Id)
	}
	return
}

func addCatalog(path string) (catalogId string) {
	url := URL + "catalog/add"
	catalogId = uploadFile(path+"/catalog.json", url)
	return
}

func addQuery(path string) (queryId string) {
	url := URL + "query/add"
	queryId = uploadFile(path+"/query.fql", url)
	return
}

func addSpout(path string) (spoutId string) {
	url := URL + "spout/add"
	spoutId = uploadFile(path+"/spout.cmd", url)
	logger.Info(
		"@@@1",
		"path", path,
		"spoutId", spoutId,
		"url", url,
	)
	return
}

func addPrep(path string) (prepId string) {
	url := URL + "prep/add"
	prepId = uploadFile(path+"/prep.sh", url)
	return
}

func addJob(catalogId string, queryId string, spoutId string, prepId string) (jobId string) {
	url := URL + "job/add"

	jobRequest := model.AddJobRequest{
		CatalogId: catalogId,
		QueryId:   queryId,
		SpoutId:   spoutId,
		PrepId:    prepId,
	}
	var data []byte
	var err error
	if data, err = json.Marshal(jobRequest); err != nil {
		log.Fatal(err)
	}

	var res *http.Response
	if res, err = http.Post(url, "application/json", bytes.NewReader(data)); err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	obj := model.AddResponse{}
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&obj)
	jobId = obj.Id
	fmt.Printf("ID: %v\n", jobId)
	return
}

func startJob(jobId string) (dashboardURL string) {
	url := URL + "job/start"

	jobRequest := model.StartJobRequest{
		Id: jobId,
	}

	var data []byte
	var err error
	if data, err = json.Marshal(jobRequest); err != nil {
		log.Fatal(err)
	}
	var res *http.Response
	if res, err = http.Post(url, "application/json", bytes.NewReader(data)); err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	obj := model.StartResponse{}
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&obj)
	fmt.Printf("URL: %v, message: %s\n", obj.DashboardURL, obj.Message)
	return
}

func stopJob(jobId string) {
	url := URL + "job/stop"

	jobRequest := model.StopJobRequest{
		Id: jobId,
	}

	var data []byte
	var err error
	if data, err = json.Marshal(jobRequest); err != nil {
		log.Fatal(err)
	}
	var res *http.Response
	if res, err = http.Post(url, "application/json", bytes.NewReader(data)); err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	obj := model.StopResponse{}
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&obj)
	fmt.Printf("ID: %v, message: %s\n", obj.Id, obj.Message)
}

func uploadFile(filePath string, url string) (id string) {
	logger.Debug(
		"@@@",
		"filePath", filePath,
		"url", url,
	)

	var body bytes.Buffer
	var writer *multipart.Writer
	var err error
	if body, writer, err = createMultipartFormData("file", filePath); err != nil {
		panic(err)
	}
	var req *http.Request
	if req, err = http.NewRequest("POST", url, &body); err != nil {
		panic(err)
	}

	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", writer.FormDataContentType())

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return
	}
	if res.StatusCode != http.StatusOK {
		fmt.Printf("bad status: %s", res.Status)
		return
	}

	obj := model.AddResponse{}
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&obj)
	id = obj.Id
	fmt.Printf("ID: %v, message: %s\n", id, obj.Message)
	res.Body.Close()
	return
}

func createMultipartFormData(fieldName, fileName string) (buffer bytes.Buffer, writer *multipart.Writer, err error) {
	var file *os.File
	if file, err = os.Open(fileName); err != nil {
		err = fmt.Errorf("error opening file: %v", err)
		return
	}
	writer = multipart.NewWriter(&buffer)
	var fw io.Writer
	if fw, err = writer.CreateFormFile(fieldName, file.Name()); err != nil {
		err = fmt.Errorf("error creating writer: %v", err)
		return
	}
	if _, err = io.Copy(fw, file); err != nil {
		err = fmt.Errorf("error with io.Copy: %v", err)
		return
	}
	writer.Close()
	return buffer, writer, nil
}

// // Person is a struct that represents the data we will send in the request body
// type Person struct {
// 	Name string
// 	Age  int
// }

// func main() {
// 	url := "http://localhost:3000"

// 	// create post body using an instance of the Person struct
// 	p := Person{
// 		Name: "John Doe",
// 		Age:  25,
// 	}
// 	// convert p to JSON data
// 	jsonData, err := json.Marshal(p)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// We can set the content type here
// 	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer resp.Body.Close()

// 	fmt.Println("Status:", resp.Status)
// }
