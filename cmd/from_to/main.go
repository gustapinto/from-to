package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/gustapinto/from-to/cmd/from_to/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	configPath := flag.String("config", "", "The configuration file path (ex: -config=./example/config.yaml)")
	flag.Parse()

	cfg, err := config.LoadConfigFromYamlFile(configPath)
	if err != nil {
		logger.Error("Failed to load config file", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("Loaded application config from file", "configPath", *configPath)

	out := config.GetOutputConnector(cfg.Output.Type)
	outSetupParams := config.GetOutputConnectorSetupParams(*cfg)

	if out == nil || outSetupParams == nil {
		logger.Error("Failed to identify output connector from config", "error", err.Error())
		os.Exit(1)
	}

	if err := out.Setup(outSetupParams); err != nil {
		logger.Error("Failed to setup output connector", "error", err.Error())
		os.Exit(1)
	}

	in := config.GetInputConnector(cfg.Input.Type)
	inSetupParams := config.GetInputConnectorSetupParams(*cfg)

	if in == nil || inSetupParams == nil {
		logger.Error("Failed to identify input connector from config", "error", err.Error())
		os.Exit(1)
	}

	if err := in.Setup(inSetupParams); err != nil {
		logger.Error("Failed to setup input connector", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("Application started, listening for new rows to process")

	if err := in.Listen(out); err != nil {
		logger.Error("Failed to listen for new rows", "error", err.Error())
		os.Exit(1)
	}
}
