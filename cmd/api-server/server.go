package main

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/xralf/fluid/cmd/utils"
	"github.com/xralf/fluid/pkg/server"

	"github.com/spf13/viper"
)

var (
	logger *slog.Logger
)

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	logger.Info("Portal server says welcome!")
}

func main() {
	cfg := readEnvVariables(readConfigFile())
	server.Initialize(cfg)
}

// Environment variables supersede (i.e., overwrite) corresponding config parameters.
func readEnvVariables(cfg utils.Configuration) utils.Configuration {
	var err error
	var v string

	if v = os.Getenv("APP_NAME"); v != "" {
		cfg.App.Name = v
	}
	if v = os.Getenv("API_SERVER_PORT"); v != "" {
		if cfg.Server.Port, err = strconv.Atoi(v); err != nil {
			panic(err)
		}
	}
	if v = os.Getenv("TIMEOUT"); v != "" {
		if cfg.App.Timeout, err = strconv.Atoi(v); err != nil {
			panic(err)
		}
	}
	if v = os.Getenv("WEBSOCKET_URL_PREFIX"); v != "" {
		cfg.App.WebSocketURLPrefix = v
	}
	if v = os.Getenv("WEBSOCKET_CLIENT_BINARY"); v != "" {
		cfg.App.WebSocketClientBinary = v
	}
	if v = os.Getenv("WEBSOCKET_SERVER_BINARY"); v != "" {
		cfg.App.WebSocketServerBinary = v
	}
	if v = os.Getenv("DASHBOARD_BINARY"); v != "" {
		cfg.App.DashboardBinary = v
	}
	if v = os.Getenv("DASHBOARD_TEMPLATE_FILE"); v != "" {
		cfg.App.DashboardTemplateFile = v
	}
	if v = os.Getenv("DASHBOARD_PORT"); v != "" {
		if cfg.App.DashboardPort, err = strconv.Atoi(v); err != nil {
			panic(err)
		}
	}
	if v = os.Getenv("PIPE_1_INGRESS_PORT"); v != "" {
		if cfg.App.Pipe1IngressPort, err = strconv.Atoi(v); err != nil {
			panic(err)
		}
	}
	if v = os.Getenv("PIPE_1_EGRESS_PORT"); v != "" {
		if cfg.App.Pipe1EgressPort, err = strconv.Atoi(v); err != nil {
			panic(err)
		}
	}
	if v = os.Getenv("PIPE_2_INGRESS_PORT"); v != "" {
		if cfg.App.Pipe2IngressPort, err = strconv.Atoi(v); err != nil {
			panic(err)
		}
	}
	if v = os.Getenv("PIPE_2_EGRESS_PORT"); v != "" {
		if cfg.App.Pipe2EgressPort, err = strconv.Atoi(v); err != nil {
			panic(err)
		}
	}

	return cfg
}

func readConfigFile() (cfg utils.Configuration) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv() // Enable VIPER to read Environment Variables
	viper.SetConfigType("yml")

	var err error
	if err = viper.ReadInConfig(); err != nil {
		logger.Error(
			"error reading config file",
			"err", err,
		)
	}
	if err = viper.Unmarshal(&cfg); err != nil {
		logger.Error(
			"unable to decode into struct",
			"err", err,
		)
	}

	logger.Warn(
		"config with variables",
		"cfg", cfg,
	)
	return
}
