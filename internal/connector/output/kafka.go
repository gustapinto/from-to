package output

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/gustapinto/from-to/internal/connector"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaTopic struct {
	Name              string
	Partitions        int32
	ReplicationFactor int16
}

type KafkaSetupParams struct {
	BootstrapServers []string
	Topics           []KafkaTopic
}

type KafkaOutputConnector struct {
	client *kgo.Client
	adm    *kadm.Client
	logger *slog.Logger
}

func (c *KafkaOutputConnector) Setup(config any) error {
	params, ok := config.(*KafkaSetupParams)
	if !ok {
		return errors.New("Invalid config type passed to KafkaOutputConnector.Setup(...), expected *KafkaSetupParams")
	}

	client, err := c.setupClient(params.BootstrapServers)
	if err != nil {
		return err
	}

	c.client = client
	c.adm = kadm.NewClient(client)
	c.logger = slog.With("connector", "KafkaOutputConnector")

	for _, topic := range params.Topics {
		if err := c.setupTopic(topic); err != nil {
			return err
		}

		c.logger.Debug("Topic setup completed", "topic", topic.Name)
	}

	c.logger.Info("Connector setup completed")

	return nil
}

func (c *KafkaOutputConnector) Publish(event connector.Event) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	record := kgo.Record{
		Key:   []byte(event.Metadata.KeyValue),
		Value: value,
		Topic: event.Metadata.Topic,
	}

	if err := c.client.ProduceSync(context.Background(), &record).FirstErr(); err != nil {
		return err
	}

	c.logger.Debug("Row published", "key", string(record.Key), "topic", event.Metadata.Topic, "value", value)

	return nil
}

func (c *KafkaOutputConnector) setupClient(bootstrapServers []string) (*kgo.Client, error) {
	client, err := kgo.NewClient(kgo.SeedBrokers(bootstrapServers...))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *KafkaOutputConnector) setupTopic(topic KafkaTopic) error {
	_, err := c.adm.CreateTopic(
		context.Background(),
		topic.Partitions,
		topic.ReplicationFactor,
		map[string]*string{},
		topic.Name)
	if err != nil {
		if strings.Contains(err.Error(), "TOPIC_ALREADY_EXISTS") {
			return nil
		}

		return err
	}

	return nil
}
