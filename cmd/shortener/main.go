package main

import (
	"context"
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
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/vvityuk/shortener/internal/app"
	"github.com/vvityuk/shortener/internal/app/middleware"
	"github.com/vvityuk/shortener/internal/config"
	grpcserver "github.com/vvityuk/shortener/internal/grpc"
	"github.com/vvityuk/shortener/internal/grpc/interceptors"
	pb "github.com/vvityuk/shortener/pkg/grpc/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Константы таймаутов для HTTP сервера.
const (
	shutdownTimeout = 30 * time.Second
	readTimeout     = 15 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 60 * time.Second
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

	// Internal эндпоинты с проверкой доверенной подсети
	r.With(middleware.TrustedSubnetMiddleware(cfg.TrustedSubnet)).Get("/api/internal/stats", handler.InternalStats)

	// Создаем контекст для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	// Запуск HTTP сервера с поддержкой graceful shutdown
	var server *http.Server
	if cfg.EnableHTTPS {
		server, err = createHTTPSServer(cfg, r, logger)
		if err != nil {
			return fmt.Errorf("failed to create HTTPS server: %w", err)
		}
	} else {
		server = createHTTPServer(cfg, r)
	}

	// Создаем и запускаем gRPC сервер
	grpcServer, err := createGRPCServer(cfg, service, logger)
	if err != nil {
		return fmt.Errorf("failed to create gRPC server: %w", err)
	}

	return runServersWithGracefulShutdown(ctx, server, grpcServer, logger, service, cfg.EnableHTTPS, cfg.GRPCAddress)
}

// createHTTPServer создает HTTP сервер.
func createHTTPServer(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}

// createHTTPSServer создает HTTPS сервер с автоматической генерацией сертификата при необходимости.
func createHTTPSServer(cfg *config.Config, handler http.Handler, logger *zap.Logger) (*http.Server, error) {
	var cert tls.Certificate
	var err error

	// Проверяем существование файлов сертификата
	certFileExists := cfg.TLSCertFile != "" && fileExists(cfg.TLSCertFile)
	keyFileExists := cfg.TLSKeyFile != "" && fileExists(cfg.TLSKeyFile)

	if certFileExists && keyFileExists {
		// Пытаемся загрузить сертификат из файлов
		cert, err = tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
		}
		logger.Info("Starting HTTPS server with certificate files",
			zap.String("address", cfg.ServerAddress),
			zap.String("cert", cfg.TLSCertFile),
			zap.String("key", cfg.TLSKeyFile),
		)
	} else {
		// Если файлы не существуют, генерируем самоподписанный сертификат
		logger.Info("Certificate files not found, generating self-signed certificate for HTTPS",
			zap.String("address", cfg.ServerAddress),
		)
		cert, err = generateSelfSignedCert()
		if err != nil {
			return nil, fmt.Errorf("failed to generate self-signed certificate: %w", err)
		}
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Создаем сервер с TLS конфигурацией и таймаутами
	return &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      handler,
		TLSConfig:    tlsConfig,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}, nil
}

// fileExists проверяет существование файла.
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// createGRPCServer создает и настраивает gRPC сервер с interceptors.
func createGRPCServer(cfg *config.Config, service *app.Service, logger *zap.Logger) (*grpc.Server, error) {
	// Создаем gRPC сервер с interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.AuthInterceptor()),
	)

	// Регистрируем сервис
	grpcService := grpcserver.NewServer(service)
	pb.RegisterShortenerServiceServer(grpcServer, grpcService)

	logger.Info("gRPC server configured", zap.String("address", cfg.GRPCAddress))
	return grpcServer, nil
}

// runServersWithGracefulShutdown запускает HTTP и gRPC серверы с обработкой сигналов завершения через контекст.
// Выполняет graceful shutdown: сначала завершает HTTP и gRPC серверы, затем закрывает сервис и его зависимости.
func runServersWithGracefulShutdown(ctx context.Context, httpServer *http.Server, grpcServer *grpc.Server, logger *zap.Logger, service *app.Service, isTLS bool, grpcAddress string) error {
	// Канал для ошибок серверов
	serverErrors := make(chan error, 2)

	// Создаем listener для gRPC сервера заранее, чтобы проверить ошибки до запуска
	grpcListener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		return fmt.Errorf("failed to create gRPC listener: %w", err)
	}

	// Запускаем HTTP сервер в горутине
	go func() {
		var err error
		if isTLS {
			logger.Info("Starting HTTPS server", zap.String("address", httpServer.Addr))
			// Используем пустые строки для cert и key, так как они уже в TLSConfig
			err = httpServer.ListenAndServeTLS("", "")
		} else {
			logger.Info("Starting HTTP server", zap.String("address", httpServer.Addr))
			err = httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Запускаем gRPC сервер в горутине
	go func() {
		logger.Info("Starting gRPC server", zap.String("address", grpcAddress))
		if err := grpcServer.Serve(grpcListener); err != nil {
			serverErrors <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Ожидаем сигнал завершения через контекст или ошибку запуска
	select {
	case <-ctx.Done():
		logger.Info("Received shutdown signal", zap.String("signal", ctx.Err().Error()))
		return gracefulShutdown(ctx, httpServer, grpcServer, grpcListener, logger, service)
	case err := <-serverErrors:
		if err != nil {
			// Если сервер не запустился, закрываем сервис перед возвратом ошибки
			if closeErr := service.Close(); closeErr != nil {
				logger.Error("Error closing service after server start failure", zap.Error(closeErr))
			}
			// Закрываем listener при ошибке
			if closeErr := grpcListener.Close(); closeErr != nil {
				logger.Error("Error closing gRPC listener after server start failure", zap.Error(closeErr))
			}
			return fmt.Errorf("failed to start server: %w", err)
		}
	}

	return nil
}

// gracefulShutdown выполняет плавное завершение работы всех компонентов.
// Порядок завершения: сначала HTTP и gRPC серверы, затем сервис (который закроет storage и другие зависимости).
func gracefulShutdown(ctx context.Context, httpServer *http.Server, grpcServer *grpc.Server, grpcListener net.Listener, logger *zap.Logger, service *app.Service) error {
	// Создаем контекст с таймаутом для завершения
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	logger.Info("Shutting down servers gracefully...")

	// Останавливаем gRPC сервер с таймаутом через горутину
	grpcStopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcStopped)
	}()

	// Инициируем graceful shutdown HTTP сервера (завершаем обработку активных запросов)
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error during HTTP server shutdown", zap.Error(err))
		// Продолжаем остановку других компонентов даже при ошибке HTTP сервера
	} else {
		logger.Info("HTTP server stopped")
	}

	// Ждем остановки gRPC сервера с таймаутом
	select {
	case <-grpcStopped:
		logger.Info("gRPC server stopped")
	case <-shutdownCtx.Done():
		logger.Warn("gRPC server graceful stop timeout, forcing stop")
		grpcServer.Stop()
		if err := grpcListener.Close(); err != nil {
			logger.Error("Error closing gRPC listener", zap.Error(err))
		}
	}

	// Закрываем сервис (который закроет storage и другие зависимости)
	if err := service.Close(); err != nil {
		logger.Error("Error closing service", zap.Error(err))
		return fmt.Errorf("failed to close service: %w", err)
	}

	logger.Info("Graceful shutdown completed")
	return nil
}

// generateSelfSignedCert генерирует самоподписанный сертификат для использования в HTTPS.
func generateSelfSignedCert() (tls.Certificate, error) {
	// Генерируем приватный ключ
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Генерируем случайный серийный номер (требование безопасности)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Создаем шаблон сертификата
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Shortener"},
			Country:      []string{"RU"},
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

// printBuildInfo выводит информацию о сборке приложения в stdout.
func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
