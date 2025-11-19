package app_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/vvityuk/shortener/internal/app"
	"github.com/vvityuk/shortener/internal/app/middleware"
	"github.com/vvityuk/shortener/internal/config"
)

func ExampleHandler_CreateURL() {
	// Создаем временный файл для хранилища
	tmpFile, _ := os.CreateTemp("", "example-*.json")
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Инициализируем конфигурацию и сервис
	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
		BaseURL:         "http://localhost:8080",
	}
	service, _ := app.NewService(cfg)
	defer service.Close()

	handler := app.NewHandler(service)

	// Создаем POST-запрос с URL в теле
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req.AddCookie(&http.Cookie{Name: middleware.ChiookieName, Value: "test-user-id"})
	w := httptest.NewRecorder()

	handler.CreateURL(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response contains base URL: %v\n", strings.Contains(string(body), "http://localhost:8080/"))
	// Output:
	// Status: 201
	// Response contains base URL: true
}

func ExampleHandler_ShortenURL() {
	// Создаем временный файл для хранилища
	tmpFile, _ := os.CreateTemp("", "example-*.json")
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Инициализируем конфигурацию и сервис
	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
		BaseURL:         "http://localhost:8080",
	}
	service, _ := app.NewService(cfg)
	defer service.Close()

	handler := app.NewHandler(service)

	// Создаем JSON-запрос
	requestBody := map[string]string{
		"url": "https://example.com",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: middleware.ChiookieName, Value: "test-user-id"})
	w := httptest.NewRecorder()

	handler.ShortenURL(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var response map[string]string
	json.NewDecoder(resp.Body).Decode(&response)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Result contains base URL: %v\n", strings.Contains(response["result"], "http://localhost:8080/"))
	// Output:
	// Status: 201
	// Result contains base URL: true
}

func ExampleHandler_BatchShortenURL() {
	// Создаем временный файл для хранилища
	tmpFile, _ := os.CreateTemp("", "example-*.json")
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Инициализируем конфигурацию и сервис
	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
		BaseURL:         "http://localhost:8080",
	}
	service, _ := app.NewService(cfg)
	defer service.Close()

	handler := app.NewHandler(service)

	// Создаем пакетный запрос
	requestBody := []map[string]string{
		{"correlation_id": "1", "original_url": "https://example.com"},
		{"correlation_id": "2", "original_url": "https://example.org"},
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: middleware.ChiookieName, Value: "test-user-id"})
	w := httptest.NewRecorder()

	handler.BatchShortenURL(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var response []map[string]string
	json.NewDecoder(resp.Body).Decode(&response)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Created %d URLs\n", len(response))
	// Output:
	// Status: 201
	// Created 2 URLs
}

func ExampleHandler_GetUserURLs() {
	// Создаем временный файл для хранилища
	tmpFile, _ := os.CreateTemp("", "example-*.json")
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Инициализируем конфигурацию и сервис
	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
		BaseURL:         "http://localhost:8080",
	}
	service, _ := app.NewService(cfg)
	defer service.Close()

	handler := app.NewHandler(service)

	// Сначала создаем несколько URL
	userID := "test-user-id"
	urls := []string{"https://example.com", "https://example.org"}
	for _, url := range urls {
		requestBody := map[string]string{"url": url}
		jsonBody, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: middleware.ChiookieName, Value: userID})
		w := httptest.NewRecorder()
		handler.ShortenURL(w, req)
	}

	// Получаем список URL пользователя
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req.AddCookie(&http.Cookie{Name: middleware.ChiookieName, Value: userID})
	w := httptest.NewRecorder()

	handler.GetUserURLs(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var response []map[string]string
	json.NewDecoder(resp.Body).Decode(&response)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("User has %d URLs\n", len(response))
	// Output:
	// Status: 200
	// User has 2 URLs
}

func ExampleHandler_DeleteUserURLs() {
	// Создаем временный файл для хранилища
	tmpFile, _ := os.CreateTemp("", "example-*.json")
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Инициализируем конфигурацию и сервис
	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
		BaseURL:         "http://localhost:8080",
	}
	service, _ := app.NewService(cfg)
	defer service.Close()

	handler := app.NewHandler(service)

	// Создаем URL для удаления
	userID := "test-user-id"
	requestBody := map[string]string{"url": "https://example.com"}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: middleware.ChiookieName, Value: userID})
	w := httptest.NewRecorder()
	handler.ShortenURL(w, req)

	var createResp map[string]string
	json.NewDecoder(w.Result().Body).Decode(&createResp)
	shortCode := strings.TrimPrefix(createResp["result"], cfg.BaseURL+"/")

	// Удаляем URL
	deleteBody := []string{shortCode}
	jsonDeleteBody, _ := json.Marshal(deleteBody)
	req = httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(jsonDeleteBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: middleware.ChiookieName, Value: userID})
	w = httptest.NewRecorder()

	handler.DeleteUserURLs(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Status: 202
}
