package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	// Сохраняем оригинальные значения
	origEnv := make(map[string]string)
	for _, key := range []string{"SERVER_ADDRESS", "BASE_URL", "FILE_STORAGE_PATH", "DATABASE_DSN"} {
		if val, exists := os.LookupEnv(key); exists {
			origEnv[key] = val
			os.Unsetenv(key)
		}
	}
	defer func() {
		// Восстанавливаем оригинальные значения
		for key, val := range origEnv {
			os.Setenv(key, val)
		}
	}()

	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
	}{
		{
			name: "Default values",
			env:  map[string]string{},
		},
		{
			name: "Custom values from env",
			env: map[string]string{
				"SERVER_ADDRESS":    "localhost:9090",
				"BASE_URL":          "http://localhost:9090",
				"FILE_STORAGE_PATH": "test.json",
				"DATABASE_DSN":      "postgres://test",
			},
		},
		{
			name: "Invalid server address",
			env: map[string]string{
				"SERVER_ADDRESS": "",
			},
			wantErr: true,
		},
		{
			name: "Invalid base URL",
			env: map[string]string{
				"BASE_URL": "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем переменные окружения для теста
			for key, val := range tt.env {
				os.Setenv(key, val)
			}

			cfg, err := NewConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cfg == nil {
					t.Error("NewConfig() returned nil config")
					return
				}

				// Проверяем значения
				for key, val := range tt.env {
					switch key {
					case "SERVER_ADDRESS":
						if cfg.ServerAddress != val {
							t.Errorf("ServerAddress = %v, want %v", cfg.ServerAddress, val)
						}
					case "BASE_URL":
						if cfg.BaseURL != val {
							t.Errorf("BaseURL = %v, want %v", cfg.BaseURL, val)
						}
					case "FILE_STORAGE_PATH":
						if cfg.FileStoragePath != val {
							t.Errorf("FileStoragePath = %v, want %v", cfg.FileStoragePath, val)
						}
					case "DATABASE_DSN":
						if cfg.DatabaseDSN != val {
							t.Errorf("DatabaseDSN = %v, want %v", cfg.DatabaseDSN, val)
						}
					}
				}
			}

			// Очищаем переменные окружения после теста
			for key := range tt.env {
				os.Unsetenv(key)
			}
		})
	}
}
