package kafka

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gustapinto/from-to/internal/event"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Publisher struct {
	client *kgo.Client
	adm    *kadm.Client
	logger *slog.Logger
}

func NewPublisher(config Config) (c *Publisher, err error) {
	c = &Publisher{
		logger: slog.With("publisher", "Kafka"),
	}

	client, err := c.setupClient(config.BootstrapServers)
	if err != nil {
		return nil, err
	}

	c.client = client
	c.adm = kadm.NewClient(client)

	for _, topic := range config.Topics {
		if err := c.setupTopic(topic); err != nil {
			return nil, err
		}

		c.logger.Debug("Topic setup completed", "topic", topic.Name)
	}

	c.logger.Info("Connector setup completed")

	return c, nil
}

func (c *Publisher) Publish(e event.Event, payload []byte, topic string) error {
	record := kgo.Record{
		Key:   []byte(strconv.Itoa(int(e.ID))),
		Value: payload,
		Topic: topic,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := c.client.ProduceSync(ctx, &record).FirstErr(); err != nil {
		return err
	}

	c.logger.Debug(
		"Row published",
		"key", string(record.Key),
		"topic", topic,
		"payload", string(payload),
	)

	return nil
}

func (c *Publisher) setupClient(bootstrapServers []string) (*kgo.Client, error) {
	client, err := kgo.NewClient(kgo.SeedBrokers(bootstrapServers...))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Publisher) setupTopic(topic TopicConfig) error {
	partitions := topic.Partitions
	if partitions == 0 {
		partitions = 3
	}

	replicationFactor := topic.ReplicationFactor
	if replicationFactor == 0 {
		replicationFactor = 1
	}

	_, err := c.adm.CreateTopic(
		context.Background(),
		partitions,
		replicationFactor,
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
