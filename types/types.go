package types

type Config struct {
	ELKHost     string `yaml:"elk_host"`
	ProjectName string `yaml:"project_name,omitempty"`
	Username    string `yaml:"username,omitempty"`
	Password    string `yaml:"password,omitempty"`
}
