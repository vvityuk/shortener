package app

import (
	"context"
	"time"

	"github.com/vvityuk/shortener/internal/config"
	"github.com/vvityuk/shortener/internal/storage/postgres"
	"golang.org/x/exp/rand"
)

type Service struct {
	storage Storage
	config  *config.Config
}

func NewService(cfg *config.Config) (*Service, error) {
	var storage Storage
	var err error

	// Пробуем PostgreSQL
	if cfg.DatabaseDSN != "" {
		storage, err = postgres.New(cfg.DatabaseDSN)
		if err == nil {
			return &Service{storage: storage, config: cfg}, nil
		}
	}

	// Пробуем файловое хранилище
	if cfg.FileStoragePath != "" {
		storage, err = NewStorage(cfg.FileStoragePath)
		if err == nil {
			return &Service{storage: storage, config: cfg}, nil
		}
	}

	// Используем хранилище в памяти
	storage = NewMemoryStorage()
	return &Service{storage: storage, config: cfg}, nil
}

func (s *Service) GetURL(shortCode string) (string, bool) {
	return s.storage.Get(shortCode)
}

func (s *Service) CreateURL(longURL string) (string, bool, error) {
	shortURL := s.randStr(4)
	return s.storage.Save(shortURL, longURL)
}

func (s *Service) randStr(n int) string {
	rnd := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))

	letters := []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rnd.Intn(len(letters))]
	}
	return string(b)
}

func (s *Service) Close() error {
	return s.storage.Close()
}

func (s *Service) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}

func (s *Service) BatchCreateURL(items map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	urls := make(map[string]string)

	for correlationID, originalURL := range items {
		shortURL := s.randStr(4)
		urls[shortURL] = originalURL
		result[correlationID] = shortURL
	}

	if err := s.storage.BatchSave(urls); err != nil {
		return nil, err
	}

	return result, nil
}
