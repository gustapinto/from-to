package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/gustapinto/from-to/internal/connectors/kafka"
	"github.com/gustapinto/from-to/internal/connectors/postgres"
	"github.com/gustapinto/from-to/internal/event"
	"github.com/gustapinto/from-to/internal/mappers/lua"
	"gopkg.in/yaml.v2"
)

var (
	TypePostgres = "postgres"
	TypeKafka    = "kafka"
	TypeLua      = "lua"
)

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
	Connector   string       `yaml:"connector"`
	KafkaConfig kafka.Config `yaml:"kafkaConfig"`
}

type Mapper struct {
	Type      string     `yaml:"type"`
	LuaConfig lua.Config `yaml:"luaConfig"`
}

func LoadConfigFromYamlFile(configPath *string) (*Config, error) {
	config, err := getConfigFromFile(*configPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func GetListener(config Config) (event.Listener, error) {
	switch config.Input.Connector {
	case TypePostgres:
		return postgres.NewListener(config.Input.PostgresConfig, config.Channels)
	}

	return nil, errors.New("invalid config type, expected one of: [postgres]")
}

func GetMappers(config Config) (mappers map[string]event.Mapper, err error) {
	mappers = make(map[string]event.Mapper)

	for key, m := range config.Mappers {
		switch m.Type {
		case TypeLua:
			mappers[key], err = lua.NewMapper(m.LuaConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	return mappers, nil
}

func GetPublishers(config Config) (publishers map[string]event.Publisher, err error) {
	publishers = make(map[string]event.Publisher)

	for key, o := range config.Outputs {
		switch o.Connector {
		case TypeKafka:
			publishers[key], err = kafka.NewPublisher(o.KafkaConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	return publishers, nil
}

func GetChannels(config Config) (channels map[string]event.Channel, err error) {
	channels = make(map[string]event.Channel)

	for key, c := range channels {
		channels[key] = c
	}

	return channels, nil
}

func isEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

func getConfigFromFile(configPath string) (*Config, error) {
	if isEmpty(configPath) {
		return nil, errors.New("missing or empty -config=* param")
	}

	ext := strings.ToLower(filepath.Ext(configPath))
	if ext != ".yml" && ext != ".yaml" {
		return nil, errors.New("config must have a .yml or .yaml extension")
	}

	configAbsPath, err := filepath.Abs(configPath)
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
