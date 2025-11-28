// Package config предоставляет функциональность для загрузки и валидации конфигурации приложения.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config содержит конфигурацию приложения.
// Значения могут быть установлены через флаги командной строки или переменные окружения.
type Config struct {
	// ServerAddress адрес и порт для запуска HTTP-сервера (например, "localhost:8080")
	ServerAddress string `mapstructure:"server_address"`

	// BaseURL базовый URL для генерации коротких ссылок (например, "http://localhost:8080")
	BaseURL string `mapstructure:"base_url"`

	// FileStoragePath путь к файлу для хранения данных (например, "urls.json")
	// Если пусто, используется хранилище в памяти
	FileStoragePath string `mapstructure:"file_storage_path"`

	// DatabaseDSN строка подключения к PostgreSQL (например, "postgres://user:pass@localhost/dbname")
	// Если указана, имеет приоритет над файловым хранилищем
	DatabaseDSN string `mapstructure:"database_dsn"`

	// EnableHTTPS включает HTTPS режим работы сервера
	EnableHTTPS bool `mapstructure:"enable_https"`

	// TLSCertFile путь к файлу сертификата TLS (например, "server.crt")
	TLSCertFile string `mapstructure:"tls_cert_file"`

	// TLSKeyFile путь к файлу приватного ключа TLS (например, "server.key")
	TLSKeyFile string `mapstructure:"tls_key_file"`

	// TrustedSubnet строковое представление доверенной подсети в формате CIDR (например, "192.168.0.0/24")
	// Используется для проверки доступа к internal эндпоинтам
	TrustedSubnet string `mapstructure:"trusted_subnet"`

	// GRPCAddress адрес и порт для запуска grpc-сервера (например, "localhost:3200")
	GRPCAddress string `mapstructure:"grpc_address"`
}

// NewConfig создает новую конфигурацию приложения.
// Читает параметры из JSON файла, флагов командной строки и переменных окружения.
// Приоритет: переменные окружения > флаги > JSON файл > значения по умолчанию.
//
// Флаги командной строки:
//
//	-a: адрес сервера (по умолчанию "localhost:8080")
//	-b: базовый URL (по умолчанию "http://localhost:8080")
//	-f: путь к файлу хранилища (по умолчанию "urls.json")
//	-d: строка подключения к БД (по умолчанию пусто)
//	-s: включить HTTPS режим
//	-c, -config: путь к JSON файлу конфигурации
//	-cert: путь к файлу сертификата TLS (по умолчанию пусто, если не указан - генерируется автоматически)
//	-key: путь к файлу приватного ключа TLS (по умолчанию пусто, если не указан - генерируется автоматически)
//
// Переменные окружения:
//
//	CONFIG: путь к JSON файлу конфигурации
//	SERVER_ADDRESS: адрес сервера
//	BASE_URL: базовый URL
//	FILE_STORAGE_PATH: путь к файлу хранилища
//	DATABASE_DSN: строка подключения к БД
//	ENABLE_HTTPS: включить HTTPS (значение "true" или "1")
//	TLS_CERT_FILE: путь к файлу сертификата TLS (если не указан, генерируется автоматически)
//	TLS_KEY_FILE: путь к файлу приватного ключа TLS (если не указан, генерируется автоматически)
//
// Формат JSON файла конфигурации:
//
//	{
//	    "server_address": "localhost:8080",
//	    "base_url": "http://localhost:8080",
//	    "file_storage_path": "urls.json",
//	    "database_dsn": "",
//	    "enable_https": false,
//	    "tls_cert_file": "",
//	    "tls_key_file": "",
//		"trusted_subnet": "",
//		"grpc_address": ""
//	}
//
// Возвращает конфигурацию приложения или ошибку валидации конфигурации.
func NewConfig() (*Config, error) {
	v := viper.New()

	// Устанавливаем значения по умолчанию
	v.SetDefault("server_address", "localhost:8080")
	v.SetDefault("base_url", "http://localhost:8080")
	v.SetDefault("file_storage_path", "urls.json")
	v.SetDefault("database_dsn", "")
	v.SetDefault("enable_https", false)
	v.SetDefault("tls_cert_file", "")
	v.SetDefault("tls_key_file", "")
	v.SetDefault("trusted_subnet", "")
	v.SetDefault("grpc_address", "localhost:3200")

	// Настраиваем чтение переменных окружения
	// Viper автоматически преобразует UPPER_CASE в lowercase с подчеркиваниями
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Настраиваем флаги командной строки через pflag
	// Создаем новый FlagSet для избежания конфликтов при повторных вызовах
	flags := pflag.NewFlagSet("shortener", pflag.ContinueOnError)
	configFlag := flags.StringP("config", "c", "", "path to JSON config file")
	serverAddressFlag := flags.StringP("server-address", "a", "", "server address")
	baseURLFlag := flags.StringP("base-url", "b", "", "base URL")
	fileStoragePathFlag := flags.StringP("file-storage-path", "f", "", "file storage path")
	databaseDSNFlag := flags.StringP("database-dsn", "d", "", "database DSN")
	enableHTTPSFlag := flags.BoolP("enable-https", "s", false, "enable HTTPS")
	certFlag := flags.String("cert", "", "TLS certificate file")
	keyFlag := flags.String("key", "", "TLS private key file")
	keyTrustedSubnet := flags.String("t", "", "trusted subnet in CIDR notation")
	grpcAddressFlag := flags.String("g", "", "trusted subnet in CIDR notation")

	// Парсим флаги из командной строки
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == pflag.ErrHelp {
			os.Exit(0)
		}
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	// Устанавливаем значения из флагов в Viper (если флаги были установлены)
	// Используем Changed() чтобы применять только установленные флаги
	if flags.Changed("server-address") {
		v.Set("server_address", *serverAddressFlag)
	}
	if flags.Changed("base-url") {
		v.Set("base_url", *baseURLFlag)
	}
	if flags.Changed("file-storage-path") {
		v.Set("file_storage_path", *fileStoragePathFlag)
	}
	if flags.Changed("database-dsn") {
		v.Set("database_dsn", *databaseDSNFlag)
	}
	if flags.Changed("enable-https") {
		v.Set("enable_https", *enableHTTPSFlag)
	}
	if flags.Changed("cert") {
		v.Set("tls_cert_file", *certFlag)
	}
	if flags.Changed("key") {
		v.Set("tls_key_file", *keyFlag)
	}
	if flags.Changed("t") {
		v.Set("trusted_subnet", *keyTrustedSubnet)
	}
	if flags.Changed("g") {
		v.Set("grpc_address", *grpcAddressFlag)
	}

	// Определяем путь к конфигурационному файлу (приоритет: флаг > переменная окружения)
	configPath := *configFlag
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	// Загружаем конфигурацию из JSON файла (если указан)
	if configPath != "" {
		v.SetConfigFile(configPath)
		v.SetConfigType("json")
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configPath, err)
		}
	}

	// Применяем переменные окружения с наивысшим приоритетом
	// Viper по умолчанию имеет приоритет: flags > env > config > defaults
	// Нам нужен: env > flags > config > defaults
	// Поэтому после чтения всех источников, вручную проверяем env и устанавливаем через Set()
	applyEnvOverrides(v)

	// Unmarshal конфигурации в структуру
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Валидация конфигурации
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// applyEnvOverrides применяет переменные окружения с наивысшим приоритетом.
// Это необходимо, т.к. Viper по умолчанию имеет приоритет flags > env > config > defaults,
// а нам нужен env > flags > config > defaults.
func applyEnvOverrides(v *viper.Viper) {
	// Используем os.LookupEnv для проверки наличия переменных окружения
	// (отличает пустую переменную от необъявленной) и устанавливаем их через Set() для наивысшего приоритета
	envVars := map[string]string{
		"SERVER_ADDRESS":    "server_address",
		"BASE_URL":          "base_url",
		"FILE_STORAGE_PATH": "file_storage_path",
		"DATABASE_DSN":      "database_dsn",
		"TLS_CERT_FILE":     "tls_cert_file",
		"TLS_KEY_FILE":      "tls_key_file",
		"TRUSTED_SUBNET":    "trusted_subnet",
		"GRPC_ADDRESS":      "grpc_address",
	}

	for envKey, viperKey := range envVars {
		if val, ok := os.LookupEnv(envKey); ok {
			v.Set(viperKey, val)
		}
	}

	// Специальная обработка для bool переменной
	if val, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		v.Set("enable_https", val == "true" || val == "1")
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
