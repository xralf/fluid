package utils

import "github.com/xralf/fluid/pkg/model"

type Configuration struct {
	Database DatabaseSettings
	Server   ServerSettings
	App      ApplicationSettings
}

type DatabaseSettings struct {
	Url        string
	DbName     string
	Collection string
}

type ServerSettings struct {
	Port int
}

type ApplicationSettings struct {
	Name                  string
	Timeout               int
	WebSocketURLPrefix    string
	WebSocketClientBinary string
	WebSocketServerBinary string
	DashboardTemplateFile string
	DashboardBinary       string
	DashboardPort         int
	Pipe1IngressPort      int
	Pipe1EgressPort       int
	Pipe2IngressPort      int
	Pipe2EgressPort       int

	Jobs    map[string]model.Job
	JobChan chan model.Job

	AvailablePorts []int
}
