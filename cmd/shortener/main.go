package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vvityuk/shortener/internal/app"
	"github.com/vvityuk/shortener/internal/config"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		panic("Failed to initialize config")
	}
	// Инициализация сервиса и обработчиков
	service := app.NewService(cfg)
	handler := app.NewHandler(service)

	r := chi.NewRouter()

	// Роуты
	r.Get("/{shortCode}", handler.GetURL)
	r.Post("/", handler.CreateURL)

	// Запуск сервера
	err = http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil {
		panic(err)
	}
}
