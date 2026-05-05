// Package middleware содержит HTTP-мидлвари сервиса: авторизацию,
// сжатие запросов/ответов и логирование входящих запросов.
package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// TokenService описывает зависимость мидлвари авторизации на сервис токенов.
type TokenService interface {
	// IsValid проверяет корректность токена.
	IsValid(token string) bool
	// GetToken выпускает токен для указанного userID.
	GetToken(userID uuid.UUID) (string, error)
	// GetUserID извлекает userID из токена.
	GetUserID(token string) (uuid.UUID, error)
}

type contextKey string

const (
	authCookieName              = "token"
	contextKeyUserID contextKey = "userID"
)

// GetUserID извлекает userID из контекста запроса.
// Возвращает uuid.Nil и false, если идентификатор в контексте отсутствует.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(contextKeyUserID).(uuid.UUID)
	return userID, ok
}

// SetUserID помещает userID в контекст и возвращает обновлённый контекст.
func SetUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}

// AuthorizationMiddleware читает cookie с токеном, проверяет его и кладёт userID в контекст.
// Если cookie отсутствует, создаётся новый пользователь и выпускается токен.
// При недействительном токене возвращается 401 Unauthorized.
func AuthorizationMiddleware(tokenService TokenService, logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie(authCookieName)
			if err != nil {
				if !errors.Is(err, http.ErrNoCookie) {
					logger.Error("failed to get cookie for authorization", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				cookie = nil
			}

			var userID uuid.UUID

			if cookie == nil {
				userID = uuid.New()

				token, err := tokenService.GetToken(userID)
				if err != nil {
					logger.Error("failed to generate token", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				cookie = &http.Cookie{
					Name:     authCookieName,
					Value:    token,
					HttpOnly: true,
					Path:     "/",
				}
			} else {
				token := cookie.Value

				if !tokenService.IsValid(token) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				userID, err = tokenService.GetUserID(token)
				if err != nil {
					logger.Error("failed to get userID from token", "error", err)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}

			http.SetCookie(w, cookie)
			ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
