package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
)

// Storage реализует хранилище на основе PostgreSQL.
// Обеспечивает персистентное хранение данных с поддержкой транзакций.
type Storage struct {
	db *sql.DB
}

// New создает новое PostgreSQL хранилище и инициализирует необходимые таблицы.
// Принимает строку подключения к PostgreSQL (например, "postgres://user:pass@localhost/dbname").
// Возвращает новый экземпляр PostgreSQL хранилища или ошибку при подключении к базе данных или создании таблиц.
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

// Get получает оригинальный URL по короткому коду из базы данных.
// Возвращает оригинальный URL, флаг удаления (true если URL был удален)
// и флаг успешного получения (true если URL найден).
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

// Save сохраняет короткий URL для указанного оригинального URL в базе данных.
// Использует UPSERT для предотвращения дублирования: если URL уже существует для данного пользователя,
// возвращает существующий короткий код без создания новой записи.
// Возвращает короткий код URL, флаг создания нового URL (true если создан новый,
// false если уже существовал) и ошибку при сохранении в базу данных.
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

// BatchSave сохраняет несколько коротких URL за один запрос в рамках транзакции.
// Использует подготовленные запросы для повышения производительности.
// При ошибке выполнения транзакция откатывается.
// Принимает карту соответствий short_code -> original_url и идентификатор пользователя.
// Возвращает ошибку при сохранении в базу данных или откате транзакции.
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

// Close закрывает соединение с базой данных PostgreSQL и освобождает ресурсы.
// Возвращает ошибку при закрытии соединения с базой данных.
func (s *Storage) Close() error {
	return s.db.Close()
}

// Ping проверяет доступность базы данных PostgreSQL.
// Используется для health-check эндпоинтов.
// Возвращает ошибку при проверке доступности базы данных.
func (s *Storage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// GetByOriginalURL находит короткий код по оригинальному URL в базе данных.
// Возвращает короткий код URL и флаг успешного поиска (true если URL найден).
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

// GetUserURLs возвращает все короткие URL, созданные указанным пользователем.
// Выполняет SQL-запрос для получения всех записей пользователя из базы данных.
// Возвращает карту соответствий short_code -> original_url и ошибку при выполнении запроса к базе данных.
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

// BatchDelete помечает указанные короткие URL как удаленные в базе данных.
// Использует паттерн fanIn для параллельной обработки больших объемов данных:
// разбивает массив URL на чанки по 100 элементов и обрабатывает их параллельно в горутинах.
// Удаляются только URL указанного пользователя.
// Возвращает ошибку при выполнении операции удаления.
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
