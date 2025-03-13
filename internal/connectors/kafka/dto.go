package kafka

type Topic struct {
	Name              string
	Partitions        int32
	ReplicationFactor int16
}

type SetupParams struct {
	BootstrapServers []string
	Topics           []Topic
}
