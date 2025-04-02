package app

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	val, ok := h.service.GetURL(shortCode)
	if ok {
		w.Header().Set("Location", val)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (h *Handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	myurl, _ := io.ReadAll(r.Body)
	shortURL := h.service.CreateURL(string(myurl))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8080/" + shortURL))
}
