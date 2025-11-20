// Package middleware предоставляет HTTP middleware для аутентификации, логирования и сжатия.
package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// loggingResponseWriter оборачивает http.ResponseWriter для отслеживания статуса ответа и размера данных.
// Используется в LoggingMiddleware для сбора метрик о HTTP-ответах (статус код и размер ответа).
type loggingResponseWriter struct {
	// ResponseWriter оригинальный HTTP ResponseWriter для делегирования операций записи
	http.ResponseWriter
	// statusCode HTTP-статус код ответа
	statusCode int
	// size размер записанных данных в байтах
	size int
}

// WriteHeader устанавливает HTTP-статус код ответа и сохраняет его для логирования.
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Write записывает данные в обёрнутый ResponseWriter и отслеживает размер записанных данных.
// Накопленный размер используется для логирования метрик запроса.
// Возвращает количество записанных байт и ошибку при записи данных.
func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b)
	lrw.size += size
	return size, err
}

// LoggingMiddleware создает middleware для логирования HTTP-запросов.
// Логирует URI, метод, статус ответа, размер ответа и время выполнения запроса.
// Принимает логгер для записи информации о запросах и возвращает функцию middleware.
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(lrw, r)

			duration := time.Since(start)

			logger.Info("request completed",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Int("status", lrw.statusCode),
				zap.Int("size", lrw.size),
				zap.Duration("duration", duration),
			)
		})
	}
}
