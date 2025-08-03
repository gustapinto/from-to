package postgres

import "time"

type Config struct {
	TimeoutSeconds uint64   `yaml:"timeoutSeconds"`
	PollSeconds    uint64   `yaml:"pollSeconds"`
	PollLimit      uint64   `yaml:"pollLimit"`
	DSN            string   `yaml:"dsn"`
	Tables         []string `yaml:"tables"`
}

func (c *Config) TimeoutSecondsOrDefault() time.Duration {
	if c.TimeoutSeconds == 0 {
		return 30 * time.Second
	}

	return time.Duration(c.TimeoutSeconds) * time.Second
}

func (c *Config) PollSecondsOrDefault() time.Duration {
	if c.PollSeconds == 0 {
		return 30 * time.Second
	}

	return time.Duration(c.PollSeconds) * time.Second
}

func (c *Config) LimitOrDefault() uint64 {
	if c.PollLimit == 0 {
		return 50
	}

	return c.PollLimit
}
