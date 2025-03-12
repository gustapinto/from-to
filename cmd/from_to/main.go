package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/gustapinto/from-to/cmd/from_to/config"
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

	out := config.GetOutputConnector(cfg.Output.Type)
	outSetupParams := config.GetOutputConnectorSetupParams(*cfg)

	if out == nil || outSetupParams == nil {
		slog.Error("Failed to identify output connector from config", "error", err.Error())
		os.Exit(1)
	}

	if err := out.Setup(outSetupParams); err != nil {
		slog.Error("Failed to setup output connector", "error", err.Error())
		os.Exit(1)
	}

	in := config.GetInputConnector(cfg.Input.Type)
	inSetupParams := config.GetInputConnectorSetupParams(*cfg)

	if in == nil || inSetupParams == nil {
		slog.Error("Failed to identify input connector from config", "error", err.Error())
		os.Exit(1)
	}

	if err := in.Setup(inSetupParams); err != nil {
		slog.Error("Failed to setup input connector", "error", err.Error())
		os.Exit(1)
	}

	slog.Info("Application started, listening for new rows to process")

	if err := in.Listen(out); err != nil {
		slog.Error("Failed to listen for new rows", "error", err.Error())
		os.Exit(1)
	}
}
