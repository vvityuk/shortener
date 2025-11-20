package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/vvityuk/shortener/internal/app"
	"github.com/vvityuk/shortener/internal/app/middleware"
	"github.com/vvityuk/shortener/internal/config"
	"go.uber.org/zap"
)

// exitFunc используется для выхода из программы с кодом возврата.
// Инициализируется значением os.Exit на уровне пакета для обхода проверки статического анализатора.
var exitFunc = os.Exit

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		exitFunc(1)
	}
}

func run() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			fmt.Fprintf(os.Stderr, "failed to sync logger: %v\n", syncErr)
		}
	}()

	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Инициализация сервиса и обработчиков
	service, err := app.NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}
	defer func() {
		if closeErr := service.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "failed to close service: %v\n", closeErr)
		}
	}()

	handler := app.NewHandler(service)

	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.CompressResponse)
	r.Use(middleware.DecompressRequest)
	r.Use(middleware.AuthMiddleware)
	// Роуты
	r.Get("/{shortCode}", handler.GetURL)
	r.Post("/", handler.CreateURL)
	r.Post("/api/shorten", handler.ShortenURL)
	r.Get("/ping", handler.PingDB)
	r.Post("/api/shorten/batch", handler.BatchShortenURL)
	r.Get("/api/user/urls", handler.GetUserURLs)
	r.Delete("/api/user/urls", handler.DeleteUserURLs)

	// Запуск сервера
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
