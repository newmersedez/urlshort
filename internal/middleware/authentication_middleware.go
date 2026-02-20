package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/newmersedez/urlshort/internal/model"
)

type AuthenticationService interface {
	IsValid(token string) bool
	GetToken(userID uuid.UUID) (string, error)
	GetUserId(token string) (string, error)
}

type UserRepository interface {
	AddUser(ctx context.Context, user *model.User) error
}

func AuthenticationMiddleware(
	authService AuthenticationService, 
	userRepository UserRepository, 
	logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const authCookieName = "token"
	
			cookie, err := r.Cookie(authCookieName)
			if err != nil && !errors.Is(err, http.ErrNoCookie) {
				logger.Error("failed to get cookie for authorization", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if cookie == nil {
				user := model.NewUser()
				userRepository.AddUser(r.Context(), user);

				token, err := authService.GetToken(user.ID)
				if err != nil {
					logger.Error("failed to generate token", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				cookie = &http.Cookie{
					Name: authCookieName,
					Value: token,
					HttpOnly: true,
					Secure: true,
				}
			} else if !authService.IsValid(cookie.Value) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(w, r)
			http.SetCookie(w, cookie)
		})
	}
}