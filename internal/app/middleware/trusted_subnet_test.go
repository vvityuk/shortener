package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrustedSubnetMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		realIP         string
		expectedStatus int
	}{
		{
			name:           "Empty trusted subnet - should deny",
			trustedSubnet:  "",
			realIP:         "192.168.1.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "IP in trusted subnet - should allow",
			trustedSubnet:  "192.168.0.0/16",
			realIP:         "192.168.1.1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP not in trusted subnet - should deny",
			trustedSubnet:  "192.168.0.0/16",
			realIP:         "10.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "No X-Real-IP header - should deny",
			trustedSubnet:  "192.168.0.0/16",
			realIP:         "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid IP address - should deny",
			trustedSubnet:  "192.168.0.0/16",
			realIP:         "invalid-ip",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid CIDR - should deny",
			trustedSubnet:  "invalid-cidr",
			realIP:         "192.168.1.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "IPv6 in trusted subnet - should allow",
			trustedSubnet:  "2001:db8::/32",
			realIP:         "2001:db8::1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IPv6 not in trusted subnet - should deny",
			trustedSubnet:  "2001:db8::/32",
			realIP:         "2001:db9::1",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый handler, который возвращает 200 OK
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Оборачиваем handler в middleware
			middleware := TrustedSubnetMiddleware(tt.trustedSubnet)
			handler := middleware(nextHandler)

			// Создаем тестовый запрос
			req := httptest.NewRequest("GET", "/api/internal/stats", nil)
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			// Создаем ResponseRecorder для записи ответа
			rr := httptest.NewRecorder()

			// Выполняем запрос
			handler.ServeHTTP(rr, req)

			// Проверяем статус код
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

