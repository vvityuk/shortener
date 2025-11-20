package config

import (
	"flag"
	"fmt"
	"os"
)

// Config содержит конфигурацию приложения.
// Значения могут быть установлены через флаги командной строки или переменные окружения.
type Config struct {
	// ServerAddress адрес и порт для запуска HTTP-сервера (например, "localhost:8080")
	ServerAddress string

	// BaseURL базовый URL для генерации коротких ссылок (например, "http://localhost:8080")
	BaseURL string

	// FileStoragePath путь к файлу для хранения данных (например, "urls.json")
	// Если пусто, используется хранилище в памяти
	FileStoragePath string

	// DatabaseDSN строка подключения к PostgreSQL (например, "postgres://user:pass@localhost/dbname")
	// Если указана, имеет приоритет над файловым хранилищем
	DatabaseDSN string
}

// NewConfig создает новую конфигурацию приложения.
// Читает параметры из флагов командной строки и переменных окружения.
// Переменные окружения имеют приоритет над флагами.
//
// Флаги командной строки:
//
//	-a: адрес сервера (по умолчанию "localhost:8080")
//	-b: базовый URL (по умолчанию "http://localhost:8080")
//	-f: путь к файлу хранилища (по умолчанию "urls.json")
//	-d: строка подключения к БД (по умолчанию пусто)
//
// Переменные окружения:
//
//	SERVER_ADDRESS: адрес сервера
//	BASE_URL: базовый URL
//	FILE_STORAGE_PATH: путь к файлу хранилища
//	DATABASE_DSN: строка подключения к БД
//
// Возвращает конфигурацию приложения или ошибку валидации конфигурации.
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
