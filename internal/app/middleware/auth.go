package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

// Кнстанты системы.
const (
	// ChiookieName имя cookie для хранения идентификатора пользователя
	ChiookieName = "user_id"
	secretKey    = "23ev43VRE35srv45" // Нужно будет перенести в конфиг
)

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

func generateUserID() string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(time.Now().String()))
	return hex.EncodeToString(h.Sum(nil))
}

func isValidCookie(value string) bool {
	return len(value) > 0
}

// GetUserID извлекает идентификатор пользователя из cookie запроса.
//
// Параметры:
//   - r: HTTP-запрос
//
// Возвращает:
//   - string: идентификатор пользователя или пустую строку, если cookie отсутствует
func GetUserID(r *http.Request) string {
	cookie, err := r.Cookie(ChiookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
