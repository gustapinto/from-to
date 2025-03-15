package postgres

type Config struct {
	PollSeconds int64    `yaml:"pollSeconds"`
	DSN         string   `yaml:"dsn"`
	Tables      []string `yaml:"tables"`
}
