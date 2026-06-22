package handler

import (
	"context"
	"net"
	"testing"

	"github.com/google/uuid"
	pb "github.com/newmersedez/urlshort/internal/grpc/proto"
	"github.com/newmersedez/urlshort/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// --- моки ---

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Get(ctx context.Context, id string) (*model.ShortenURL, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*model.ShortenURL), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepo) GetList(ctx context.Context, userID uuid.UUID) ([]model.ShortenURL, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.ShortenURL), args.Error(1)
}

func (m *mockRepo) Add(ctx context.Context, url *model.ShortenURL) error {
	return m.Called(ctx, url).Error(0)
}

type mockShortener struct{ mock.Mock }

func (m *mockShortener) Shorten(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

type mockTokenService struct{ mock.Mock }

func (m *mockTokenService) IsValid(token string) bool {
	return m.Called(token).Bool(0)
}

func (m *mockTokenService) GetUserID(token string) (uuid.UUID, error) {
	args := m.Called(token)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

// --- вспомогательная функция: запускает gRPC сервер и возвращает клиент ---

func startTestServer(t *testing.T, repo Repository, shortener ShortenerService, tokens TokenService) pb.ShortenerServiceClient {
	t.Helper()

	grpcServer := NewShortenerServer("http://localhost:8080", repo, shortener, tokens)

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	t.Cleanup(grpcServer.Stop)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return pb.NewShortenerServiceClient(conn)
}

func ctxWithToken(token string) context.Context {
	md := metadata.Pairs("authorization", token)
	return metadata.NewOutgoingContext(context.Background(), md)
}

// --- тесты ---

func TestShortenURL_Success(t *testing.T) {
	userID := uuid.New()
	token := "valid-token"

	repo := new(mockRepo)
	shortener := new(mockShortener)
	tokens := new(mockTokenService)

	tokens.On("IsValid", token).Return(true)
	tokens.On("GetUserID", token).Return(userID, nil)
	shortener.On("Shorten", "https://example.com").Return("abc123", nil)
	repo.On("Add", mock.Anything, mock.MatchedBy(func(u *model.ShortenURL) bool {
		return u.ID == "abc123" && u.OriginalValue == "https://example.com"
	})).Return(nil)

	client := startTestServer(t, repo, shortener, tokens)

	resp, err := client.ShortenURL(ctxWithToken(token), pb.URLShortenRequest_builder{
		Url: proto.String("https://example.com"),
	}.Build())

	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/abc123", resp.GetResult())
}

func TestShortenURL_Unauthenticated(t *testing.T) {
	repo := new(mockRepo)
	shortener := new(mockShortener)
	tokens := new(mockTokenService)

	tokens.On("IsValid", "bad-token").Return(false)

	client := startTestServer(t, repo, shortener, tokens)

	_, err := client.ShortenURL(ctxWithToken("bad-token"), pb.URLShortenRequest_builder{
		Url: proto.String("https://example.com"),
	}.Build())

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestShortenURL_MissingToken(t *testing.T) {
	client := startTestServer(t, new(mockRepo), new(mockShortener), new(mockTokenService))

	_, err := client.ShortenURL(context.Background(), pb.URLShortenRequest_builder{
		Url: proto.String("https://example.com"),
	}.Build())

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestExpandURL_Success(t *testing.T) {
	userID := uuid.New()
	token := "valid-token"
	url := model.NewShortenURL(userID, "abc123", "https://example.com")

	repo := new(mockRepo)
	tokens := new(mockTokenService)

	tokens.On("IsValid", token).Return(true)
	tokens.On("GetUserID", token).Return(userID, nil)
	repo.On("Get", mock.Anything, "abc123").Return(url, nil)

	client := startTestServer(t, repo, new(mockShortener), tokens)

	resp, err := client.ExpandURL(ctxWithToken(token), pb.URLExpandRequest_builder{
		Id: proto.String("abc123"),
	}.Build())

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", resp.GetResult())
}

func TestExpandURL_NotFound(t *testing.T) {
	userID := uuid.New()
	token := "valid-token"

	repo := new(mockRepo)
	tokens := new(mockTokenService)

	tokens.On("IsValid", token).Return(true)
	tokens.On("GetUserID", token).Return(userID, nil)
	repo.On("Get", mock.Anything, "notexist").Return(nil, nil)

	client := startTestServer(t, repo, new(mockShortener), tokens)

	_, err := client.ExpandURL(ctxWithToken(token), pb.URLExpandRequest_builder{
		Id: proto.String("notexist"),
	}.Build())

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestListUserURLs_Success(t *testing.T) {
	userID := uuid.New()
	token := "valid-token"
	urls := []model.ShortenURL{
		*model.NewShortenURL(userID, "id1", "https://one.com"),
		*model.NewShortenURL(userID, "id2", "https://two.com"),
	}

	repo := new(mockRepo)
	tokens := new(mockTokenService)

	tokens.On("IsValid", token).Return(true)
	tokens.On("GetUserID", token).Return(userID, nil)
	repo.On("GetList", mock.Anything, userID).Return(urls, nil)

	client := startTestServer(t, repo, new(mockShortener), tokens)

	resp, err := client.ListUserURLs(ctxWithToken(token), &emptypb.Empty{})

	require.NoError(t, err)
	assert.Len(t, resp.GetUrl(), 2)
	assert.Equal(t, "http://localhost:8080/id1", resp.GetUrl()[0].GetShortUrl())
	assert.Equal(t, "https://one.com", resp.GetUrl()[0].GetOriginalUrl())
}
