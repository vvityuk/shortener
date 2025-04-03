package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/vvityuk/shortener/internal/app"
	"github.com/vvityuk/shortener/internal/config"
)

func TestHandlers(t *testing.T) {
	cfg, _ := config.NewConfig()
	service := app.NewService(cfg)
	handler := app.NewHandler(service)

	// Тест получения длинной ссылки
	t.Run("Get long URL", func(t *testing.T) {
		// Проверим создание короткой ссылки
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://ya.ru"))
		w := httptest.NewRecorder()
		handler.CreateURL(w, req)

	})

	// Тест несуществующей короткой ссылки
	t.Run("Non-existent short URL", func(t *testing.T) {
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

}
