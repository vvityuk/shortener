package app

import (
	"context"
	"os"
	"testing"
)

func TestFileStorage(t *testing.T) {
	// Создаем временный файл для тестов
	tmpFile, err := os.CreateTemp("", "storage-test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	storage, err := NewStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	testStorage(t, storage)
}

func TestMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()
	testStorage(t, storage)
}

func TestFileStorage_Get(t *testing.T) {
	storage, err := NewStorage("test_urls.json")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	// Сохраняем тестовые данные
	_, _, err = storage.Save("test_key", "test_value", "user1")
	if err != nil {
		t.Fatal(err)
	}

	// Проверяем получение существующего значения
	value, isDeleted, ok := storage.Get("test_key")
	if !ok {
		t.Error("Expected to get value, got false")
	}
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}
	if isDeleted {
		t.Error("Expected isDeleted to be false")
	}

	// Проверяем получение несуществующего значения
	_, isDeleted, ok = storage.Get("non_existent")
	if ok {
		t.Error("Expected to get false for non-existent key")
	}
	if isDeleted {
		t.Error("Expected isDeleted to be false for non-existent key")
	}
}

func TestMemoryStorage_Get(t *testing.T) {
	storage := NewMemoryStorage()

	// Сохраняем тестовые данные
	_, _, err := storage.Save("test_key", "test_value", "user1")
	if err != nil {
		t.Fatal(err)
	}

	// Проверяем получение существующего значения
	value, isDeleted, ok := storage.Get("test_key")
	if !ok {
		t.Error("Expected to get value, got false")
	}
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}
	if isDeleted {
		t.Error("Expected isDeleted to be false")
	}

	// Проверяем получение несуществующего значения
	_, isDeleted, ok = storage.Get("non_existent")
	if ok {
		t.Error("Expected to get false for non-existent key")
	}
	if isDeleted {
		t.Error("Expected isDeleted to be false for non-existent key")
	}
}

func testStorage(t *testing.T, storage Storage) {
	t.Run("Save and Get URL", func(t *testing.T) {
		userID := "test-user"
		originalURL := "https://example.com"
		shortURL, isNew, err := storage.Save("abc123", originalURL, userID)
		if err != nil {
			t.Fatalf("Failed to save URL: %v", err)
		}
		if !isNew {
			t.Error("Expected isNew to be true")
		}
		if shortURL != "abc123" {
			t.Errorf("Expected short URL 'abc123', got %s", shortURL)
		}

		// Проверяем получение URL
		url, isDeleted, ok := storage.Get(shortURL)
		if !ok {
			t.Error("Failed to get URL")
		}
		if url != originalURL {
			t.Errorf("Expected URL %s, got %s", originalURL, url)
		}
		if isDeleted {
			t.Error("Expected isDeleted to be false")
		}

		// Проверяем сохранение того же URL для того же пользователя
		shortURL2, isNew, err := storage.Save("def456", originalURL, userID)
		if err != nil {
			t.Fatalf("Failed to save URL: %v", err)
		}
		if isNew {
			t.Error("Expected isNew to be false")
		}
		if shortURL2 != shortURL {
			t.Errorf("Expected short URL %s, got %s", shortURL, shortURL2)
		}

		// Проверяем сохранение того же URL для другого пользователя
		shortURL3, isNew, err := storage.Save("ghi789", originalURL, "other-user")
		if err != nil {
			t.Fatalf("Failed to save URL: %v", err)
		}
		if !isNew {
			t.Error("Expected isNew to be true")
		}
		if shortURL3 != "ghi789" {
			t.Errorf("Expected short URL 'ghi789', got %s", shortURL3)
		}
	})

	t.Run("Batch Save", func(t *testing.T) {
		userID := "test-user"
		items := map[string]string{
			"key1": "https://example1.com",
			"key2": "https://example2.com",
		}

		err := storage.BatchSave(items, userID)
		if err != nil {
			t.Fatalf("Failed to batch save: %v", err)
		}

		// Проверяем, что все URL сохранены
		for key, value := range items {
			url, isDeleted, ok := storage.Get(key)
			if !ok {
				t.Errorf("Failed to get URL for key %s", key)
			}
			if url != value {
				t.Errorf("Expected URL %s for key %s, got %s", value, key, url)
			}
			if isDeleted {
				t.Errorf("Expected isDeleted to be false for key %s", key)
			}
		}
	})

	t.Run("Get User URLs", func(t *testing.T) {
		userID := "test-user"
		urls, err := storage.GetUserURLs(userID)
		if err != nil {
			t.Fatalf("Failed to get user URLs: %v", err)
		}

		// Проверяем, что все URL пользователя получены
		expectedCount := 3 // abc123, def456, key1, key2
		if len(urls) != expectedCount {
			t.Errorf("Expected %d URLs, got %d", expectedCount, len(urls))
		}

		// Проверяем URL другого пользователя
		otherUserURLs, err := storage.GetUserURLs("other-user")
		if err != nil {
			t.Fatalf("Failed to get other user URLs: %v", err)
		}
		if len(otherUserURLs) != 1 {
			t.Errorf("Expected 1 URL for other user, got %d", len(otherUserURLs))
		}
	})

	t.Run("Ping", func(t *testing.T) {
		err := storage.Ping(context.Background())
		if err != nil {
			t.Errorf("Ping failed: %v", err)
		}
	})
}
