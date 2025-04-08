package event

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

	Key string `yaml:"-"`
}
