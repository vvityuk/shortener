package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMainPage(t *testing.T) {

	t.Run("Get long URL", func(t *testing.T) {
		// Сначала создаем короткую ссылку
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://ya.ru"))
		w := httptest.NewRecorder()
		mainPage(w, req)

		shortURL := strings.TrimPrefix(w.Body.String(), "http://localhost:8080")

		// Теперь пытаемся получить длинную ссылку
		req = httptest.NewRequest(http.MethodGet, shortURL, nil)
		w = httptest.NewRecorder()
		mainPage(w, req)

		if w.Code != http.StatusTemporaryRedirect {
			t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
		}

		if w.Header().Get("Location") != "https://ya.ru" {
			t.Errorf("Expected Location header to be https://ya.ru, got %s", w.Header().Get("Location"))
		}
	})

	// Тест несуществующей короткой ссылки
	t.Run("Non-existent short URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/not", nil)
		w := httptest.NewRecorder()
		mainPage(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

}
