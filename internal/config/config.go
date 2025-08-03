package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/gustapinto/from-to/internal/connectors/kafka"
	"github.com/gustapinto/from-to/internal/connectors/postgres"
	"github.com/gustapinto/from-to/internal/connectors/webhook"
	"github.com/gustapinto/from-to/internal/event"
	"github.com/gustapinto/from-to/internal/mappers/lua"
	"gopkg.in/yaml.v2"
)

const (
	_typePostgres = "postgres"
	_typeKafka    = "kafka"
	_typeLua      = "lua"
	_typeWebhook  = "webhook"
)

type Manifest struct {
	Version uint64 `yaml:"version"`
	Config  Config `yaml:"config"`
}

type Config struct {
	Input    Input                    `yaml:"input"`
	Outputs  map[string]Output        `yaml:"outputs"`
	Mappers  map[string]Mapper        `yaml:"mappers"`
	Channels map[string]event.Channel `yaml:"channels"`
}

type Input struct {
	Connector      string          `yaml:"connector"`
	PostgresConfig postgres.Config `yaml:"postgresConfig"`
}

type Output struct {
	Connector     string         `yaml:"connector"`
	KafkaConfig   kafka.Config   `yaml:"kafkaConfig"`
	WebHookConfig webhook.Config `yaml:"webhookConfig"`
}

type Mapper struct {
	Type      string     `yaml:"type"`
	LuaConfig lua.Config `yaml:"luaConfig"`
}

func LoadConfigFromYamlFile(configPath string) (*Config, error) {
	config, err := getConfigFromFile(configPath)
	if err != nil {
		return nil, err
	}

	if config.Channels != nil {
		for key, channel := range config.Channels {
			channel.Key = key
			config.Channels[key] = channel
		}
	}

	return config, nil
}

func GetListener(config Config) (event.Listener, error) {
	switch config.Input.Connector {
	case _typePostgres:
		return postgres.NewListener(config.Input.PostgresConfig, config.Channels)
	}

	return nil, errors.New("invalid config type, expected one of: [postgres]")
}

func GetMappers(config Config) (mappers map[string]event.Mapper, err error) {
	mappers = make(map[string]event.Mapper, len(config.Mappers))

	for key, m := range config.Mappers {
		switch m.Type {
		case _typeLua:
			mappers[key], err = lua.NewMapper(m.LuaConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	return mappers, nil
}

func GetPublishers(config Config) (publishers map[string]event.Publisher, err error) {
	publishers = make(map[string]event.Publisher, len(config.Outputs))

	for key, o := range config.Outputs {
		switch o.Connector {
		case _typeKafka:
			publishers[key], err = kafka.NewPublisher(o.KafkaConfig)
			if err != nil {
				return nil, err
			}

		case _typeWebhook:
			publishers[key] = webhook.NewPublisher(o.WebHookConfig)
		}
	}

	return publishers, nil
}

func GetChannels(config Config) (channels map[string]event.Channel, err error) {
	return config.Channels, nil
}

func isEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

func readManifestFromFile(path string) (*Manifest, error) {
	if isEmpty(path) {
		return nil, errors.New("missing or empty -config=* param")
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".yml" && ext != ".yaml" {
		return nil, errors.New("config must have a .yml or .yaml extension")
	}

	configAbsPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	configBytes, err := os.ReadFile(configAbsPath)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := yaml.Unmarshal(configBytes, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func getConfigFromFile(path string) (*Config, error) {
	manifest, err := readManifestFromFile(path)
	if err != nil {
		return nil, err
	}

	version := manifest.Version
	if version == 0 {
		version = 1
	}

	if version == 1 {
		return &manifest.Config, nil
	}

	return nil, errors.New("invalid config")
}
