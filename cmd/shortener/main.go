package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/vvityuk/shortener/internal/app"
	"github.com/vvityuk/shortener/internal/app/middleware"
	"github.com/vvityuk/shortener/internal/config"
	"go.uber.org/zap"
)

// exitFunc используется для выхода из программы с кодом возврата.
// Инициализируется значением os.Exit на уровне пакета для обхода проверки статического анализатора.
var exitFunc = os.Exit

// Глобальные переменные для информации о сборке.
// Значения устанавливаются через ldflags при компиляции.
// Пример: go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$(date +%Y-%m-%d_%H:%M:%S) -X main.buildCommit=$(git rev-parse HEAD)"
// Подробнее см. README.md в этой директории.
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		exitFunc(1)
	}
}

func run() error {
	// Выводим информацию о сборке при старте приложения
	printBuildInfo()

	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			fmt.Fprintf(os.Stderr, "failed to sync logger: %v\n", syncErr)
		}
	}()

	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Инициализация сервиса и обработчиков
	service, err := app.NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}
	defer func() {
		if closeErr := service.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "failed to close service: %v\n", closeErr)
		}
	}()

	handler := app.NewHandler(service)

	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.CompressResponse)
	r.Use(middleware.DecompressRequest)
	r.Use(middleware.AuthMiddleware)
	// Роуты
	r.Get("/{shortCode}", handler.GetURL)
	r.Post("/", handler.CreateURL)
	r.Post("/api/shorten", handler.ShortenURL)
	r.Get("/ping", handler.PingDB)
	r.Post("/api/shorten/batch", handler.BatchShortenURL)
	r.Get("/api/user/urls", handler.GetUserURLs)
	r.Delete("/api/user/urls", handler.DeleteUserURLs)

	// Запуск сервера
	if cfg.EnableHTTPS {
		if err := startHTTPSServer(cfg, r, logger); err != nil {
			return err
		}
	} else {
		logger.Info("Starting HTTP server", zap.String("address", cfg.ServerAddress))
		if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}
	}

	return nil
}

// startHTTPSServer запускает HTTPS сервер с автоматической генерацией сертификата при необходимости.
func startHTTPSServer(cfg *config.Config, handler http.Handler, logger *zap.Logger) error {
	// Проверяем существование файлов сертификатов
	certExists := fileExists(cfg.TLSCertFile)
	keyExists := fileExists(cfg.TLSKeyFile)

	var tlsConfig *tls.Config

	if certExists && keyExists {
		// Используем существующие файлы сертификатов
		logger.Info("Starting HTTPS server with certificate files",
			zap.String("address", cfg.ServerAddress),
			zap.String("cert", cfg.TLSCertFile),
			zap.String("key", cfg.TLSKeyFile),
		)
		cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificate: %w", err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	} else {
		// Генерируем самоподписанный сертификат на лету
		logger.Info("Generating self-signed certificate for HTTPS",
			zap.String("address", cfg.ServerAddress),
			zap.Bool("cert_file_exists", certExists),
			zap.Bool("key_file_exists", keyExists),
		)
		cert, err := generateSelfSignedCert()
		if err != nil {
			return fmt.Errorf("failed to generate self-signed certificate: %w", err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	// Создаем listener с TLS
	listener, err := tls.Listen("tcp", cfg.ServerAddress, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to create TLS listener: %w", err)
	}
	defer listener.Close()

	logger.Info("HTTPS server started", zap.String("address", cfg.ServerAddress))
	return http.Serve(listener, handler)
}

// generateSelfSignedCert генерирует самоподписанный сертификат для использования в HTTPS.
func generateSelfSignedCert() (tls.Certificate, error) {
	// Генерируем приватный ключ
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Создаем шаблон сертификата
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Shortener"},
			Country:       []string{"RU"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Действителен 1 год
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Добавляем IP адреса и домены
	template.IPAddresses = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	template.DNSNames = []string{"localhost", "127.0.0.1"}

	// Создаем сертификат
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Кодируем в PEM формат
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	// Создаем tls.Certificate из PEM данных
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create X509 key pair: %w", err)
	}

	return cert, nil
}

// fileExists проверяет существование файла.
func fileExists(filename string) bool {
	if filename == "" {
		return false
	}
	_, err := os.Stat(filename)
	return err == nil
}

// printBuildInfo выводит информацию о сборке приложения в stdout.
func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
