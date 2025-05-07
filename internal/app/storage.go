package app

import (
	"context"
	"encoding/json"
	"os"
)

type Storage interface {
	Get(key string) (string, bool)
	Save(key, value string) (string, error)
	GetByOriginalURL(originalURL string) (string, bool)
	BatchSave(items map[string]string) error
	Close() error
	Ping(ctx context.Context) error
}

type FileStorage struct {
	urls map[string]string
	file *os.File
}

func NewStorage(filePath string) (*FileStorage, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	storage := &FileStorage{
		urls: make(map[string]string),
		file: file,
	}

	if err := storage.load(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *FileStorage) Get(key string) (string, bool) {
	val, ok := s.urls[key]
	return val, ok
}

func (s *FileStorage) Save(key, value string) (string, error) {
	if existingKey, ok := s.getKeyByValue(value); ok {
		return existingKey, nil
	}
	s.urls[key] = value
	if err := s.save(); err != nil {
		return "", err
	}
	return key, nil
}

func (s *FileStorage) GetByOriginalURL(originalURL string) (string, bool) {
	for key, value := range s.urls {
		if value == originalURL {
			return key, true
		}
	}
	return "", false
}

func (s *FileStorage) getKeyByValue(value string) (string, bool) {
	for key, val := range s.urls {
		if val == value {
			return key, true
		}
	}
	return "", false
}

func (s *FileStorage) BatchSave(items map[string]string) error {
	for key, value := range items {
		s.urls[key] = value
	}
	return s.save()
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

type MemoryStorage struct {
	urls map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urls: make(map[string]string),
	}
}

func (s *MemoryStorage) Get(key string) (string, bool) {
	val, ok := s.urls[key]
	return val, ok
}

func (s *MemoryStorage) Save(key, value string) (string, error) {
	if existingKey, ok := s.getKeyByValue(value); ok {
		return existingKey, nil
	}
	s.urls[key] = value
	return key, nil
}

func (s *MemoryStorage) GetByOriginalURL(originalURL string) (string, bool) {
	for key, value := range s.urls {
		if value == originalURL {
			return key, true
		}
	}
	return "", false
}

func (s *MemoryStorage) getKeyByValue(value string) (string, bool) {
	for key, val := range s.urls {
		if val == value {
			return key, true
		}
	}
	return "", false
}

func (s *MemoryStorage) BatchSave(items map[string]string) error {
	for key, value := range items {
		s.urls[key] = value
	}
	return nil
}

func (s *MemoryStorage) Close() error {
	return nil
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}
