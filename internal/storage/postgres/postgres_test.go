package postgres

import (
	"context"
	"os"
	"testing"
)

func TestPostgresStorage(t *testing.T) {
	// Получаем DSN из переменной окружения или используем тестовую БД
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/shortener_test?sslmode=disable"
	}

	storage, err := New(dsn)
	if err != nil {
		t.Skipf("Failed to connect to test database: %v", err)
	}
	defer storage.Close()

	// Очищаем таблицу перед тестами
	_, err = storage.db.Exec("TRUNCATE TABLE urls")
	if err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}

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
		url, ok := storage.Get(shortURL)
		if !ok {
			t.Error("Failed to get URL")
		}
		if url != originalURL {
			t.Errorf("Expected URL %s, got %s", originalURL, url)
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
			url, ok := storage.Get(key)
			if !ok {
				t.Errorf("Failed to get URL for key %s", key)
			}
			if url != value {
				t.Errorf("Expected URL %s for key %s, got %s", value, key, url)
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
		expectedCount := 3 // abc123, key1, key2
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
