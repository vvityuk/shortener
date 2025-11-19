package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b)
	lrw.size += size
	return size, err
}

// LoggingMiddleware создает middleware для логирования HTTP-запросов.
// Логирует URI, метод, статус ответа, размер ответа и время выполнения запроса.
//
// Параметры:
//   - logger: логгер для записи информации о запросах
//
// Возвращает:
//   - func(http.Handler) http.Handler: функция middleware
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
