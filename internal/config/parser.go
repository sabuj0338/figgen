package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Dependencies struct {
	UI    []string `yaml:"ui"`
	State []string `yaml:"state"`
	Forms []string `yaml:"forms"`
}

type Config struct {
	Framework      string       `yaml:"framework"`
	BoilerplateURL string       `yaml:"boilerplate_url"`
	Dependencies   Dependencies `yaml:"dependencies"`
	Rules          []string     `yaml:"rules"`
}

// LoadConfig reads and parses the configuration file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config: %w", err)
	}

	// Basic validation
	if cfg.Framework == "" {
		cfg.Framework = "nextjs16" // default
	}

	return &cfg, nil
}
