package postgres

type Config struct {
	TimeoutSeconds uint64   `yaml:"timeoutSeconds"`
	PollSeconds    uint64   `yaml:"pollSeconds"`
	PollLimit      uint64   `yaml:"pollLimit"`
	DSN            string   `yaml:"dsn"`
	Tables         []string `yaml:"tables"`
}
