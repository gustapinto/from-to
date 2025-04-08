package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/gustapinto/from-to/cmd/from_to/config"
	"github.com/gustapinto/from-to/internal/event"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
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
		return fmt.Errorf("Failed to load config file, got error %s", err.Error())
	}

	slog.Info("Loaded application config from file", "configPath", *configPath)

	listener, err := config.GetListener(*cfg)
	if err != nil {
		return fmt.Errorf("Failed to setup output connector from config, got error %s", err.Error())
	}

	publishers, err := config.GetPublishers(*cfg)
	if err != nil {
		return fmt.Errorf("Failed to setup outputs from config, got error %s", err.Error())
	}

	mappers, err := config.GetMappers(*cfg)
	if err != nil {
		return fmt.Errorf("Failed to setup mappers from config, got error %s", err.Error())
	}

	channels, err := config.GetChannels(*cfg)
	if err != nil {
		return fmt.Errorf("Failed to setup channels from config, got error %s", err.Error())
	}

	processor := event.NewProcessor(listener, publishers, mappers, channels)

	slog.Info("Application started, listening for new rows to process")

	return processor.ListenAndProcess()
}
