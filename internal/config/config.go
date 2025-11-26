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
//	    "tls_key_file": ""
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

	// Настраиваем чтение переменных окружения
	// Viper автоматически преобразует UPPER_CASE в lowercase с подчеркиваниями
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Настраиваем флаги командной строки через pflag
	// Используем pflag.CommandLine для совместимости со стандартным flag
	pflag.CommandLine = pflag.NewFlagSet("shortener", pflag.ExitOnError)
	pflag.StringP("config", "c", "", "path to JSON config file")
	pflag.StringP("server-address", "a", "", "server address")
	pflag.StringP("base-url", "b", "", "base URL")
	pflag.StringP("file-storage-path", "f", "", "file storage path")
	pflag.StringP("database-dsn", "d", "", "database DSN")
	pflag.BoolP("enable-https", "s", false, "enable HTTPS")
	pflag.String("cert", "", "TLS certificate file")
	pflag.String("key", "", "TLS private key file")

	// Парсим флаги из командной строки
	pflag.Parse()

	// Привязываем флаги к Viper
	// Viper автоматически преобразует дефисы в подчеркивания
	if err := v.BindPFlags(pflag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Определяем путь к конфигурационному файлу (приоритет: флаг > переменная окружения)
	configPath := v.GetString("config")
	if configPath == "" {
		configPath = v.GetString("c")
	}
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
	// Используем os.Getenv напрямую для проверки наличия переменных окружения
	// и устанавливаем их через Set() для наивысшего приоритета
	envVars := map[string]string{
		"SERVER_ADDRESS":    "server_address",
		"BASE_URL":          "base_url",
		"FILE_STORAGE_PATH": "file_storage_path",
		"DATABASE_DSN":      "database_dsn",
		"TLS_CERT_FILE":     "tls_cert_file",
		"TLS_KEY_FILE":      "tls_key_file",
	}

	for envKey, viperKey := range envVars {
		if val := os.Getenv(envKey); val != "" {
			v.Set(viperKey, val)
		}
	}

	// Специальная обработка для bool переменной
	if val := os.Getenv("ENABLE_HTTPS"); val != "" {
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
