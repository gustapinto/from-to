package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/gustapinto/from-to/internal"
	"github.com/gustapinto/from-to/internal/input"
	"github.com/gustapinto/from-to/internal/output"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Input  Input  `yaml:"input"`
	Output Output `yaml:"output"`
}

type Input struct {
	Type        string  `yaml:"type"`
	DSN         string  `yaml:"dsn"`
	PollSeconds int64   `yaml:"pollSeconds"`
	Tables      []Table `yaml:"tables"`
}

type Output struct {
	Type             string `yaml:"type"`
	BootstrapServers string `yaml:"bootstrapServers"`
}

type Table struct {
	From From `yaml:"from"`
	To   To   `yaml:"to"`
}

type From struct {
	Name      string `yaml:"name"`
	KeyColumn string `yaml:"keyColumn"`
}

type To struct {
	Topic             string `yaml:"topic"`
	Partitions        int32  `yaml:"partitions"`
	ReplicationFactor int16  `yaml:"replicationFactor"`
}

func LoadConfigFromYamlFile(configPath *string) (*Config, error) {
	if configPath == nil || *configPath == "" {
		return nil, errors.New("missing or empty -config=* param")
	}

	ext := strings.ToLower(filepath.Ext(*configPath))
	if ext != ".yml" && ext != ".yaml" {
		return nil, errors.New("config must have a .yml or .yaml extension")
	}

	configAbsPath, err := filepath.Abs(*configPath)
	if err != nil {
		return nil, err
	}

	configBytes, err := os.ReadFile(configAbsPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func GetOutputConnector(outputType string) internal.OutputConnector {
	switch outputType {
	case "kafka":
		return &output.KafkaOutputConnector{}
	default:
		return nil
	}
}

func GetOutputConnectorSetupParams(config Config) any {
	switch config.Output.Type {
	case "kafka":
		topics := make([]output.KafkaTopic, len(config.Input.Tables))
		for i, table := range config.Input.Tables {
			topics[i] = output.KafkaTopic{
				Name:              table.To.Topic,
				Partitions:        table.To.Partitions,
				ReplicationFactor: table.To.ReplicationFactor,
			}
		}

		return &output.KafkaSetupParams{
			BootstrapServers: strings.Split(config.Output.BootstrapServers, ","),
			Topics:           topics,
		}
	}

	return nil
}

func GetInputConnector(inputType string) internal.InputConnector {
	switch inputType {
	case "postgres":
		return &input.PostgresInputConnector{}
	}

	return nil
}

func GetInputConnectorSetupParams(config Config) any {
	switch config.Input.Type {
	case "postgres":
		tables := make([]input.PostgresSetupParamsTable, len(config.Input.Tables))
		for i, table := range config.Input.Tables {
			tables[i] = input.PostgresSetupParamsTable{
				Name:      table.From.Name,
				KeyColumn: table.From.KeyColumn,
				Topic:     table.To.Topic,
			}

			return &input.PostgresSetupParams{
				DSN:         config.Input.DSN,
				PollSeconds: config.Input.PollSeconds,
				Tables:      tables,
			}
		}

		return nil
	}

	return nil
}
