package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/gustapinto/from-to/internal/config"
	"github.com/gustapinto/from-to/internal/event"
	"github.com/gustapinto/from-to/internal/logging"
)

func main() {
	configPath := flag.String("manifest", "from_to.yaml", "The configuration manifest file path")
	logFormat := flag.String("logFormat", "text", "The logging format, one of [text, json]")
	noColor := flag.Bool("noColor", false, "Use to disable colored logging, only valid for text logFormat")
	isDebug := flag.Bool("debug", false, "Use to enable debug level logging")
	flag.Parse()

	if err := run(*configPath, *logFormat, *noColor, *isDebug); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(configPath, logFormat string, noColor, isDebug bool) error {
	if err := logging.SetupSlog(isDebug, noColor, logFormat); err != nil {
		return err
	}

	cfg, err := config.LoadConfigFromYamlFile(configPath)
	if err != nil {
		return fmt.Errorf("Failed to load config file, got error %s", err.Error())
	}

	slog.Info("Loaded application config from file", "configPath", configPath)

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

	if err := processor.ListenAndProcess(); err != nil {
		return fmt.Errorf("Failed to listen and process, got error %s", err.Error())
	}

	return nil
}
