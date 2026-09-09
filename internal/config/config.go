package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Database   DatabaseConfig   `yaml:"database"`
	Generation GenerationConfig `yaml:"generation"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	URL    string `yaml:"url"`
}

type GenerationConfig struct {
	Output  string                   `yaml:"output"`
	Queries map[string][]QueryConfig `yaml:"queries"`
}

type QueryConfig struct {
	Name  string   `yaml:"name"`
	Where []string `yaml:"where"`
}

func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Load .env if present — godotenv.Load is a no-op if file is missing
	godotenv.Load()

	cfg.Database.URL = os.ExpandEnv(cfg.Database.URL)

	return &cfg, nil
}
