// Package app предоставляет бизнес-логику и HTTP-обработчики для сервиса сокращения URL.
package app

import (
	"context"
	"time"

	"github.com/vvityuk/shortener/internal/config"
	"github.com/vvityuk/shortener/internal/storage/postgres"
	"golang.org/x/exp/rand"
)

// Service предоставляет бизнес-логику для работы с короткими URL.
// Поддерживает создание, получение, пакетную обработку и удаление URL.
type Service struct {
	storage Storage
	config  *config.Config
}

// NewService создает новый экземпляр Service с автоматическим выбором хранилища.
// Приоритет выбора хранилища:
//  1. PostgreSQL (если указан DatabaseDSN)
//  2. Файловое хранилище (если указан FileStoragePath)
//  3. Хранилище в памяти (по умолчанию)
//
// Возвращает новый экземпляр сервиса или ошибку при инициализации хранилища.
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

// GetURL получает оригинальный URL по короткому коду.
// Возвращает оригинальный URL, флаг удаления (true если URL был удален)
// и флаг успешного получения (true если URL найден).
func (s *Service) GetURL(shortCode string) (string, bool, bool) {
	return s.storage.Get(shortCode)
}

// CreateURL создает короткий URL для указанного оригинального URL.
// Если URL уже существует для данного пользователя, возвращает существующий короткий код.
// Возвращает короткий код URL, флаг создания нового URL (true если создан новый,
// false если уже существовал) и ошибку при создании URL.
func (s *Service) CreateURL(longURL string, userID string) (string, bool, error) {
	shortURL := s.randStr(4)
	return s.storage.Save(shortURL, longURL, userID)
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

// Close закрывает соединение с хранилищем и освобождает ресурсы.
// Возвращает ошибку при закрытии хранилища.
func (s *Service) Close() error {
	return s.storage.Close()
}

// Ping проверяет доступность хранилища данных.
// Возвращает ошибку при проверке доступности хранилища.
func (s *Service) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}

// BatchCreateURL создает несколько коротких URL за один запрос.
// Принимает карту соответствий correlation_id -> original_url и идентификатор пользователя.
// Возвращает карту соответствий correlation_id -> short_code и ошибку при создании URL.
func (s *Service) BatchCreateURL(items map[string]string, userID string) (map[string]string, error) {
	result := make(map[string]string)
	urls := make(map[string]string)

	for correlationID, originalURL := range items {
		shortURL := s.randStr(4)
		urls[shortURL] = originalURL
		result[correlationID] = shortURL
	}

	if err := s.storage.BatchSave(urls, userID); err != nil {
		return nil, err
	}

	return result, nil
}

// GetUserURLs возвращает все короткие URL, созданные указанным пользователем.
// Возвращает карту соответствий short_code -> original_url и ошибку при получении URL.
func (s *Service) GetUserURLs(userID string) (map[string]string, error) {
	return s.storage.GetUserURLs(userID)
}

// BatchDelete удаляет указанные короткие URL пользователя.
// Удаление выполняется асинхронно в отдельной горутине.
func (s *Service) BatchDelete(shortURLs []string, userID string) {
	go func() {
		_ = s.storage.BatchDelete(shortURLs, userID)
	}()
}

// GetStats возвращает статистику сервиса: количество сокращённых URL и количество пользователей.
// Возвращает количество URL, количество пользователей и ошибку при получении статистики.
func (s *Service) GetStats() (int, int, error) {
	return s.storage.GetStats()
}

// GetBaseURL возвращает базовый URL для генерации коротких ссылок.
func (s *Service) GetBaseURL() string {
	return s.config.BaseURL
}
