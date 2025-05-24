package app

import (
	"context"
	"encoding/json"
	"os"
)

type Storage interface {
	Get(key string) (string, bool, bool)
	Save(key, value string, userID string) (string, bool, error)
	GetByOriginalURL(originalURL string) (string, bool)
	BatchSave(items map[string]string, userID string) error
	GetUserURLs(userID string) (map[string]string, error)
	Close() error
	Ping(ctx context.Context) error
	BatchDelete(shortURLs []string, userID string) error
}

type FileStorage struct {
	urls map[string]struct {
		OriginalURL string
		UserID      string
	}
	file *os.File
}

func NewStorage(filePath string) (*FileStorage, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	storage := &FileStorage{
		urls: make(map[string]struct {
			OriginalURL string
			UserID      string
		}),
		file: file,
	}

	if err := storage.load(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *FileStorage) Get(key string) (string, bool, bool) {
	val, ok := s.urls[key]
	return val.OriginalURL, false, ok
}

func (s *FileStorage) Save(key, value string, userID string) (string, bool, error) {
	if existingKey, ok := s.getKeyByValueAndUser(value, userID); ok {
		return existingKey, false, nil
	}
	s.urls[key] = struct {
		OriginalURL string
		UserID      string
	}{
		OriginalURL: value,
		UserID:      userID,
	}
	if err := s.save(); err != nil {
		return "", false, err
	}
	return key, true, nil
}

func (s *FileStorage) GetByOriginalURL(originalURL string) (string, bool) {
	for key, val := range s.urls {
		if val.OriginalURL == originalURL {
			return key, true
		}
	}
	return "", false
}

func (s *FileStorage) getKeyByValueAndUser(value, userID string) (string, bool) {
	for key, val := range s.urls {
		if val.OriginalURL == value && val.UserID == userID {
			return key, true
		}
	}
	return "", false
}

func (s *FileStorage) BatchSave(items map[string]string, userID string) error {
	for key, value := range items {
		s.urls[key] = struct {
			OriginalURL string
			UserID      string
		}{
			OriginalURL: value,
			UserID:      userID,
		}
	}
	return s.save()
}

func (s *FileStorage) GetUserURLs(userID string) (map[string]string, error) {
	urls := make(map[string]string)
	for key, val := range s.urls {
		if val.UserID == userID {
			urls[key] = val.OriginalURL
		}
	}
	return urls, nil
}

func (s *FileStorage) load() error {
	stat, err := s.file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() == 0 {
		return nil
	}

	decoder := json.NewDecoder(s.file)
	return decoder.Decode(&s.urls)
}

func (s *FileStorage) save() error {
	if err := s.file.Truncate(0); err != nil {
		return err
	}
	if _, err := s.file.Seek(0, 0); err != nil {
		return err
	}
	encoder := json.NewEncoder(s.file)
	return encoder.Encode(s.urls)
}

func (s *FileStorage) Close() error {
	return s.file.Close()
}

func (s *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *FileStorage) BatchDelete(shortURLs []string, userID string) error {
	return nil
}

type MemoryStorage struct {
	urls map[string]struct {
		OriginalURL string
		UserID      string
	}
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urls: make(map[string]struct {
			OriginalURL string
			UserID      string
		}),
	}
}

func (s *MemoryStorage) Get(key string) (string, bool, bool) {
	val, ok := s.urls[key]
	return val.OriginalURL, false, ok
}

func (s *MemoryStorage) Save(key, value string, userID string) (string, bool, error) {
	if existingKey, ok := s.getKeyByValueAndUser(value, userID); ok {
		return existingKey, false, nil
	}
	s.urls[key] = struct {
		OriginalURL string
		UserID      string
	}{
		OriginalURL: value,
		UserID:      userID,
	}
	return key, true, nil
}

func (s *MemoryStorage) GetByOriginalURL(originalURL string) (string, bool) {
	for key, val := range s.urls {
		if val.OriginalURL == originalURL {
			return key, true
		}
	}
	return "", false
}

func (s *MemoryStorage) getKeyByValueAndUser(value, userID string) (string, bool) {
	for key, val := range s.urls {
		if val.OriginalURL == value && val.UserID == userID {
			return key, true
		}
	}
	return "", false
}

func (s *MemoryStorage) BatchSave(items map[string]string, userID string) error {
	for key, value := range items {
		s.urls[key] = struct {
			OriginalURL string
			UserID      string
		}{
			OriginalURL: value,
			UserID:      userID,
		}
	}
	return nil
}

func (s *MemoryStorage) GetUserURLs(userID string) (map[string]string, error) {
	urls := make(map[string]string)
	for key, val := range s.urls {
		if val.UserID == userID {
			urls[key] = val.OriginalURL
		}
	}
	return urls, nil
}

func (s *MemoryStorage) Close() error {
	return nil
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MemoryStorage) BatchDelete(shortURLs []string, userID string) error {
	return nil
}
