package middleware

import (
	"testing"
)

// Бенчмарк генерации UserID
func BenchmarkGenerateUserID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateUserID()
	}
}

// Бенчмарк проверки валидности cookie
func BenchmarkIsValidCookie(b *testing.B) {
	validCookie := "valid_cookie_value_12345"
	invalidCookie := ""

	b.Run("ValidCookie", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = isValidCookie(validCookie)
		}
	})

	b.Run("InvalidCookie", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = isValidCookie(invalidCookie)
		}
	})
}
