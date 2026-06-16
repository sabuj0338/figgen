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
	Framework           string       `yaml:"framework"`
	BoilerplateURL      string       `yaml:"boilerplate_url"`
	PackageManager      string       `yaml:"package_manager"`
	PlannerRules        []string     `yaml:"planner_rules"`
	CoderRules          []string     `yaml:"coder_rules"`
	PlannerRulesContent string       `yaml:"-"` // Injected dynamically
	CoderRulesContent   string       `yaml:"-"` // Injected dynamically
	Dependencies        Dependencies `yaml:"dependencies"`
	Rules               []string     `yaml:"rules"`
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

	// Load Planner Rules
	var combinedPlannerRules string
	for _, pPath := range cfg.PlannerRules {
		pData, err := os.ReadFile(pPath)
		if err == nil {
			combinedPlannerRules += string(pData) + "\n\n---\n\n"
		} else {
			fmt.Printf("Warning: Could not load planner rule from %s: %v\n", pPath, err)
		}
	}
	cfg.PlannerRulesContent = combinedPlannerRules

	// Load Coder Rules
	var combinedCoderRules string
	for _, cPath := range cfg.CoderRules {
		cData, err := os.ReadFile(cPath)
		if err == nil {
			combinedCoderRules += string(cData) + "\n\n---\n\n"
		} else {
			fmt.Printf("Warning: Could not load coder rule from %s: %v\n", cPath, err)
		}
	}
	cfg.CoderRulesContent = combinedCoderRules

	// Basic validation
	if cfg.Framework == "" {
		cfg.Framework = "nextjs16" // default
	}
	if cfg.PackageManager == "" {
		cfg.PackageManager = "pnpm" // default to user's preference
	}

	return &cfg, nil
}
