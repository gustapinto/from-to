package kafka

type TopicConfig struct {
	Name              string `yaml:"name"`
	Partitions        int32  `yaml:"partitions"`
	ReplicationFactor int16  `yaml:"replicationFactor"`
}

func (tc *TopicConfig) PartitionsOrDefault() int32 {
	if tc.Partitions <= 0 {
		return 3
	}

	return tc.Partitions
}

func (tc *TopicConfig) ReplicationFactorOrDefault() int16 {
	if tc.ReplicationFactor <= 0 {
		return 1
	}

	return tc.ReplicationFactor
}

type Config struct {
	BootstrapServers []string    `yaml:"bootstrapServers"`
	Topic            TopicConfig `yaml:"topic"`
}
