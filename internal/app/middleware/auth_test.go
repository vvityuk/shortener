package middleware

import (
	"net/http"
	"net/http/httptest"
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
