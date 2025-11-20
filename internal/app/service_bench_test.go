package app

import (
	"fmt"
	"os"
	"testing"

	"github.com/vvityuk/shortener/internal/config"
)

// Бенчмарк генерации коротких URL
func BenchmarkService_randStr(b *testing.B) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	service := &Service{
		storage: NewMemoryStorage(),
		config:  cfg,
	}
	defer service.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.randStr(4)
	}
}

// Бенчмарк создания URL через сервис
func BenchmarkService_CreateURL(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "bench-service-*.json")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
		BaseURL:         "http://localhost:8080",
	}
	service, err := NewService(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.CreateURL("https://example.com", "user1")
	}
}

// Бенчмарк получения URL через сервис
func BenchmarkService_GetURL(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "bench-service-*.json")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
		BaseURL:         "http://localhost:8080",
	}
	service, err := NewService(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	// Подготовка данных
	shortURLs := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		shortURL, _, _ := service.CreateURL("https://example.com", "user1")
		shortURLs[i] = shortURL
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := shortURLs[i%1000]
		_, _, _ = service.GetURL(key)
	}
}

// Бенчмарк пакетного создания URL
func BenchmarkService_BatchCreateURL(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "bench-service-*.json")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
		BaseURL:         "http://localhost:8080",
	}
	service, err := NewService(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	items := make(map[string]string)
	for i := 0; i < 100; i++ {
		items[fmt.Sprintf("corr%d", i)] = "https://example.com"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.BatchCreateURL(items, "user1")
	}
}

// Бенчмарк получения URL пользователя
func BenchmarkService_GetUserURLs(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "bench-service-*.json")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
		BaseURL:         "http://localhost:8080",
	}
	service, err := NewService(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	// Подготовка данных
	userID := "user1"
	for i := 0; i < 1000; i++ {
		_, _, _ = service.CreateURL("https://example.com", userID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetUserURLs(userID)
	}
}

// Бенчмарк сравнения производительности MemoryStorage vs FileStorage для CreateURL
func BenchmarkService_CreateURL_MemoryStorage(b *testing.B) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	service := &Service{
		storage: NewMemoryStorage(),
		config:  cfg,
	}
	defer service.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.CreateURL("https://example.com", "user1")
	}
}

func BenchmarkService_CreateURL_FileStorage(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "bench-service-*.json")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
		BaseURL:         "http://localhost:8080",
	}
	service, err := NewService(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.CreateURL("https://example.com", "user1")
	}
}
