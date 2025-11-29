package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		cookie         *http.Cookie
		expectedStatus int
		checkCookie    bool
	}{
		{
			name:           "No cookie",
			cookie:         nil,
			expectedStatus: http.StatusOK,
			checkCookie:    true,
		},
		{
			name:           "Valid cookie",
			cookie:         &http.Cookie{Name: ChiookieName, Value: "test-user-id"},
			expectedStatus: http.StatusOK,
			checkCookie:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			w := httptest.NewRecorder()
			handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkCookie {
				resp := w.Result()
				defer resp.Body.Close()
				cookies := resp.Cookies()
				if len(cookies) != 1 {
					t.Errorf("Expected 1 cookie, got %d", len(cookies))
				}
				if cookies[0].Name != ChiookieName {
					t.Errorf("Expected cookie name %s, got %s", ChiookieName, cookies[0].Name)
				}
				if cookies[0].Value == "" {
					t.Error("Expected non-empty cookie value")
				}
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name     string
		cookie   *http.Cookie
		expected string
	}{
		{
			name:     "No cookie",
			cookie:   nil,
			expected: "",
		},
		{
			name:     "Valid cookie",
			cookie:   &http.Cookie{Name: ChiookieName, Value: "test-user-id"},
			expected: "test-user-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			userID := GetUserID(req)
			if userID != tt.expected {
				t.Errorf("Expected user ID %s, got %s", tt.expected, userID)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	// Инициализируем JWT Manager для тестов
	testSecretKey := "test-secret-key-minimum-32-characters-long"
	if err := InitJWTManager(testSecretKey); err != nil {
		t.Fatalf("Failed to initialize JWT manager: %v", err)
	}

	userID, token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error = %v, want nil", err)
	}

	if userID == "" {
		t.Error("GenerateToken() userID is empty")
	}

	if token == "" {
		t.Error("GenerateToken() token is empty")
	}

	// Проверяем, что токен валидный
	validatedUserID, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v, want nil", err)
	}

	if validatedUserID != userID {
		t.Errorf("ValidateToken() userID = %s, want %s", validatedUserID, userID)
	}
}

func TestValidateToken(t *testing.T) {
	// Инициализируем JWT Manager для тестов
	testSecretKey := "test-secret-key-minimum-32-characters-long"
	if err := InitJWTManager(testSecretKey); err != nil {
		t.Fatalf("Failed to initialize JWT manager: %v", err)
	}

	tests := []struct {
		name    string
		setup   func() string
		wantErr bool
		errType error
	}{
		{
			name: "Valid token",
			setup: func() string {
				_, token, _ := GenerateToken()
				return token
			},
			wantErr: false,
		},
		{
			name: "Invalid token - empty string",
			setup: func() string {
				return ""
			},
			wantErr: true,
			errType: ErrTokenInvalid,
		},
		{
			name: "Invalid token - malformed",
			setup: func() string {
				return "invalid.token.string"
			},
			wantErr: true,
			errType: ErrTokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setup()
			userID, err := ValidateToken(token)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateToken() expected error, got nil")
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("ValidateToken() error = %v, want %v", err, tt.errType)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateToken() error = %v, want nil", err)
				}
				if userID == "" {
					t.Error("ValidateToken() userID is empty")
				}
			}
		})
	}
}

func TestInitJWTManager(t *testing.T) {
	tests := []struct {
		name      string
		secretKey string
		wantErr   bool
	}{
		{
			name:      "Valid secret key",
			secretKey: "valid-secret-key-minimum-32-characters-long",
			wantErr:   false,
		},
		{
			name:      "Empty secret key",
			secretKey: "",
			wantErr:   true,
		},
		{
			name:      "Short secret key",
			secretKey: "short",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сбрасываем глобальное состояние для каждого теста
			// Используем новый sync.Once для каждого теста
			jwtManager = nil
			jwtOnce = sync.Once{}

			err := InitJWTManager(tt.secretKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitJWTManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
