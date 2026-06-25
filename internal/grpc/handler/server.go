// Package handler реализует gRPC-сервер сервиса сокращения URL.
package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/google/uuid"
	pb "github.com/newmersedez/urlshort/internal/grpc/proto"
	"github.com/newmersedez/urlshort/internal/model"
	"github.com/newmersedez/urlshort/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Repository — интерфейс хранилища (тот же, что и у HTTP-сервера).
type Repository interface {
	Get(ctx context.Context, id string) (*model.ShortenURL, error)
	GetList(ctx context.Context, userID uuid.UUID) ([]model.ShortenURL, error)
	Add(ctx context.Context, shortenURL *model.ShortenURL) error
}

// ShortenerService — интерфейс сервиса генерации коротких идентификаторов.
type ShortenerService interface {
	Shorten(url string) (string, error)
}

// TokenService — интерфейс работы с токенами авторизации.
type TokenService interface {
	IsValid(token string) bool
	GetUserID(token string) (uuid.UUID, error)
}

// ShortenerServer реализует gRPC-интерфейс ShortenerServiceServer.
type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer

	baseURL          string
	store            Repository
	shortenerService ShortenerService
	tokenService     TokenService
	logger           *slog.Logger
}

// NewShortenerServer создаёт и возвращает настроенный gRPC-сервер.
// Авторизация проверяется через перехватчик по заголовку metadata authorization.
func NewShortenerServer(
	baseURL string,
	store Repository,
	shortenerService ShortenerService,
	tokenService TokenService,
	logger *slog.Logger,
) *grpc.Server {
	server := &ShortenerServer{
		baseURL:          strings.TrimRight(baseURL, "/") + "/",
		store:            store,
		shortenerService: shortenerService,
		tokenService:     tokenService,
		logger:           logger,
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(server.authInterceptor))
	pb.RegisterShortenerServiceServer(s, server)
	return s
}

// Serve запускает gRPC-сервер по указанному адресу.
func Serve(ctx context.Context, grpcServer *grpc.Server, addr string) error {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(listen); err != nil {
		return fmt.Errorf("grpc server failed: %w", err)
	}
	return nil
}

// authInterceptor проверяет токен авторизации из metadata заголовка authorization.
func (s *ShortenerServer) authInterceptor(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 || values[0] == "" {
		return nil, status.Error(codes.Unauthenticated, "missing authorization token")
	}

	token := values[0]
	if !s.tokenService.IsValid(token) {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization token")
	}

	userID, err := s.tokenService.GetUserID(token)
	if err != nil || userID == uuid.Nil {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization token")
	}

	ctx = context.WithValue(ctx, userIDKey{}, userID)
	return handler(ctx, req)
}

type userIDKey struct{}

func getUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey{}).(uuid.UUID)
	return id, ok && id != uuid.Nil
}

// ShortenURL реализует POST /api/shorten через gRPC.
func (s *ShortenerServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	id, err := s.shortenerService.Shorten(req.GetUrl())
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			return nil, status.Error(codes.InvalidArgument, "invalid url")
		}
		s.logger.Error("shortener service error", "error", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	shortenURL := model.NewShortenURL(userID, id, req.GetUrl())
	if err := s.store.Add(ctx, shortenURL); err != nil {
		s.logger.Error("failed to save url", "error", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var resp pb.URLShortenResponse
	resp.SetResult(s.baseURL + id)
	return &resp, nil
}

// ExpandURL реализует GET /<id> через gRPC.
func (s *ShortenerServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	url, err := s.store.Get(ctx, req.GetId())
	if err != nil {
		s.logger.Error("failed to get url", "id", req.GetId(), "error", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}
	if url == nil {
		return nil, status.Errorf(codes.NotFound, "url with id %q not found", req.GetId())
	}
	if url.Deleted {
		return nil, status.Error(codes.NotFound, "url has been deleted")
	}

	var resp pb.URLExpandResponse
	resp.SetResult(url.OriginalValue)
	return &resp, nil
}

// ListUserURLs реализует GET /api/user/urls через gRPC.
func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *pb.ListUserURLsRequest) (*pb.UserURLsResponse, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	urls, err := s.store.GetList(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user urls", "userID", userID, "error", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	urlData := make([]*pb.URLData, 0, len(urls))
	for _, u := range urls {
		item := pb.URLData_builder{
			ShortUrl:    proto.String(s.baseURL + u.ID),
			OriginalUrl: proto.String(u.OriginalValue),
		}
		urlData = append(urlData, item.Build())
	}

	var resp pb.UserURLsResponse
	resp.SetUrl(urlData)
	return &resp, nil
}
