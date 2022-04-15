package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Service struct {
	Name    string   `yaml:"name"`
	Image   string   `yaml:"image"`
	Env     []string `yaml:"env"`
	Ports   []string `yaml:"ports"`
	Volumes []string `yaml:"volumes"`
}

type Config struct {
	Services []Service `yaml:"services"`
}

func Parse(path string) (*Config, error) {
	yfile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %s", err)
	}

	config := Config{}

	err = yaml.Unmarshal(yfile, &config)
	if err != nil {
		return nil, fmt.Errorf("invalid config file: %s", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	return nil
}
