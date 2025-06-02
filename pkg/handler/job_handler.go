package handler

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/xralf/fluid/cmd/utils"
	"github.com/xralf/fluid/pkg/model"
)

const (
	WebSocketURLPrefix = "ws://"

	WebSocketClientBinary         = "demo-pipe-client"
	WebSocketServerBinary         = "demo-pipe-server"
	DashboardBinary               = "demo-web-server"
	DemoFinancialDataServerBinary = "demo-findata-server"
	DemoSyslogBinary              = "syslog"
	DemoThrottleBinary            = "throttle"
	DemoGeneratorBinary           = "generator"
	DashboardTemplateFile         = "dashboard-template.html"

	// FilePermissionReadable   = 0644
	// FilePermissionExecutable = 0755
	FilePermissionReadable   = 0666
	FilePermissionExecutable = 0777
)

var (
	logger                 *slog.Logger
	templatesDirectoryPath string
	demoDirectoryPath      string
)

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: false, //true,
		Level:     slog.LevelDebug,
	}))
	logger.Info("Handler says welcome!")

	demoDirectoryPath = filepath.Join("/", "tmp", "demo")
	//templatesDirectoryPath = filepath.Join("/", "tmp", "repos", "fluid", "templates")
	templatesDirectoryPath = filepath.Join("/", "tmp", "demo", "templates")
}

type JobHandler interface {
	Add(*gin.Context)
	Delete(*gin.Context)
	List(*gin.Context)
	Start(*gin.Context)
	Stop(*gin.Context)
}

type EngineCompartment struct {
	JobId string
}

type jobHandler struct {
	cfg utils.Configuration
}

func NewJobHandler(cfg utils.Configuration) JobHandler {
	return &jobHandler{
		cfg: cfg,
	}
}

func (app *jobHandler) List(c *gin.Context) {
	_, ctxErr := context.WithTimeout(c.Request.Context(), time.Duration(app.cfg.App.Timeout)*time.Second)
	defer ctxErr()

	jobs := make([]model.Job, 0, len(app.cfg.App.Jobs))
	for _, v := range app.cfg.App.Jobs {
		jobs = append(jobs, v)
	}

	res := model.ListJobsResponse{
		Jobs: jobs,
	}
	c.JSON(http.StatusOK, res)
}

func (app *jobHandler) Start(c *gin.Context) {
	_, ctxErr := context.WithTimeout(c.Request.Context(), time.Duration(app.cfg.App.Timeout)*time.Second)
	defer ctxErr()

	req := model.StartRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	fmt.Printf("START 1: %v\n", req)

	var job model.Job
	var ok bool
	if job, ok = app.cfg.App.Jobs[req.Id]; !ok {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cannot find job with ID %s", req.Id))
		return
	}

	url := "http://localhost:" + strconv.Itoa(job.DashboardPort)
	fmt.Printf("Dashboard URL: %s\n", url)

	var statusMsg string
	var err error
	if statusMsg, err = startEngine(job); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	res := model.StartResponse{
		DashboardURL: url,
		Message:      statusMsg,
	}
	c.JSON(http.StatusOK, res)
}

func (app *jobHandler) Stop(c *gin.Context) {
	_, ctxErr := context.WithTimeout(c.Request.Context(), time.Duration(app.cfg.App.Timeout)*time.Second)
	defer ctxErr()

	req := model.StopRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var job model.Job
	var ok bool
	if job, ok = app.cfg.App.Jobs[req.Id]; !ok {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cannot find job with ID %s", req.Id))
		return
	}

	var statusMsg string
	var err error
	if statusMsg, err = stopEngine(job.JobDirectoryPath); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	res := model.StopResponse{
		Id:      job.Id,
		Message: statusMsg,
	}
	c.JSON(http.StatusOK, res)
}

func (app *jobHandler) Delete(c *gin.Context) {
	_, ctxErr := context.WithTimeout(c.Request.Context(), time.Duration(app.cfg.App.Timeout)*time.Second)
	defer ctxErr()

	req := model.DeleteRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var job model.Job
	var ok bool
	if job, ok = app.cfg.App.Jobs[req.Id]; !ok {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cannot find job with ID %s", req.Id))
	}

	// Return the reserved ports back to the pool.
	app.cfg.App.AvailablePorts = append(app.cfg.App.AvailablePorts, job.Pipe1IngressPort)
	app.cfg.App.AvailablePorts = append(app.cfg.App.AvailablePorts, job.Pipe1EgressPort)
	app.cfg.App.AvailablePorts = append(app.cfg.App.AvailablePorts, job.Pipe2IngressPort)
	app.cfg.App.AvailablePorts = append(app.cfg.App.AvailablePorts, job.Pipe2EgressPort)
	app.cfg.App.AvailablePorts = append(app.cfg.App.AvailablePorts, job.DashboardPort)

	delete(app.cfg.App.Jobs, req.Id)
	c.JSON(http.StatusOK, "success")
}

func (app *jobHandler) Add(c *gin.Context) {
	_, ctxErr := context.WithTimeout(c.Request.Context(), time.Duration(app.cfg.App.Timeout)*time.Second)
	defer ctxErr()

	jobRequest := model.AddJobRequest{}
	if err := c.ShouldBindBodyWith(&jobRequest, binding.JSON); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	id := uuid.NewString()
	path := createJobDirectory(id)

	if len(app.cfg.App.AvailablePorts) < 5 {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("not enough ports available"))
		return
	}

	pipe1IngressPort := app.cfg.App.AvailablePorts[0]
	app.cfg.App.AvailablePorts = app.cfg.App.AvailablePorts[1:]
	pipe1EgressPort := app.cfg.App.AvailablePorts[0]
	app.cfg.App.AvailablePorts = app.cfg.App.AvailablePorts[1:]
	pipe2IngressPort := app.cfg.App.AvailablePorts[0]
	app.cfg.App.AvailablePorts = app.cfg.App.AvailablePorts[1:]
	pipe2EgressPort := app.cfg.App.AvailablePorts[0]
	app.cfg.App.AvailablePorts = app.cfg.App.AvailablePorts[1:]
	dashboardPort := app.cfg.App.AvailablePorts[0]
	app.cfg.App.AvailablePorts = app.cfg.App.AvailablePorts[1:]

	job := model.Job{
		Id:                    id,
		QueryId:               jobRequest.QueryId,
		CatalogId:             jobRequest.CatalogId,
		SpoutId:               jobRequest.SpoutId,
		PrepId:                jobRequest.PrepId,
		Created:               fmt.Sprint(time.Now().UTC().Format(time.RFC3339Nano)),
		Pipe1IngressPort:      pipe1IngressPort,
		Pipe1EgressPort:       pipe1EgressPort,
		Pipe2IngressPort:      pipe2IngressPort,
		Pipe2EgressPort:       pipe2EgressPort,
		EnginePath:            filepath.Join("/", "tmp", "jobs", id, "fluid"),
		BinaryPlanPath:        filepath.Join("/", "tmp", "jobs", id, "plan.bin"),
		LogFilePath:           filepath.Join("fluid.log"),
		SampleCSVFilePath:     filepath.Join("sample.csv"),
		SpoutPath:             filepath.Join("/", "tmp", "jobs", id, "spout.cmd"),
		DemoFinDataServerPath: filepath.Join("/", "tmp", "jobs", id, DemoFinancialDataServerBinary),
		DemoSyslogPath:        filepath.Join("/", "tmp", "jobs", id, DemoSyslogBinary),
		DemoThrottlePath:      filepath.Join("/", "tmp", "jobs", id, DemoThrottleBinary),
		DashboardPort:         dashboardPort,
		DashboardURL:          "http://localhost:" + strconv.Itoa(dashboardPort),
		ExitAfterSeconds:      3600,
		ReaderWebSocket:       WebSocketURLPrefix + ":",
		WriterWebSocket:       WebSocketURLPrefix + ":",
		JobDirectoryPath:      path,
	}
	app.cfg.App.Jobs[job.Id] = job

	var err error
	if err = copyFile(CatalogsDirectoryPath+"/"+job.CatalogId+".json", job.JobDirectoryPath+"/catalog.json", FilePermissionReadable); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	if err = copyFile(QueriesDirectoryPath+"/"+job.QueryId+".fql", job.JobDirectoryPath+"/query.fql", FilePermissionReadable); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	logger.Debug(
		"@@@401",
		"FROM", SpoutsDirectoryPath+"/"+job.SpoutId+".cmd",
		"TO", job.JobDirectoryPath+"/spout.cmd",
	)
	if err = copyFile(SpoutsDirectoryPath+"/"+job.SpoutId+".cmd", job.JobDirectoryPath+"/spout.cmd", FilePermissionReadable); err != nil {
		logger.Debug("@@@401a")
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	logger.Debug("@@@401b")
	if err = copyFile(PrepsDirectoryPath+"/"+job.PrepId+".sh", job.JobDirectoryPath+"/prep.sh", FilePermissionExecutable); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	if err = copyFile(demoDirectoryPath+"/"+WebSocketClientBinary, job.JobDirectoryPath+"/"+WebSocketClientBinary, FilePermissionExecutable); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	if err = copyFile(demoDirectoryPath+"/"+WebSocketServerBinary, job.JobDirectoryPath+"/"+WebSocketServerBinary, FilePermissionExecutable); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	if err = copyFile(demoDirectoryPath+"/"+DashboardBinary, job.JobDirectoryPath+"/"+DashboardBinary, FilePermissionExecutable); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	if err = copyFile(demoDirectoryPath+"/"+DemoFinancialDataServerBinary, job.JobDirectoryPath+"/"+DemoFinancialDataServerBinary, FilePermissionExecutable); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	if err = copyFile(demoDirectoryPath+"/"+DemoSyslogBinary, job.JobDirectoryPath+"/"+DemoSyslogBinary, FilePermissionExecutable); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	if err = copyFile(demoDirectoryPath+"/"+DemoThrottleBinary, job.JobDirectoryPath+"/"+DemoThrottleBinary, FilePermissionExecutable); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	if err = copyFile(demoDirectoryPath+"/"+DemoGeneratorBinary, job.JobDirectoryPath+"/"+DemoGeneratorBinary, FilePermissionExecutable); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	if err = copyFile(templatesDirectoryPath+"/"+DashboardTemplateFile, job.JobDirectoryPath+"/"+DashboardTemplateFile, FilePermissionReadable); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	// if err = copyFile(templatesDirectoryPath+"/"+"run-console-template.sh", job.JobDirectoryPath+"/"+"run-console-template.sh", FilePermissionRead); err != nil {
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// }
	// if err = copyFile(templatesDirectoryPath+"/"+"run-dashboard-template.sh", job.JobDirectoryPath+"/"+"run-dashboard-template.sh", FilePermissionRead); err != nil {
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// }

	logger.Info("@@@300.0")
	var statusMsg string
	if statusMsg, err = prepare(job.JobDirectoryPath); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	if statusMsg, err = buildEngine(job.JobDirectoryPath); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	logger.Info("@@@300.3")

	app.cfg.App.JobChan <- job

	if err = createSampleRunScript(job, templatesDirectoryPath+"/run-console-template.sh", job.JobDirectoryPath+"/run-console.sh", FilePermissionExecutable); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	if err = createSampleRunScript(job, templatesDirectoryPath+"/run-dashboard-template.sh", job.JobDirectoryPath+"/run-dashboard.sh", FilePermissionExecutable); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	// if err = createSampleRunScript(job, job.JobDirectoryPath+"/run-console-template.sh", job.JobDirectoryPath+"/run-console.sh", FilePermissionExecutable); err != nil {
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// }
	// if err = createSampleRunScript(job, job.JobDirectoryPath+"/run-dashboard-template.sh", job.JobDirectoryPath+"/run-dashboard.sh", FilePermissionExecutable); err != nil {
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// }

	res := model.AddResponse{
		Id:      job.Id,
		Message: statusMsg,
	}
	c.JSON(http.StatusOK, res)
}

// Replace some placeholders in the template file with actual values.
func createSampleRunScript(job model.Job, src string, dst string, perm fs.FileMode) (err error) {
	logger.Debug(
		"@@@701 - createSampleRunScript - @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@",
		"job.SpoutPath", job.SpoutPath,
		"src", src,
		"dst", dst,
	)

	logger.Debug(job.SpoutPath)
	dataSpout := readFile(job.SpoutPath)

	s := readFile(src)
	s = strings.Replace(s, "@@@PIPE_1_INGRESS_PORT@@@", strconv.Itoa(job.Pipe1IngressPort), -1)
	s = strings.Replace(s, "@@@PIPE_1_EGRESS_PORT@@@", strconv.Itoa(job.Pipe1EgressPort), -1)
	s = strings.Replace(s, "@@@PIPE_2_INGRESS_PORT@@@", strconv.Itoa(job.Pipe2IngressPort), -1)
	s = strings.Replace(s, "@@@PIPE_2_EGRESS_PORT@@@", strconv.Itoa(job.Pipe2EgressPort), -1)
	s = strings.Replace(s, "@@@DASHBOARD_PORT@@@", strconv.Itoa(job.DashboardPort), -1)
	//s = strings.Replace(s, "@@@CSV_FILE@@@", job.SampleCSVFilePath, -1)
	s = strings.Replace(s, "@@@JOB_LOG@@@", job.LogFilePath, -1)
	//s = strings.Replace(s, "@@@THROTTLE_MILLISECONDS@@@", strconv.Itoa(job.ThrottleMilliseconds), -1)
	s = strings.Replace(s, "@@@EXIT_AFTER_SECONDS@@@", strconv.Itoa(job.ExitAfterSeconds), -1)
	s = strings.Replace(s, "@@@DATA_SPOUT@@@", dataSpout, -1)

	if err = os.WriteFile(dst, []byte(s), perm); err != nil {
		logger.Debug("@@@704a")
		return
	}

	if err = os.Remove(src); err != nil {
		logger.Debug("@@@705a")
		return
	}
	logger.Debug("@@@706")

	return
}

func readFile(path string) string {
	var bytes []byte
	var err error
	if bytes, err = os.ReadFile(path); err != nil {
		logger.Error(
			"cannot read file",
			"err", err,
		)
		panic(err)
	}
	logger.Info("@@@300b")
	return string(bytes)
}

/*
func check(err error) {
	if err != nil {
		logger.Error(
			"error found",
			"err", err,
		)
		panic(err)
	}
}
*/

/*
func copyFile(src string, dst string, perm os.FileMode) (err error) {
	var f *os.File
	f, err = os.Create(dst)
	check(err)
	defer f.Close()
	logger.Debug(
		"@@@6000.0 - created file",
		"name", dst,
	)

	err = os.Chown(dst, os.Getuid(), os.Getgid())
	check(err)
	logger.Debug(
		"@@@6000.1 - changed file owership",
		"name", dst,
	)

	err = os.Chmod(dst, perm)
	check(err)
	logger.Debug(
		"@@@6000.2 - changed file permissions",
		"name", dst,
	)

	var bytesRead []byte
	bytesRead, err = os.ReadFile(src)
	check(err)
	logger.Debug(
		"@@@6000.3 - read bytes from file",
		"name", src,
		"bytes", strconv.Itoa(len(bytesRead)),
	)

	var n int
	n, err = f.Write(bytesRead)
	check(err)
	logger.Debug(
		"@@@6000.4 - wrote bytes to file",
		"name", dst,
		"bytes", strconv.Itoa(n),
	)

	err = f.Sync()
	check(err)
	logger.Debug(
		"@@@6000.5 - file synced",
	)

	// w := bufio.NewWriter(f)
	// n4, err := w.WriteString("buffered\n")
	// check(err)
	// fmt.Printf("wrote %d bytes\n", n4)
	// w.Flush()
	//
	return
}
*/

/*
func copyFile(src string, dst string, perm fs.FileMode) (err error) {

	logger.Debug(
		"@@@500 - copyFile",
		"src", src,
		"dst", dst,
	)

	logger.Debug("@@@500 - BEGIN")

	var bytesRead []byte
	if bytesRead, err = os.ReadFile(src); err != nil {
		logger.Debug("@@@500.1 ##############################################")
		return
	}
	if err = os.WriteFile(dst, bytesRead, perm); err != nil {
		logger.Debug("@@@500.2 ##############################################")
		return
	}
	logger.Debug(
		"@@@500 - END",
		"numBytesRead", strconv.Itoa(len(bytesRead)),
	)
	return
}
*/

/*
func createJobDirectory(id string) (path string) {
	path = filepath.Join("/", "tmp", "jobs", id)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		panic(err)
	}
	return
}
*/

func createJobDirectory(id string) (path string) {
	path = filepath.Join("/", "tmp", "jobs", id)
	if err := os.MkdirAll(path, FilePermissionExecutable); err != nil {
		logger.Error(
			"cannot create directory",
			"err", err,
		)
		panic(err)
	}
	return
}

func prepare(jobDirectoryPath string) (statusMsg string, err error) {
	cmd := exec.Command("./prep.sh")
	cmd.Dir = jobDirectoryPath
	var out bytes.Buffer
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		statusMsg = "problem starting preparation script: " + err.Error()
		return
	}
	statusMsg = out.String()
	return
}

func buildEngine(jobDirectoryPath string) (statusMsg string, err error) {
	params := "JOB_DIR=" + jobDirectoryPath
	cmd := exec.Command("make", "all", params)
	cmd.Dir = "/tmp/repos/fluid"

	logger.Info(
		"@@@1 - buildEngine",
		"params", params,
		"dir", cmd.Dir,
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	logger.Info("@@@2")

	if err = cmd.Run(); err != nil {
		statusMsg = "problem building engine"
		logger.Error(
			statusMsg,
			"err", err,
		)
		return
	}

	logger.Info("@@@4")
	statusMsg = out.String()
	return
}

func startEngine(job model.Job) (statusMsg string, err error) {
	cmd := exec.Command("./run-dashboard.sh")
	cmd.Dir = job.JobDirectoryPath
	var out bytes.Buffer
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		statusMsg = "problem starting engine: " + err.Error()
		return
	}
	statusMsg = out.String()
	return
}

func stopEngine(jobDirectoryPath string) (statusMsg string, err error) {
	var out bytes.Buffer

	cmd := exec.Command("pkill", "demo")
	cmd.Dir = jobDirectoryPath
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		statusMsg = "problem stopping demo"
		return
	}
	statusMsg += out.String()

	cmd = exec.Command("pkill", "fluid")
	cmd.Dir = jobDirectoryPath
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		statusMsg = "problem stopping fluid"
		return
	}
	statusMsg += out.String()

	return
}

/*
func ralfTest() {
	logger.Debug("Hello")
	src := "/tmp/xxx/yyy/src.txt"
	dst := "/tmp/xxx/zzz/dst.txt"

	ralfCreateFileAndDirectories(src, "Hello, beautiful World!")
	copyFile(src, dst, FilePermissionExecutable)
}
*/

/*
func ralfCreateFileAndDirectories(name string, content string) (err error) {
	createPathDirectories(name, FilePermissionExecutable)

	var f *os.File
	f, err = os.Create(name)
	check(err)
	defer f.Close()
	logger.Debug(
		"@@@5000.0 - created file",
		"name", name,
	)

	var n int
	n, err = f.WriteString("Hello, world!\n")
	check(err)
	logger.Debug(
		"@@@5000.1 - text written to file",
		"bytes written", n,
	)
	return
}
*/

func copyFile(src string, dst string, perm os.FileMode) (err error) {
	dir := filepath.Dir(dst)
	if err = os.MkdirAll(dir, perm); err != nil {
		logger.Error(
			"cannot create base directory for file",
			"path", dst,
			"err", err,
		)
		panic(err)
	}

	var f *os.File
	if f, err = os.Create(dst); err != nil {
		logger.Error(
			"cannot create file",
			"path", dst,
			"err", err,
		)
		panic(err)
	}
	defer f.Close()

	if err = os.Chown(dst, os.Getuid(), os.Getgid()); err != nil {
		logger.Error(
			"cannot change file ownership",
			"path", dst,
			"err", err,
		)
		panic(err)
	}

	if err = os.Chmod(dst, perm); err != nil {
		logger.Error(
			"cannot change file permission",
			"path", dst,
			"err", err,
		)
		panic(err)
	}

	var bytesRead []byte
	if bytesRead, err = os.ReadFile(src); err != nil {
		logger.Error(
			"cannot read bytes from file",
			"path", src,
			"err", err,
		)
		panic(err)
	}

	if _, err = f.Write(bytesRead); err != nil {
		logger.Error(
			"cannot write bytes to file",
			"err", err,
		)
		panic(err)
	}

	if err = f.Sync(); err != nil {
		logger.Error(
			"cannot sync file",
			"err", err,
		)
		panic(err)
	}

	if err = f.Close(); err != nil {
		logger.Error(
			"cannot close file",
			"err", err,
		)
		panic(err)
	}

	return
}
