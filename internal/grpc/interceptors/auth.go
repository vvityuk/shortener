// Package interceptors предоставляет gRPC interceptors для обработки запросов.
package interceptors

import (
	"context"

	"github.com/vvityuk/shortener/internal/app/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// contextKey тип для ключей контекста
type contextKey string

const (
	// userIDKey ключ для хранения ID пользователя в контексте
	userIDKey contextKey = "userID"
)

// AuthInterceptor извлекает токен авторизации из metadata и добавляет userID в контекст.
// Если токен отсутствует, генерирует новый.
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Извлекаем metadata из контекста
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata not found")
		}

		// Извлекаем токен из заголовка authorization
		var userID string
		authHeaders := md.Get("authorization")
		if len(authHeaders) > 0 {
			token := authHeaders[0]
			// Проверяем токен и извлекаем userID
			var err error
			userID, err = middleware.ValidateToken(token)
			if err != nil {
				// Если токен невалидный, генерируем новый
				userID, token, err = middleware.GenerateToken()
				if err != nil {
					return nil, status.Error(codes.Internal, "failed to generate token")
				}
				// Добавляем новый токен в исходящий metadata
				header := metadata.Pairs("authorization", token)
				if err := grpc.SetHeader(ctx, header); err != nil {
					return nil, status.Error(codes.Internal, "failed to set header")
				}
			}
		} else {
			// Если токен не передан, генерируем новый
			var token string
			var err error
			userID, token, err = middleware.GenerateToken()
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to generate token")
			}
			// Добавляем новый токен в исходящий metadata
			header := metadata.Pairs("authorization", token)
			if err := grpc.SetHeader(ctx, header); err != nil {
				return nil, status.Error(codes.Internal, "failed to set header")
			}
		}

		// Добавляем userID в контекст
		ctx = context.WithValue(ctx, userIDKey, userID)

		// Вызываем handler
		return handler(ctx, req)
	}
}

// GetUserID извлекает userID из контекста
func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUserIDKey возвращает ключ для userID в контексте (для тестов)
func GetUserIDKey() contextKey {
	return userIDKey
}

