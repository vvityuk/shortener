package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/vvityuk/shortener/internal/config"
)

func TestHandlers(t *testing.T) {
	// Создаем временный файл для тестов
	tmpFile, err := os.CreateTemp("", "my-test-file.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Создаем конфигурацию
	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
		BaseURL:         "http://localhost:8080",
	}

	// Создаем сервис и обработчик
	service, err := NewService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	handler := NewHandler(service)

	// Тест создания URL
	t.Run("Create URL", func(t *testing.T) {
		originalURL := "https://ya.ru"
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
		w := httptest.NewRecorder()
		handler.CreateURL(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		// Проверяем тело ответа
		responseBody := w.Body.String()
		if !strings.HasPrefix(responseBody, cfg.BaseURL+"/") {
			t.Errorf("Expected response to start with %s, got %s", cfg.BaseURL+"/", responseBody)
		}
	})

	// Тест несуществующего URL
	t.Run("Non-existent URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/not", nil)
		w := httptest.NewRecorder()

		// Создаем Chi контекст для теста
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("shortCode", "not")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.GetURL(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	// Тест создания URL через API
	t.Run("Create URL via API", func(t *testing.T) {
		originalURL := "https://ya.ru"
		reqBody := struct {
			URL string `json:"url"`
		}{
			URL: originalURL,
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ShortenURL(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		// Проверяем заголовок Content-Type
		if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		var resp shortenResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}

		// Проверяем формат ответа
		if !strings.HasPrefix(resp.Result, cfg.BaseURL+"/") {
			t.Errorf("Expected result to start with %s, got %s", cfg.BaseURL+"/", resp.Result)
		}

	})

}
