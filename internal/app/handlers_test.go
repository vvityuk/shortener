package app

import (
	"bytes"
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
	t.Run("Create_URL_via_API", func(t *testing.T) {
		req := shortenRequest{URL: "https://example.com"}
		body, _ := json.Marshal(req)

		request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ShortenURL(w, request)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 201 or 409, got %d", resp.StatusCode)
		}

		var response shortenResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}

		if response.Result == "" {
			t.Error("Expected non-empty result")
		}
	})

	// Тест проверки подключения к БД
	t.Run("Ping DB", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		handler.PingDB(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	// Тест пакетного создания URL
	t.Run("Batch Shorten URL", func(t *testing.T) {
		batchReq := []batchRequest{
			{CorrelationID: "1", OriginalURL: "https://example1.com"},
			{CorrelationID: "2", OriginalURL: "https://example2.com"},
		}
		body, _ := json.Marshal(batchReq)

		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.BatchShortenURL(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response []batchResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}

		if len(response) != 2 {
			t.Errorf("Expected 2 responses, got %d", len(response))
		}

		for _, resp := range response {
			if resp.ShortURL == "" {
				t.Error("Expected non-empty short URL")
			}
		}
	})

	// Тест пакетного создания URL с пустым запросом
	t.Run("Batch Shorten URL Empty Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader([]byte("[]")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.BatchShortenURL(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	// Тест пакетного создания URL с некорректным JSON
	t.Run("Batch Shorten URL Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.BatchShortenURL(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}
