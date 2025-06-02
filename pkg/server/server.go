package server

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xralf/fluid/cmd/utils"
	"github.com/xralf/fluid/pkg/handler"
	"github.com/xralf/fluid/pkg/model"
)

var (
	logger *slog.Logger
)

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	logger.Info("API server says welcome!")
}

func Initialize(cfg utils.Configuration) {
	logger.Info(
		"application starting",
		"name", cfg.App.Name,
	)

	cfg.App.JobChan = make(chan model.Job, 3)
	cfg.App.Jobs = map[string]model.Job{}
	for p := 50002; p <= 50099; p++ {
		cfg.App.AvailablePorts = append(cfg.App.AvailablePorts, p)
	}

	go scheduler(cfg)

	catalogHandler := handler.NewCatalogHandler(cfg)
	queryHandler := handler.NewQueryHandler(cfg)
	spoutHandler := handler.NewSpoutHandler(cfg)
	prepHandler := handler.NewPrepHandler(cfg)
	jobHandler := handler.NewJobHandler(cfg)

	router := gin.Default()

	api := router.Group("api/v1")
	{
		api.POST("/catalog/add", catalogHandler.UploadFile)
		api.POST("/catalog/delete", catalogHandler.Delete)

		api.POST("/query/add", queryHandler.UploadFile)
		api.POST("/query/delete", queryHandler.Delete)

		api.POST("/spout/add", spoutHandler.UploadFile)
		api.POST("/spout/delete", spoutHandler.Delete)

		api.POST("/prep/add", prepHandler.UploadFile)
		api.POST("/prep/delete", prepHandler.Delete)

		api.POST("/job/add", jobHandler.Add)
		api.POST("/job/delete", jobHandler.Delete)
		api.POST("/job/start", jobHandler.Start)
		api.POST("/job/stop", jobHandler.Stop)
		api.GET("/job/list", jobHandler.List)
		//api.GET("/job/list/:limit", jobHandler.List)
	}
	formattedUrl := fmt.Sprintf(":%d", cfg.Server.Port)
	router.Run(formattedUrl)
}

func scheduler(cfg utils.Configuration) {
	for {
		job := <-cfg.App.JobChan
		logger.Info(
			"scheduler",
			"job id", job.Id,
		)
		go worker(job)
	}
}

func worker(job model.Job) {
	// Run the Fluid engine in the background
	//server.Run(job)

	//
	// OLD stuff:
	//
	//server.Run(job.ReaderWebSocketPort, job.EnginePath)
	/*
		    for {
				log.Infoanyf("worker: job id: %s", job.Id))
				time.Sleep(time.Second)
			}
	*/
}
