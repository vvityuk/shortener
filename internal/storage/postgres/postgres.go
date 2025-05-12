package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
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
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(original_url)
		);
	`
	_, err := db.Exec(query)
	return err
}

func (s *Storage) Get(key string) (string, bool) {
	var originalURL string
	err := s.db.QueryRow("SELECT original_url FROM urls WHERE short_url = $1", key).Scan(&originalURL)
	if err == sql.ErrNoRows {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return originalURL, true
}

func (s *Storage) Save(key, value string) (string, bool, error) {
	var shortURL string
	var isNew bool
	query := `
		WITH upsert AS (
			INSERT INTO urls (short_url, original_url)
			VALUES ($1, $2)
			ON CONFLICT (original_url) DO NOTHING
			RETURNING short_url, true as is_new
		)
		SELECT short_url, COALESCE(is_new, false) as is_new 
		FROM upsert
		UNION ALL
		SELECT short_url, false as is_new 
		FROM urls 
		WHERE original_url = $2
		LIMIT 1
	`
	err := s.db.QueryRow(query, key, value).Scan(&shortURL, &isNew)
	if err != nil {
		return "", false, err
	}
	return shortURL, isNew, nil
}

func (s *Storage) BatchSave(items map[string]string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO urls (short_url, original_url) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for key, value := range items {
		_, err = stmt.Exec(key, value)
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
