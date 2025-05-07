package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	// Флаги
	serverAddress := flag.String("a", "localhost:8080", "server address")
	baseURL := flag.String("b", "http://localhost:8080", "base URL")
	fileStoragePath := flag.String("f", "urls.json", "file storage path")
	databaseDSN := flag.String("d", "", "database DSN")

	flag.Parse()

	// Переменные окружения
	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		*serverAddress = envServerAddress
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		*baseURL = envBaseURL
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		*fileStoragePath = envFileStoragePath
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		*databaseDSN = envDatabaseDSN
	}

	cfg.ServerAddress = *serverAddress
	cfg.BaseURL = *baseURL
	cfg.FileStoragePath = *fileStoragePath
	cfg.DatabaseDSN = *databaseDSN

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
