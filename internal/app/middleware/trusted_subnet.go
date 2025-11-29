// Package middleware предоставляет HTTP middleware для аутентификации, логирования и сжатия.
package middleware

import (
	"net"
	"net/http"
)

// TrustedSubnetMiddleware проверяет, что IP-адрес клиента из заголовка X-Real-IP
// входит в доверенную подсеть. Если trustedSubnet пустой или IP не входит
// в доверенную подсеть, возвращает статус 403 Forbidden.
func TrustedSubnetMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Если доверенная подсеть не указана, запрещаем доступ
			if trustedSubnet == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Получаем IP-адрес из заголовка X-Real-IP
			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Парсим IP-адрес клиента
			clientIP := net.ParseIP(realIP)
			if clientIP == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Парсим доверенную подсеть в формате CIDR
			_, subnet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Проверяем, входит ли IP клиента в доверенную подсеть
			if !subnet.Contains(clientIP) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// IP входит в доверенную подсеть, продолжаем обработку
			next.ServeHTTP(w, r)
		})
	}
}
