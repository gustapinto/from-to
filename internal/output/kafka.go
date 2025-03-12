package output

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gustapinto/from-to/internal"
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
}

func (c *KafkaOutputConnector) Setup(config any) error {
	params, ok := config.(*KafkaSetupParams)
	if !ok {
		return errors.New("Invalid config type passed to KafkaOutputConnector.Setup(...), expected *KafkaSetupParams")
	}

	kafkaClient, err := kgo.NewClient(kgo.SeedBrokers(params.BootstrapServers...))
	if err != nil {
		return err
	}

	if err := kafkaClient.Ping(context.Background()); err != nil {
		return err
	}

	kafkaAdminClient := kadm.NewClient(kafkaClient)

	for _, topic := range params.Topics {
		_, err := kafkaAdminClient.CreateTopic(
			context.Background(),
			topic.Partitions,
			topic.ReplicationFactor,
			map[string]*string{},
			topic.Name)
		if err != nil {
			if strings.Contains(err.Error(), "TOPIC_ALREADY_EXISTS") {
				continue
			}

			return err
		}
	}

	return nil
}

func (c *KafkaOutputConnector) Publish(row internal.Row) error {
	value, err := json.Marshal(row.Data)
	if err != nil {
		return err
	}

	record := kgo.Record{
		Key:   []byte(row.Key),
		Value: value,
		Topic: row.Topic,
	}

	return c.client.ProduceSync(context.Background(), &record).FirstErr()
}
