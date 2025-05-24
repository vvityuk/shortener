package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(dsn string) (*Storage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &Storage{db: db}, nil
}

func createTables(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_url VARCHAR(255) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			is_deleted BOOLEAN DEFAULT FALSE,
			UNIQUE(original_url, user_id)
		);
	`
	_, err := db.Exec(query)
	return err
}

func (s *Storage) Get(key string) (string, bool, bool) {
	var originalURL string
	var isDeleted bool
	err := s.db.QueryRow("SELECT original_url, is_deleted FROM urls WHERE short_url = $1", key).Scan(&originalURL, &isDeleted)
	if err == sql.ErrNoRows {
		return "", false, false
	}
	if err != nil {
		return "", false, false
	}
	return originalURL, isDeleted, true
}

func (s *Storage) Save(key, value string, userID string) (string, bool, error) {
	var shortURL string
	var isNew bool
	query := `
		WITH upsert AS (
			INSERT INTO urls (short_url, original_url, user_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (original_url, user_id) DO NOTHING
			RETURNING short_url, true as is_new
		)
		SELECT short_url, COALESCE(is_new, false) as is_new 
		FROM upsert
		UNION ALL
		SELECT short_url, false as is_new 
		FROM urls 
		WHERE original_url = $2 AND user_id = $3
		LIMIT 1
	`
	err := s.db.QueryRow(query, key, value, userID).Scan(&shortURL, &isNew)
	if err != nil {
		return "", false, err
	}
	return shortURL, isNew, nil
}

func (s *Storage) BatchSave(items map[string]string, userID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for key, value := range items {
		_, err = stmt.Exec(key, value, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Storage) GetByOriginalURL(originalURL string) (string, bool) {
	var shortURL string
	err := s.db.QueryRow("SELECT short_url FROM urls WHERE original_url = $1", originalURL).Scan(&shortURL)
	if err == sql.ErrNoRows {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return shortURL, true
}

func (s *Storage) GetUserURLs(userID string) (map[string]string, error) {
	query := `SELECT short_url, original_url FROM urls WHERE user_id = $1`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make(map[string]string)
	for rows.Next() {
		var shortURL, originalURL string
		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			return nil, err
		}
		urls[shortURL] = originalURL
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}

func (s *Storage) BatchDelete(shortURLs []string, userID string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	// Используем паттерн fanIn для эффективного обновления
	const batchSize = 100
	chunks := make([][]string, 0, (len(shortURLs)+batchSize-1)/batchSize)

	for i := 0; i < len(shortURLs); i += batchSize {
		end := i + batchSize
		if end > len(shortURLs) {
			end = len(shortURLs)
		}
		chunks = append(chunks, shortURLs[i:end])
	}

	errChan := make(chan error, len(chunks))

	for _, chunk := range chunks {
		go func(urls []string) {
			query := "UPDATE urls SET is_deleted = TRUE WHERE short_url = ANY($1) AND user_id = $2"
			_, err := s.db.Exec(query, pq.Array(urls), userID)
			errChan <- err
		}(chunk)
	}

	// Собираем ошибки
	for i := 0; i < len(chunks); i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}
