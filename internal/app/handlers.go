package app

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

type batchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type batchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
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
	w.Write([]byte(h.service.config.BaseURL + "/" + shortURL))
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	shortURL := h.service.CreateURL(req.URL)
	resp := shortenResponse{
		Result: h.service.config.BaseURL + "/" + shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) PingDB(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(r.Context()); err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) BatchShortenURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req []batchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req) == 0 {
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	items := make(map[string]string)
	for _, item := range req {
		if item.OriginalURL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}
		items[item.CorrelationID] = item.OriginalURL
	}

	result, err := h.service.BatchCreateURL(items)
	if err != nil {
		http.Error(w, "Failed to create short URLs", http.StatusInternalServerError)
		return
	}

	resp := make([]batchResponse, 0, len(result))
	for correlationID, shortURL := range result {
		resp = append(resp, batchResponse{
			CorrelationID: correlationID,
			ShortURL:      h.service.config.BaseURL + "/" + shortURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
