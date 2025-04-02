package config

import (
	"flag"
	"fmt"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.Parse()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func (cfg *Config) validate() error {
	if cfg.ServerAddress == "" {
		return fmt.Errorf("server address is required")
	}
	if cfg.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	return nil
}
