package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/gustapinto/from-to/cmd/from_to/config"
	"github.com/gustapinto/from-to/internal/event"
)

func main() {

	configPath := flag.String("config", "", "The configuration file path (ex: -config=./example/config.yaml)")
	isDebug := flag.Bool("debug", false, "Use to enable debug level logging")
	flag.Parse()

	if *isDebug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	cfg, err := config.LoadConfigFromYamlFile(configPath)
	if err != nil {
		slog.Error("Failed to load config file", "error", err.Error())
		os.Exit(1)
	}

	slog.Info("Loaded application config from file", "configPath", *configPath)

	listener, err := config.GetListener(*cfg)
	if err != nil {
		slog.Error("Failed to setup output connector from config", "error", err.Error())
		os.Exit(1)
	}

	publishers, err := config.GetPublishers(*cfg)
	if err != nil {
		slog.Error("Failed to setup outputs from config", "error", err.Error())
		os.Exit(1)
	}

	mappers, err := config.GetMappers(*cfg)
	if err != nil {
		slog.Error("Failed to setup mappers from config", "error", err.Error())
		os.Exit(1)
	}

	channels, err := config.GetChannels(*cfg)
	if err != nil {
		slog.Error("Failed to setup channels from config", "error", err.Error())
		os.Exit(1)
	}

	processor := event.NewProcessor(listener, publishers, mappers, channels)

	slog.Info("Application started, listening for new rows to process")

	if err := processor.ListenAndProcess(); err != nil {
		slog.Error("Failed to listen and process to events", "error", err.Error())
		os.Exit(1)
	}
}
