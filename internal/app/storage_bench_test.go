package app

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// Бенчмарки для MemoryStorage

func BenchmarkMemoryStorage_Save(b *testing.B) {
	storage := NewMemoryStorage()
	defer storage.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		_, _, _ = storage.Save(key, "https://example.com", "user1")
	}
}

func BenchmarkMemoryStorage_Get(b *testing.B) {
	storage := NewMemoryStorage()
	defer storage.Close()

	// Подготовка данных
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		_, _, _ = storage.Save(key, "https://example.com", "user1")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		_, _, _ = storage.Get(key)
	}
}

func BenchmarkMemoryStorage_BatchSave(b *testing.B) {
	storage := NewMemoryStorage()
	defer storage.Close()

	items := make(map[string]string)
	for i := 0; i < 100; i++ {
		items[fmt.Sprintf("key%d", i)] = "https://example.com"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.BatchSave(items, "user1")
	}
}

func BenchmarkMemoryStorage_GetUserURLs(b *testing.B) {
	storage := NewMemoryStorage()
	defer storage.Close()

	// Подготовка данных
	userID := "user1"
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		_, _, _ = storage.Save(key, "https://example.com", userID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetUserURLs(userID)
	}
}

// Бенчмарки для FileStorage

func BenchmarkFileStorage_Save(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "bench-storage-*.json")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	storage, err := NewStorage(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}
	defer storage.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		_, _, _ = storage.Save(key, "https://example.com", "user1")
	}
}

func BenchmarkFileStorage_Get(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "bench-storage-*.json")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	storage, err := NewStorage(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}
	defer storage.Close()

	// Подготовка данных
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		_, _, _ = storage.Save(key, "https://example.com", "user1")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		_, _, _ = storage.Get(key)
	}
}

func BenchmarkFileStorage_BatchSave(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "bench-storage-*.json")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	storage, err := NewStorage(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}
	defer storage.Close()

	items := make(map[string]string)
	for i := 0; i < 100; i++ {
		items[fmt.Sprintf("key%d", i)] = "https://example.com"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.BatchSave(items, "user1")
	}
}

func BenchmarkFileStorage_GetUserURLs(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "bench-storage-*.json")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	storage, err := NewStorage(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}
	defer storage.Close()

	// Подготовка данных
	userID := "user1"
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		_, _, _ = storage.Save(key, "https://example.com", userID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetUserURLs(userID)
	}
}

func BenchmarkFileStorage_Ping(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "bench-storage-*.json")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	storage, err := NewStorage(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}
	defer storage.Close()

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.Ping(ctx)
	}
}

