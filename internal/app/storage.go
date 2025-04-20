package app

import (
	"encoding/json"
	"os"
	"sync"
)

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage struct {
	mu   sync.RWMutex
	urls map[string]string
	file *os.File
}

func NewStorage(filePath string) (*Storage, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		urls: make(map[string]string),
		file: file,
	}

	// Загружаем существующие URL из файла
	if err := storage.load(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Перемещаем указатель в начало файла
	if _, err := s.file.Seek(0, 0); err != nil {
		return err
	}

	decoder := json.NewDecoder(s.file)
	for decoder.More() {
		var record URLRecord
		if err := decoder.Decode(&record); err != nil {
			return err
		}
		s.urls[record.ShortURL] = record.OriginalURL
	}

	return nil
}

func (s *Storage) Save(shortURL, originalURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := URLRecord{
		UUID:        shortURL, // Используем shortURL как UUID
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	encoder := json.NewEncoder(s.file)
	if err := encoder.Encode(record); err != nil {
		return err
	}

	s.urls[shortURL] = originalURL
	return nil
}

func (s *Storage) Get(shortURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.urls[shortURL]
	return val, ok
}

func (s *Storage) Close() error {
	return s.file.Close()
}
