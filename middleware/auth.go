package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/samyak-jain/agora_backend/models"
)

var userContextKey = &contextKey{"user"}

type contextKey struct {
	name string
}

// AuthHandler is a middleware for authentication
func AuthHandler(db *models.Database) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")

			if header == "" {
				next.ServeHTTP(w, r)
			} else {
				splitToken := strings.Split(header, "Bearer ")
				token := splitToken[1]

				var user models.User
				if db.Where("token = ?", token).First(&user).RecordNotFound() {
					w.WriteHeader(http.StatusUnauthorized)
				} else {
					ctx := context.WithValue(r.Context(), userContextKey, user)
					next.ServeHTTP(w, r.WithContext(ctx))
				}
			}
		})
	}
}

// GetUserFromContext fetches the user from the context
func GetUserFromContext(ctx context.Context) *models.User {
	return ctx.Value(userContextKey).(*models.User)
}
