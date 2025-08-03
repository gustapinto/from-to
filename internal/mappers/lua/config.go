package lua

type Config struct {
	FilePath string  `yaml:"filePath"`
	Function string  `yaml:"function"`
	Source   *string `yaml:"source"`
}
