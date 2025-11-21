// Package config предоставляет функциональность для загрузки и валидации конфигурации приложения.
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

	// EnableHTTPS включает HTTPS режим работы сервера
	EnableHTTPS bool

	// TLSCertFile путь к файлу сертификата TLS (например, "server.crt")
	TLSCertFile string

	// TLSKeyFile путь к файлу приватного ключа TLS (например, "server.key")
	TLSKeyFile string
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
//	-s: включить HTTPS режим
//	-cert: путь к файлу сертификата TLS (по умолчанию "server.crt", если не указан - генерируется автоматически)
//	-key: путь к файлу приватного ключа TLS (по умолчанию "server.key", если не указан - генерируется автоматически)
//
// Переменные окружения:
//
//	SERVER_ADDRESS: адрес сервера
//	BASE_URL: базовый URL
//	FILE_STORAGE_PATH: путь к файлу хранилища
//	DATABASE_DSN: строка подключения к БД
//	ENABLE_HTTPS: включить HTTPS (значение "true" или "1")
//	TLS_CERT_FILE: путь к файлу сертификата TLS (если не указан, генерируется автоматически)
//	TLS_KEY_FILE: путь к файлу приватного ключа TLS (если не указан, генерируется автоматически)
//
// Возвращает конфигурацию приложения или ошибку валидации конфигурации.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	// Флаги
	serverAddress := flag.String("a", "localhost:8080", "server address")
	baseURL := flag.String("b", "http://localhost:8080", "base URL")
	fileStoragePath := flag.String("f", "urls.json", "file storage path")
	databaseDSN := flag.String("d", "", "database DSN")
	enableHTTPS := flag.Bool("s", false, "enable HTTPS")
	tlsCertFile := flag.String("cert", "", "TLS certificate file (if not specified, self-signed certificate will be generated)")
	tlsKeyFile := flag.String("key", "", "TLS private key file (if not specified, self-signed certificate will be generated)")

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
	if envEnableHTTPS := os.Getenv("ENABLE_HTTPS"); envEnableHTTPS != "" {
		*enableHTTPS = envEnableHTTPS == "true" || envEnableHTTPS == "1"
	}
	if envTLSCertFile := os.Getenv("TLS_CERT_FILE"); envTLSCertFile != "" {
		*tlsCertFile = envTLSCertFile
	}
	if envTLSKeyFile := os.Getenv("TLS_KEY_FILE"); envTLSKeyFile != "" {
		*tlsKeyFile = envTLSKeyFile
	}

	cfg.ServerAddress = *serverAddress
	cfg.BaseURL = *baseURL
	cfg.FileStoragePath = *fileStoragePath
	cfg.DatabaseDSN = *databaseDSN
	cfg.EnableHTTPS = *enableHTTPS
	cfg.TLSCertFile = *tlsCertFile
	cfg.TLSKeyFile = *tlsKeyFile

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
	// Файлы сертификатов не обязательны - если они не указаны или не существуют,
	// будет сгенерирован самоподписанный сертификат автоматически
	return nil
}
