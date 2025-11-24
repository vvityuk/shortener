package grpc

import (
	"context"
	"os"
	"testing"

	"github.com/vvityuk/shortener/internal/app"
	"github.com/vvityuk/shortener/internal/config"
	"github.com/vvityuk/shortener/internal/grpc/interceptors"
	pb "github.com/vvityuk/shortener/pkg/grpc/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestGRPCServer(t *testing.T) {
	// Создаем временный файл для тестов
	tmpFile, err := os.CreateTemp("", "grpc-test-*.json")
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

	// Создаем сервис
	service, err := app.NewService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()

	// Создаем gRPC сервер
	server := NewServer(service)

	// Тест ShortenURL
	t.Run("ShortenURL", func(t *testing.T) {
		// Создаем контекст с userID
		ctx := context.WithValue(context.Background(), interceptors.GetUserIDKey(), "test-user-1")

		req := &pb.URLShortenRequest{
			Url: "https://example.com",
		}

		resp, err := server.ShortenURL(ctx, req)
		if err != nil {
			t.Errorf("ShortenURL failed: %v", err)
		}

		if resp.Result == "" {
			t.Error("Expected non-empty result")
		}
	})

	// Тест ExpandURL
	t.Run("ExpandURL", func(t *testing.T) {
		// Сначала создаем короткий URL
		ctx := context.WithValue(context.Background(), interceptors.GetUserIDKey(), "test-user-2")

		shortURL, _, err := service.CreateURL("https://example2.com", "test-user-2")
		if err != nil {
			t.Fatal(err)
		}

		req := &pb.URLExpandRequest{
			Id: shortURL,
		}

		resp, err := server.ExpandURL(ctx, req)
		if err != nil {
			t.Errorf("ExpandURL failed: %v", err)
		}

		if resp.Result != "https://example2.com" {
			t.Errorf("Expected https://example2.com, got %s", resp.Result)
		}
	})

	// Тест ListUserURLs
	t.Run("ListUserURLs", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), interceptors.GetUserIDKey(), "test-user-3")

		// Создаем несколько URL для пользователя
		_, _, err := service.CreateURL("https://url1.com", "test-user-3")
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = service.CreateURL("https://url2.com", "test-user-3")
		if err != nil {
			t.Fatal(err)
		}

		req := &emptypb.Empty{}

		resp, err := server.ListUserURLs(ctx, req)
		if err != nil {
			t.Errorf("ListUserURLs failed: %v", err)
		}

		if len(resp.Url) != 2 {
			t.Errorf("Expected 2 URLs, got %d", len(resp.Url))
		}
	})

	// Тест без userID
	t.Run("ShortenURL without userID", func(t *testing.T) {
		ctx := context.Background()

		req := &pb.URLShortenRequest{
			Url: "https://example.com",
		}

		_, err := server.ShortenURL(ctx, req)
		if err == nil {
			t.Error("Expected error for missing userID")
		}
	})
}

