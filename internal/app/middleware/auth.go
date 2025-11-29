// Package middleware предоставляет HTTP middleware для аутентификации, логирования и сжатия.
package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Константы системы.
const (
	// ChiookieName имя cookie для хранения идентификатора пользователя
	ChiookieName = "user_id"
	// jwtIssuer идентификатор издателя JWT токенов
	jwtIssuer = "shortener-service"
	// jwtTTL срок действия JWT токена (24 часа)
	jwtTTL = 24 * time.Hour
)

// Ошибки валидации JWT токенов.
var (
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenInvalid      = errors.New("token invalid")
	ErrTokenMalformed    = errors.New("token malformed")
	ErrJWTNotInitialized = errors.New("JWT manager not initialized")
)

// Claims представляет структуру JWT токена.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// JWTManager управляет созданием и валидацией JWT токенов.
type JWTManager struct {
	secretKey []byte
	issuer    string
	ttl       time.Duration
}

var (
	jwtManager *JWTManager
	jwtOnce    sync.Once
	jwtMutex   sync.RWMutex
)

// InitJWTManager инициализирует глобальный JWT Manager с указанным секретным ключом.
// Должен быть вызван один раз при старте приложения.
func InitJWTManager(secretKey string) error {
	if secretKey == "" {
		return fmt.Errorf("JWT secret key cannot be empty")
	}
	if len(secretKey) < 32 {
		return fmt.Errorf("JWT secret key must be at least 32 characters long")
	}

	jwtOnce.Do(func() {
		jwtMutex.Lock()
		defer jwtMutex.Unlock()
		jwtManager = &JWTManager{
			secretKey: []byte(secretKey),
			issuer:    jwtIssuer,
			ttl:       jwtTTL,
		}
	})

	return nil
}

// getJWTManager возвращает глобальный JWT Manager.
func getJWTManager() (*JWTManager, error) {
	jwtMutex.RLock()
	defer jwtMutex.RUnlock()
	if jwtManager == nil {
		return nil, ErrJWTNotInitialized
	}
	return jwtManager, nil
}

// AuthMiddleware обеспечивает аутентификацию пользователей через cookie.
// Если у пользователя нет валидной cookie, создается новая и устанавливается в ответ.
// Идентификатор пользователя доступен через GetUserID.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(ChiookieName)
		if err != nil || !isValidCookie(cookie.Value) {
			userID := generateUserID()
			http.SetCookie(w, &http.Cookie{
				Name:     ChiookieName,
				Value:    userID,
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			})
			r.AddCookie(&http.Cookie{Name: ChiookieName, Value: userID})
		}
		next.ServeHTTP(w, r)
	})
}

// generateUserID генерирует уникальный идентификатор пользователя.
func generateUserID() string {
	// Используем секретный ключ из JWT Manager, если он инициализирован
	// Иначе используем временный ключ (для обратной совместимости при генерации userID)
	secret := "23ev43VRE35srv45"
	if mgr, err := getJWTManager(); err == nil {
		secret = string(mgr.secretKey)
	}
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(time.Now().String()))
	return hex.EncodeToString(h.Sum(nil))
}

func isValidCookie(value string) bool {
	return len(value) > 0
}

// GetUserID извлекает идентификатор пользователя из cookie запроса.
// Возвращает идентификатор пользователя или пустую строку, если cookie отсутствует.
func GetUserID(r *http.Request) string {
	cookie, err := r.Cookie(ChiookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// GenerateToken генерирует новый JWT токен авторизации и возвращает userID и токен.
// Используется для аутентификации в gRPC.
func GenerateToken() (string, string, error) {
	mgr, err := getJWTManager()
	if err != nil {
		return "", "", err
	}

	userID := generateUserID()
	now := time.Now()

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(mgr.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    mgr.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(mgr.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err)
	}

	return userID, tokenString, nil
}

// ValidateToken проверяет валидность JWT токена и возвращает userID.
// Если токен невалидный, истек или поврежден, возвращает соответствующую ошибку.
func ValidateToken(tokenString string) (string, error) {
	mgr, err := getJWTManager()
	if err != nil {
		return "", err
	}

	// Парсим токен
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return mgr.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrTokenExpired
		}
		return "", fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	// Проверяем валидность токена и извлекаем claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", ErrTokenInvalid
	}

	// Проверяем issuer
	if claims.Issuer != mgr.issuer {
		return "", ErrTokenInvalid
	}

	return claims.UserID, nil
}
