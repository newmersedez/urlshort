package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/newmersedez/urlshort/internal/model"
)

type TokenService interface {
	IsValid(token string) bool
	GetToken(userID uuid.UUID) (string, error)
	GetUserId(token string) (uuid.UUID, error)
}

func AuthorizationMiddleware(tokenService TokenService, logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const authCookieName = "token"
			const contextKeyUserID = "userID"

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
				user := model.NewUser()
				userID = user.ID

				token, err := tokenService.GetToken(user.ID)
				if err != nil {
					logger.Error("failed to generate token", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				cookie = &http.Cookie{
					Name:     authCookieName,
					Value:    token,
					HttpOnly: true,
					Secure:   true,
					Path:     "/",
				}
			} else {
				token := cookie.Value

				if !tokenService.IsValid(token) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				userID, err = tokenService.GetUserId(token)
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
