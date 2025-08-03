package webhook

import "time"

type Config struct {
	URL            string            `yaml:"url"`
	Headers        map[string]string `yaml:"headers"`
	TimeoutSeconds uint64            `yaml:"requestTimeout"`
	Retries        *uint64           `yaml:"retries"`
}

func (c *Config) HeadersOrDefault() map[string]string {
	if c.Headers == nil {
		return map[string]string{}
	}

	return c.Headers
}

func (c *Config) TimeoutSecondsOrDefault() time.Duration {
	if c.TimeoutSeconds == 0 {
		return 30 * time.Second
	}

	return time.Duration(c.TimeoutSeconds) * time.Second
}

func (c *Config) RetriesOrDefault() uint64 {
	if c.Retries == nil {
		return 3
	}

	return *c.Retries
}
