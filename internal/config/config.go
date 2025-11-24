// Package config предоставляет функциональность для загрузки и валидации конфигурации приложения.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// Config содержит конфигурацию приложения.
// Значения могут быть установлены через флаги командной строки или переменные окружения.
type Config struct {
	// ServerAddress адрес и порт для запуска HTTP-сервера (например, "localhost:8080")
	ServerAddress string

	// GRPCAddress адрес и порт для запуска gRPC-сервера (например, "localhost:3200")
	GRPCAddress string

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

	// TrustedSubnet строковое представление доверенной подсети в формате CIDR (например, "192.168.0.0/24")
	// Используется для проверки доступа к internal эндпоинтам
	TrustedSubnet string
}

// jsonConfig представляет структуру конфигурации в JSON формате.
type jsonConfig struct {
	ServerAddress   string `json:"server_address"`
	GRPCAddress     string `json:"grpc_address,omitempty"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
	TLSCertFile     string `json:"tls_cert_file,omitempty"`
	TLSKeyFile      string `json:"tls_key_file,omitempty"`
	TrustedSubnet   string `json:"trusted_subnet,omitempty"`
}

// NewConfig создает новую конфигурацию приложения.
// Читает параметры из JSON файла, флагов командной строки и переменных окружения.
// Приоритет: флаги/переменные окружения > JSON файл > значения по умолчанию.
//
// Флаги командной строки:
//
//	-a: адрес сервера (по умолчанию "localhost:8080")
//	-g: адрес gRPC сервера (по умолчанию пусто)
//	-b: базовый URL (по умолчанию "http://localhost:8080")
//	-f: путь к файлу хранилища (по умолчанию "urls.json")
//	-d: строка подключения к БД (по умолчанию пусто)
//	-s: включить HTTPS режим
//	-c, -config: путь к JSON файлу конфигурации
//	-cert: путь к файлу сертификата TLS (по умолчанию пусто, если не указан - генерируется автоматически)
//	-key: путь к файлу приватного ключа TLS (по умолчанию пусто, если не указан - генерируется автоматически)
//	-t: доверенная подсеть в формате CIDR (по умолчанию пусто)
//
// Переменные окружения:
//
//	CONFIG: путь к JSON файлу конфигурации
//	SERVER_ADDRESS: адрес сервера
//	GRPC_ADDRESS: адрес gRPC сервера
//	BASE_URL: базовый URL
//	FILE_STORAGE_PATH: путь к файлу хранилища
//	DATABASE_DSN: строка подключения к БД
//	ENABLE_HTTPS: включить HTTPS (значение "true" или "1")
//	TLS_CERT_FILE: путь к файлу сертификата TLS (если не указан, генерируется автоматически)
//	TLS_KEY_FILE: путь к файлу приватного ключа TLS (если не указан, генерируется автоматически)
//	TRUSTED_SUBNET: доверенная подсеть в формате CIDR
//
// Формат JSON файла конфигурации:
//
//	{
//	    "server_address": "localhost:8080",
//	    "grpc_address": "localhost:3200",
//	    "base_url": "http://localhost:8080",
//	    "file_storage_path": "urls.json",
//	    "database_dsn": "",
//	    "enable_https": false,
//	    "tls_cert_file": "",
//	    "tls_key_file": "",
//	    "trusted_subnet": ""
//	}
//
// Возвращает конфигурацию приложения или ошибку валидации конфигурации.
func NewConfig() (*Config, error) {
	// Флаги для определения пути к конфигурационному файлу
	configFile := flag.String("c", "", "path to JSON config file")
	configFileAlt := flag.String("config", "", "path to JSON config file (alternative to -c)")

	// Флаги с значениями по умолчанию
	serverAddress := flag.String("a", "", "server address")
	grpcAddress := flag.String("g", "", "gRPC server address")
	baseURL := flag.String("b", "", "base URL")
	fileStoragePath := flag.String("f", "", "file storage path")
	databaseDSN := flag.String("d", "", "database DSN")
	enableHTTPS := flag.Bool("s", false, "enable HTTPS")
	tlsCertFile := flag.String("cert", "", "TLS certificate file (if not specified, self-signed certificate will be generated)")
	tlsKeyFile := flag.String("key", "", "TLS private key file (if not specified, self-signed certificate will be generated)")
	trustedSubnet := flag.String("t", "", "trusted subnet in CIDR notation")

	flag.Parse()

	// Определяем путь к конфигурационному файлу (приоритет: флаг > переменная окружения)
	configPath := *configFile
	if configPath == "" {
		configPath = *configFileAlt
	}
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	// Инициализируем конфигурацию значениями по умолчанию
	cfg := &Config{
		ServerAddress:   "localhost:8080",
		GRPCAddress:     "localhost:3200",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "urls.json",
		DatabaseDSN:     "",
		EnableHTTPS:     false,
		TLSCertFile:     "",
		TLSKeyFile:      "",
		TrustedSubnet:   "",
	}

	// Загружаем конфигурацию из JSON файла (если указан)
	if configPath != "" {
		jsonCfg, err := loadJSONConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configPath, err)
		}
		// Применяем значения из JSON только если они не пустые
		applyJSONConfig(cfg, jsonCfg)
	}

	// Применяем флаги командной строки (перезаписывают JSON)
	// Используем flag.Changed() для проверки, был ли флаг установлен
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.ServerAddress = *serverAddress
		case "g":
			cfg.GRPCAddress = *grpcAddress
		case "b":
			cfg.BaseURL = *baseURL
		case "f":
			cfg.FileStoragePath = *fileStoragePath
		case "d":
			cfg.DatabaseDSN = *databaseDSN
		case "s":
			cfg.EnableHTTPS = *enableHTTPS
		case "cert":
			cfg.TLSCertFile = *tlsCertFile
		case "key":
			cfg.TLSKeyFile = *tlsKeyFile
		case "t":
			cfg.TrustedSubnet = *trustedSubnet
		}
	})

	// Применяем переменные окружения (перезаписывают JSON и флаги)
	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}
	if envGRPCAddress := os.Getenv("GRPC_ADDRESS"); envGRPCAddress != "" {
		cfg.GRPCAddress = envGRPCAddress
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}
	if envEnableHTTPS := os.Getenv("ENABLE_HTTPS"); envEnableHTTPS != "" {
		cfg.EnableHTTPS = envEnableHTTPS == "true" || envEnableHTTPS == "1"
	}
	if envTLSCertFile := os.Getenv("TLS_CERT_FILE"); envTLSCertFile != "" {
		cfg.TLSCertFile = envTLSCertFile
	}
	if envTLSKeyFile := os.Getenv("TLS_KEY_FILE"); envTLSKeyFile != "" {
		cfg.TLSKeyFile = envTLSKeyFile
	}
	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		cfg.TrustedSubnet = envTrustedSubnet
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// loadJSONConfig загружает конфигурацию из JSON файла.
func loadJSONConfig(path string) (*jsonConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var jsonCfg jsonConfig
	if err := json.Unmarshal(data, &jsonCfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	return &jsonCfg, nil
}

// applyJSONConfig применяет значения из JSON конфигурации к Config.
// Применяет только непустые значения.
func applyJSONConfig(cfg *Config, jsonCfg *jsonConfig) {
	if jsonCfg.ServerAddress != "" {
		cfg.ServerAddress = jsonCfg.ServerAddress
	}
	if jsonCfg.GRPCAddress != "" {
		cfg.GRPCAddress = jsonCfg.GRPCAddress
	}
	if jsonCfg.BaseURL != "" {
		cfg.BaseURL = jsonCfg.BaseURL
	}
	if jsonCfg.FileStoragePath != "" {
		cfg.FileStoragePath = jsonCfg.FileStoragePath
	}
	if jsonCfg.DatabaseDSN != "" {
		cfg.DatabaseDSN = jsonCfg.DatabaseDSN
	}
	// Для bool всегда применяем значение из JSON, если оно задано
	cfg.EnableHTTPS = jsonCfg.EnableHTTPS
	if jsonCfg.TLSCertFile != "" {
		cfg.TLSCertFile = jsonCfg.TLSCertFile
	}
	if jsonCfg.TLSKeyFile != "" {
		cfg.TLSKeyFile = jsonCfg.TLSKeyFile
	}
	if jsonCfg.TrustedSubnet != "" {
		cfg.TrustedSubnet = jsonCfg.TrustedSubnet
	}
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
