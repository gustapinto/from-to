package kafka

type TopicConfig struct {
	Name              string `yaml:"name"`
	Partitions        int32  `yaml:"partitions"`
	ReplicationFactor int16  `yaml:"replicationFactor"`
}

type Config struct {
	BootstrapServers []string      `yaml:"bootstrapServers"`
	Topics           []TopicConfig `yaml:"topics"`
}
