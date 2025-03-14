package config

import (
	"errors"
	"fmt"
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
	Input  Input  `yaml:"input"`
	Output Output `yaml:"output"`
}

type Input struct {
	Connector   string  `yaml:"connector"`
	DSN         string  `yaml:"dsn"`
	PollSeconds *int64  `yaml:"pollSeconds"`
	Tables      []Table `yaml:"tables"`
}

type Output struct {
	Connector        string `yaml:"connector"`
	BootstrapServers string `yaml:"bootstrapServers"`
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

type Table struct {
	From   From   `yaml:"from"`
	To     To     `yaml:"to"`
	Mapper Mapper `yaml:"mapper"`
}

type Mapper struct {
	Type     string `yaml:"type"`
	FilePath string `yaml:"filePath"`
	Function string `yaml:"function"`
}

func LoadConfigFromYamlFile(configPath *string) (*Config, error) {
	config, err := getConfigFromFile(*configPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func GetPublisher(config Config) (event.Publisher, error) {
	switch config.Output.Connector {
	case TypeKafka:
		topics := make([]kafka.Topic, len(config.Input.Tables))
		for i, table := range config.Input.Tables {
			partitions := int32(3)
			if table.To.Partitions != 0 {
				partitions = table.To.Partitions
			}

			replicationFactor := int16(1)
			if table.To.ReplicationFactor != 0 {
				replicationFactor = table.To.ReplicationFactor
			}

			topics[i] = kafka.Topic{
				Name:              table.To.Topic,
				Partitions:        partitions,
				ReplicationFactor: replicationFactor,
			}
		}

		return kafka.NewPublisher(kafka.SetupParams{
			BootstrapServers: strings.Split(config.Output.BootstrapServers, ","),
			Topics:           topics,
		})
	}

	return nil, errors.New("invalid publisher type")
}

func GetListener(config Config) (event.Listener, error) {
	switch config.Input.Connector {
	case TypePostgres:
		tables := make([]postgres.SetupParamsTable, len(config.Input.Tables))
		for i, table := range config.Input.Tables {
			mapper, err := getMapper(table.Mapper)
			if err != nil {
				return nil, err
			}

			tables[i] = postgres.SetupParamsTable{
				Name:      table.From.Name,
				KeyColumn: table.From.KeyColumn,
				EventMetadata: event.Config{
					Mapper: mapper,
					Lua: event.LuaConfig{
						FilePath: table.Mapper.FilePath,
						Function: table.Mapper.Function,
					},
					Kafka: event.KafkaConfig{
						Topic: table.To.Topic,
					},
				},
			}
		}

		pollSeconds := int64(30)
		if config.Input.PollSeconds != nil {
			pollSeconds = *config.Input.PollSeconds
		}

		return postgres.NewListener(postgres.SetupParams{
			DSN:         config.Input.DSN,
			PollSeconds: pollSeconds,
			Tables:      tables,
		})
	}

	return nil, errors.New("invalid listener type")
}

func getMapper(mapper Mapper) (event.Mapper, error) {
	switch mapper.Type {
	case TypeLua:
		return lua.NewMapper()
	}

	return nil, nil
}

func isEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

func newEmptyFieldErr(field string) error {
	return fmt.Errorf("missing or empty field [%s]", field)
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

func validateMapperConfigForLua(mapperPath string, mapper Mapper) error {
	if mapper.Type != TypeLua {
		return nil
	}

	if isEmpty(mapper.FilePath) {
		return newEmptyFieldErr(fmt.Sprintf("%s.mapper.filePath", mapperPath))
	}

	if isEmpty(mapper.Function) {
		return newEmptyFieldErr(fmt.Sprintf("%s.mapper.function", mapperPath))
	}

	return nil
}

func validateInputConfigForPostgres(input Input, outType string) error {
	if input.Connector != TypePostgres {
		return nil
	}

	if isEmpty(input.DSN) {
		return newEmptyFieldErr("input.dsn")
	}

	if len(input.Tables) == 0 {
		return newEmptyFieldErr("input.tables")
	}

	for i, table := range input.Tables {
		if isEmpty(table.From.Name) {
			return newEmptyFieldErr(fmt.Sprintf("input.tables[%d].from.name", i))
		}

		if isEmpty(table.From.KeyColumn) {
			return newEmptyFieldErr(fmt.Sprintf("input.tables[%d].from.keyColumn", i))
		}

		if outType == TypeKafka {
			if isEmpty(table.To.Topic) {
				return newEmptyFieldErr(fmt.Sprintf("input.tables[%d].to.topic", i))
			}
		}

		if !isEmpty(table.Mapper.Type) {
			if err := validateMapperConfigForLua(fmt.Sprintf("input.tables[%d]", i), table.Mapper); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateOutputConfigForKafka(out Output) error {
	if out.Connector != TypeKafka {
		return nil
	}

	if isEmpty(out.BootstrapServers) {
		return newEmptyFieldErr("output.bootstrapServers")
	}

	return nil
}

func validateConfig(config *Config) error {
	if config == nil {
		return errors.New("failed to load config")
	}

	if isEmpty(config.Input.Connector) {
		return newEmptyFieldErr("input.connector")
	}

	if isEmpty(config.Output.Connector) {
		return newEmptyFieldErr("output.connector")
	}

	if err := validateInputConfigForPostgres(config.Input, config.Output.Connector); err != nil {
		return err
	}

	if err := validateOutputConfigForKafka(config.Output); err != nil {
		return err
	}

	return nil
}
