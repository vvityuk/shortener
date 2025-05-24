package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

const (
	cookieName = "user_id"
	secretKey  = "23ev43VRE35srv45" // Нужно будет перенести в конфиг
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil || !isValidCookie(cookie.Value) {
			userID := generateUserID()
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    userID,
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			})
			r.AddCookie(&http.Cookie{Name: cookieName, Value: userID})
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

func GetUserID(r *http.Request) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
