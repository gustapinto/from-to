package event

// type KafkaConfig struct {
// 	Topic string
// }

// type LuaConfig struct {
// 	FilePath string
// 	Function string
// }

// type Config struct {
// 	Key      string
// 	KeyValue string
// 	Mapper   Mapper

// 	// Connector/Mapper specific stuff
// 	Lua   LuaConfig
// 	Kafka KafkaConfig
// }

type Event struct {
	ID    int64          `json:"id,omitempty"`
	Ts    int64          `json:"ts,omitempty"`
	Op    string         `json:"op,omitempty"`
	Table string         `json:"table,omitempty"`
	Row   map[string]any `json:"row,omitempty"`

	Channels []Channel `json:"-"`
}

type Channel struct {
	From   string `yaml:"from"`
	To     string `yaml:"to"`
	Output string `yaml:"output"`
	Mapper string `yaml:"mapper"`
}
