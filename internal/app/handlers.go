package app

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vvityuk/shortener/internal/app/middleware"
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

type userURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	originalURL, isDeleted, ok := h.service.GetURL(shortCode)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if isDeleted {
		w.WriteHeader(http.StatusGone)
		return
	}
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	myurl, _ := io.ReadAll(r.Body)
	shortURL, isNew, err := h.service.CreateURL(string(myurl), userID)
	if err != nil {
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	if !isNew {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	w.Write([]byte(h.service.config.BaseURL + "/" + shortURL))
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	shortURL, isNew, err := h.service.CreateURL(req.URL, userID)
	if err != nil {
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	resp := shortenResponse{
		Result: h.service.config.BaseURL + "/" + shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	if !isNew {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
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

	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	result, err := h.service.BatchCreateURL(items, userID)
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

func (h *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	urls, err := h.service.GetUserURLs(userID)
	if err != nil {
		http.Error(w, "Failed to get user URLs", http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]userURLResponse, 0, len(urls))
	for shortURL, originalURL := range urls {
		resp = append(resp, userURLResponse{
			ShortURL:    h.service.config.BaseURL + "/" + shortURL,
			OriginalURL: originalURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var shortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(shortURLs) == 0 {
		http.Error(w, "Empty list of URLs", http.StatusBadRequest)
		return
	}

	h.service.BatchDelete(shortURLs, userID)
	w.WriteHeader(http.StatusAccepted)
}
