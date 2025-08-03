package webhook

type Config struct {
	URL            string            `yaml:"url"`
	Headers        map[string]string `yaml:"headers"`
	TimeoutSeconds uint64            `yaml:"requestTimeout"`
	Retries        uint64            `yaml:"retries"`
}
