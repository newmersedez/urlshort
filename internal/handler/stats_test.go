package handler

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/newmersedez/urlshort/internal/handler/mocks"
	middlewareMocks "github.com/newmersedez/urlshort/internal/middleware/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func parseCIDROrNil(cidr string) *net.IPNet {
	if cidr == "" {
		return nil
	}
	_, subnet, _ := net.ParseCIDR(cidr)
	return subnet
}

func TestGetStatsHandle(t *testing.T) {
	const trustedCIDR = "192.168.1.0/24"
	const trustedIP = "192.168.1.42"
	const outsideIP = "10.0.0.1"

	makeHandler := func(t *testing.T, cidr string, store *mocks.MockRepository) *handlers {
		t.Helper()
		var subnet = parseCIDROrNil(cidr)
		tokenService := middlewareMocks.NewMockTokenService(t)
		cleanupService := mocks.NewMockCleanupService(t)
		auditService := mocks.NewMockAuditService(t)
		h, err := newHandlers("http://localhost:8080", subnet, store, slog.Default(),
			mocks.NewMockShortener(t), tokenService, cleanupService, auditService)
		require.NoError(t, err)
		return h
	}

	t.Run("returns stats for trusted IP", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		mockRepo.EXPECT().Stats(mock.Anything).Return(5, 3, nil).Once()

		h := makeHandler(t, trustedCIDR, mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		req.Header.Set("X-Real-IP", trustedIP)
		w := httptest.NewRecorder()

		h.getStatsHandle(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

		var body statsResponse
		require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
		assert.Equal(t, 5, body.URLs)
		assert.Equal(t, 3, body.Users)
	})

	t.Run("returns 403 when trusted_subnet is not set", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		h := makeHandler(t, "", mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		req.Header.Set("X-Real-IP", trustedIP)
		w := httptest.NewRecorder()

		h.getStatsHandle(w, req)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	})

	t.Run("returns 403 when IP is outside trusted subnet", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		h := makeHandler(t, trustedCIDR, mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		req.Header.Set("X-Real-IP", outsideIP)
		w := httptest.NewRecorder()

		h.getStatsHandle(w, req)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	})

	t.Run("returns 403 when X-Real-IP header is missing", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		h := makeHandler(t, trustedCIDR, mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		w := httptest.NewRecorder()

		h.getStatsHandle(w, req)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	})

	t.Run("returns 403 when X-Real-IP is not a valid IP", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		h := makeHandler(t, trustedCIDR, mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		req.Header.Set("X-Real-IP", "not-an-ip")
		w := httptest.NewRecorder()

		h.getStatsHandle(w, req)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	})
}
