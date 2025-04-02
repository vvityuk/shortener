package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vvityuk/shortener/internal/app"
)

func main() {
	// Инициализация сервиса и обработчиков
	service := app.NewService()
	handler := app.NewHandler(service)

	r := chi.NewRouter()

	// Роуты
	r.Get("/{shortCode}", handler.GetURL)
	r.Post("/", handler.CreateURL)

	// Запуск сервера
	err := http.ListenAndServe("localhost:8080", r)
	if err != nil {
		panic(err)
	}
}
