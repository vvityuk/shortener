package app

import (
	"context"
	"encoding/json"
	"os"
)

type Storage interface {
	Get(key string) (string, bool)
	Save(key, value string) error
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

func (s *FileStorage) Save(key, value string) error {
	s.urls[key] = value
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
