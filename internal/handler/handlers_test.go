package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/newmersedez/urlshort/internal/model"
	"github.com/newmersedez/urlshort/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestCanShortenValidUrl(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	repo := NewMockRepository()
	urlShortener := service.NewURLShortenerService()
	handler := ShortenURLHandler(baseURL, repo, urlShortener)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://stackoverflow.com"))
	w := httptest.NewRecorder()

	// Act
	handler(w, request)

	//Assert
	res := w.Result()
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)
	assert.NotEmpty(t, string(resBody))
	assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
}

func TestCannotShortenInvalidUrl(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	repo := NewMockRepository()
	urlShortener := service.NewURLShortenerService()
	handler := ShortenURLHandler(baseURL, repo, urlShortener)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("url//string"))
	w := httptest.NewRecorder()

	// Act
	handler(w, request)

	//Assert
	res := w.Result()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestCanGetFullUrlByShortenValue(t *testing.T) {
	// Arrange
	baseUrl := "http://localhost:8080"

	repo := NewMockRepository()
	shortenUrl := model.ShortenURL{
		ShortenValue:  "12345678",
		OriginalValue: "https://stackoverflow.com",
	}
	repo.Add(&shortenUrl)

	handler := GetURLByShortenValueHandler(baseUrl, repo)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "12345678")

	// Добавляем контекст в запрос
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	// Act
	handler(w, request)

	//Assert
	res := w.Result()

	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.NotEmpty(t, string(res.Header.Get("Location")))
}

func TestCannotGetFullUrlByShortenValueIfIdIsNotSpecified(t *testing.T) {
	// Arrange
	baseUrl := "http://localhost:8080"

	repo := NewMockRepository()
	handler := GetURLByShortenValueHandler(baseUrl, repo)

	request := httptest.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	w := httptest.NewRecorder()

	// Act
	handler(w, request)

	//Assert
	res := w.Result()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestCannotGetFullUrlByShortenValueIfItDoesNotExist(t *testing.T) {
	// Arrange
	baseUrl := "http://localhost:8080"

	repo := NewMockRepository()
	handler := GetURLByShortenValueHandler(baseUrl, repo)

	request := httptest.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	request.SetPathValue("id", "1337")
	w := httptest.NewRecorder()

	// Act
	handler(w, request)

	//Assert
	res := w.Result()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

type MockRepository struct {
	data map[string]model.ShortenURL
}

func NewMockRepository() Repository {
	return &MockRepository{
		data: make(map[string]model.ShortenURL),
	}
}

func (r *MockRepository) GetByShortenValue(shortenValue string) *model.ShortenURL {
	shortenUrl, exists := r.data[shortenValue]
	if !exists {
		return nil
	}

	return &shortenUrl
}

func (r *MockRepository) Add(shortenUrl *model.ShortenURL) {
	r.data[shortenUrl.ShortenValue] = *shortenUrl
}
