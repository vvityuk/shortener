// Package app предоставляет бизнес-логику и HTTP-обработчики для сервиса сокращения URL.
package app

import (
	"context"
	"encoding/json"
	"os"
)

// Storage определяет интерфейс для работы с хранилищем коротких URL.
// Реализации могут использовать память, файлы или базу данных.
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

// FileStorage реализует хранилище на основе JSON-файла.
// Данные сохраняются в файл и загружаются при инициализации.
type FileStorage struct {
	urls map[string]struct {
		OriginalURL string
		UserID      string
	}
	file *os.File
}

// NewStorage создает новое файловое хранилище.
// Принимает путь к JSON-файлу для хранения данных.
// Возвращает новый экземпляр файлового хранилища или ошибку при создании или открытии файла.
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

// Get получает оригинальный URL по короткому коду.
// Возвращает оригинальный URL, флаг удаления (всегда false для FileStorage)
// и флаг успешного получения (true если URL найден).
func (s *FileStorage) Get(key string) (string, bool, bool) {
	val, ok := s.urls[key]
	return val.OriginalURL, false, ok
}

// Save сохраняет короткий URL для указанного оригинального URL.
// Если URL уже существует для данного пользователя, возвращает существующий короткий код.
// Данные автоматически сохраняются в файл.
// Возвращает короткий код URL, флаг создания нового URL (true если создан новый,
// false если уже существовал) и ошибку при сохранении в файл.
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

// GetByOriginalURL находит короткий код по оригинальному URL.
// Возвращает короткий код URL и флаг успешного поиска (true если URL найден).
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

// BatchSave сохраняет несколько коротких URL за один запрос.
// Данные автоматически сохраняются в файл после добавления всех элементов.
// Принимает карту соответствий short_code -> original_url и идентификатор пользователя.
// Возвращает ошибку при сохранении в файл.
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

// GetUserURLs возвращает все короткие URL, созданные указанным пользователем.
// Возвращает карту соответствий short_code -> original_url.
// Для FileStorage ошибка всегда nil.
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

// Close закрывает файл хранилища и освобождает ресурсы.
// Возвращает ошибку при закрытии файла.
func (s *FileStorage) Close() error {
	return s.file.Close()
}

// Ping проверяет доступность файлового хранилища.
// Для FileStorage всегда возвращает nil, так как файл уже открыт.
func (s *FileStorage) Ping(ctx context.Context) error {
	return nil
}

// BatchDelete помечает указанные короткие URL как удаленные.
// Для FileStorage реализация отсутствует, всегда возвращает nil.
func (s *FileStorage) BatchDelete(shortURLs []string, userID string) error {
	return nil
}

// MemoryStorage реализует хранилище в оперативной памяти.
// Данные не сохраняются между перезапусками приложения.
type MemoryStorage struct {
	urls map[string]struct {
		OriginalURL string
		UserID      string
	}
}

// NewMemoryStorage создает новое хранилище в памяти.
// Возвращает новый экземпляр хранилища в памяти.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urls: make(map[string]struct {
			OriginalURL string
			UserID      string
		}),
	}
}

// Get получает оригинальный URL по короткому коду.
// Возвращает оригинальный URL, флаг удаления (всегда false для MemoryStorage)
// и флаг успешного получения (true если URL найден).
func (s *MemoryStorage) Get(key string) (string, bool, bool) {
	val, ok := s.urls[key]
	return val.OriginalURL, false, ok
}

// Save сохраняет короткий URL для указанного оригинального URL.
// Если URL уже существует для данного пользователя, возвращает существующий короткий код.
// Возвращает короткий код URL, флаг создания нового URL (true если создан новый,
// false если уже существовал). Для MemoryStorage ошибка всегда nil.
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

// GetByOriginalURL находит короткий код по оригинальному URL.
// Возвращает короткий код URL и флаг успешного поиска (true если URL найден).
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

// BatchSave сохраняет несколько коротких URL за один запрос.
// Принимает карту соответствий short_code -> original_url и идентификатор пользователя.
// Для MemoryStorage ошибка всегда nil.
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

// GetUserURLs возвращает все короткие URL, созданные указанным пользователем.
// Возвращает карту соответствий short_code -> original_url.
// Для MemoryStorage ошибка всегда nil.
func (s *MemoryStorage) GetUserURLs(userID string) (map[string]string, error) {
	urls := make(map[string]string)
	for key, val := range s.urls {
		if val.UserID == userID {
			urls[key] = val.OriginalURL
		}
	}
	return urls, nil
}

// Close освобождает ресурсы хранилища.
// Для MemoryStorage не выполняет никаких действий, всегда возвращает nil.
func (s *MemoryStorage) Close() error {
	return nil
}

// Ping проверяет доступность хранилища в памяти.
// Для MemoryStorage всегда возвращает nil, так как хранилище всегда доступно.
func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

// BatchDelete помечает указанные короткие URL как удаленные.
// Для MemoryStorage реализация отсутствует, всегда возвращает nil.
func (s *MemoryStorage) BatchDelete(shortURLs []string, userID string) error {
	return nil
}
