// Package grpc предоставляет gRPC сервер для работы с сокращением URL.
package grpc

import (
	"context"

	"github.com/vvityuk/shortener/internal/app"
	"github.com/vvityuk/shortener/internal/grpc/interceptors"
	pb "github.com/vvityuk/shortener/pkg/grpc/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server реализует gRPC сервер для работы с сокращением URL.
type Server struct {
	pb.UnimplementedShortenerServiceServer
	service *app.Service
}

// NewServer создает новый экземпляр gRPC сервера.
func NewServer(service *app.Service) *Server {
	return &Server{
		service: service,
	}
}

// ShortenURL создает короткий URL для переданного длинного URL.
func (s *Server) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	// Проверяем наличие URL в запросе
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "URL is required")
	}

	// Получаем userID из контекста
	userID := interceptors.GetUserID(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// Создаем короткий URL
	shortURL, _, err := s.service.CreateURL(req.Url, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create short URL")
	}

	// Формируем полный URL
	fullURL := s.service.GetBaseURL() + "/" + shortURL

	return &pb.URLShortenResponse{
		Result: fullURL,
	}, nil
}

// ExpandURL возвращает оригинальный URL по короткому коду.
func (s *Server) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	// Проверяем наличие ID в запросе
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	// Получаем оригинальный URL
	originalURL, isDeleted, ok := s.service.GetURL(req.Id)
	if !ok {
		return nil, status.Error(codes.NotFound, "URL not found")
	}

	if isDeleted {
		return nil, status.Error(codes.FailedPrecondition, "URL has been deleted")
	}

	return &pb.URLExpandResponse{
		Result: originalURL,
	}, nil
}

// ListUserURLs возвращает список всех URL пользователя.
func (s *Server) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	// Получаем userID из контекста
	userID := interceptors.GetUserID(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// Получаем список URL пользователя
	urls, err := s.service.GetUserURLs(userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user URLs")
	}

	// Формируем ответ
	response := &pb.UserURLsResponse{
		Url: make([]*pb.URLData, 0, len(urls)),
	}

	for shortURL, originalURL := range urls {
		response.Url = append(response.Url, &pb.URLData{
			ShortUrl:    s.service.GetBaseURL() + "/" + shortURL,
			OriginalUrl: originalURL,
		})
	}

	return response, nil
}

