package app

import (
	"time"

	"github.com/vvityuk/shortener/internal/config"
	"golang.org/x/exp/rand"
)

type Service struct {
	urls   map[string]string
	config *config.Config
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		urls:   make(map[string]string),
		config: cfg,
	}
}

func (s *Service) GetURL(shortCode string) (string, bool) {
	val, ok := s.urls["/"+shortCode]
	return val, ok
}

func (s *Service) CreateURL(longURL string) string {
	ok := true
	shortURL := ""
	for ok {
		shortURL = s.randStr(4)
		_, ok = s.urls["/"+shortURL]
	}
	s.urls["/"+shortURL] = longURL
	return shortURL
}

func (s *Service) randStr(n int) string {
	rnd := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))

	letters := []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rnd.Intn(len(letters))]
	}
	return string(b)
}
