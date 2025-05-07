package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vvityuk/shortener/internal/app"
	"github.com/vvityuk/shortener/internal/app/middleware"
	"github.com/vvityuk/shortener/internal/config"
	"go.uber.org/zap"
)

func main() {

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	cfg, err := config.NewConfig()
	if err != nil {
		panic("Failed to initialize config")
	}
	// Инициализация сервиса и обработчиков
	service, err := app.NewService(cfg)
	if err != nil {
		panic("Failed to initialize service")
	}
	defer service.Close()

	handler := app.NewHandler(service)

	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.CompressResponse)
	r.Use(middleware.DecompressRequest)
	// Роуты
	r.Get("/{shortCode}", handler.GetURL)
	r.Post("/", handler.CreateURL)
	r.Post("/api/shorten", handler.ShortenURL)
	r.Get("/ping", handler.PingDB)
	r.Post("/api/shorten/batch", handler.BatchShortenURL)

	// Запуск сервера
	err = http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil {
		panic(err)
	}
}
