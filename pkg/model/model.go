package model

type GetRequest findOneByIdRequest

type DeleteRequest findOneByIdRequest

type StartRequest findOneByIdRequest

type StopRequest findOneByIdRequest

type Catalog struct {
	Id      string `json:"id"`
	Content string `json:"content"`
	Created string `json:"created"`
}

type Query struct {
	Id      string `json:"id"`
	Content string `json:"query"`
	Created string `json:"created"`
}

type Spout struct {
	Id      string `json:"id"`
	Content string `json:"spout"`
	Created string `json:"created"`
}

type Prep struct {
	Id      string `json:"id"`
	Content string `json:"prep"`
	Created string `json:"created"`
}

type Job struct {
	Id                string `json:"id"`
	QueryId           string `json:"queryId"`
	CatalogId         string `json:"catalogId"`
	SpoutId           string `json:"spoutId"`
	PrepId            string `json:"prepId"`
	Created           string `json:"created"`
	Pipe1IngressPort  int    `json:"reader-ws-port-1"`
	Pipe1EgressPort   int    `json:"writer-ws-port-1"`
	Pipe2IngressPort  int    `json:"reader-ws-port-2"`
	Pipe2EgressPort   int    `json:"writer-ws-port-2"`
	DashboardPort     int    `json:"dashboard-port"`
	DashboardURL      string `json:"dashboard-url"`
	EnginePath        string `json:"enginePath"`
	BinaryPlanPath    string `json:"binaryPlanPath"`
	LogFilePath       string `json:"logFilePath"`
	SampleCSVFilePath string `json:"sampleCsvFilePath"`
	//ThrottlePath         string `json:"throttlePath"`
	//ThrottleMilliseconds int    `json:"throttleMilliseconds"`
	DemoFinDataServerPath string `json:"demoFinDataServerPath"`
	DemoSyslogPath        string `json:"demoSyslogPath"`
	DemoThrottlePath      string `json:"demoThrottlePath"`
	SpoutPath             string `json:"spoutPath"`
	ExitAfterSeconds      int    `json:"exitAfterSeconds"`
	ReaderWebSocket       string `json:"reader-ws"`
	WriterWebSocket       string `json:"writer-ws"`
	JobDirectoryPath      string `json:"jobDirectoryPath"`
}

type AddJobRequest struct {
	CatalogId string `json:"catalogId"`
	QueryId   string `json:"queryId"`
	SpoutId   string `json:"spoutId"`
	PrepId    string `json:"prepId"`
}

type AddResponse struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

type StartJobRequest struct {
	Id string `json:"id"`
}

type StartResponse struct {
	DashboardURL string `json:"dashboardURL"`
	Message      string `json:"message"`
}

type StopJobRequest struct {
	Id string `json:"id"`
}

type StopResponse struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

type findOneByIdRequest struct {
	Id string `json:"id"`
}

type ListJobsResponse struct {
	Jobs []Job `json:"jobs"`
}
