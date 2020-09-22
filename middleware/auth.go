package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/samyak-jain/agora_backend/models"
	"github.com/urfave/negroni"
)

var userContextKey = &contextKey{"user"}

type contextKey struct {
	name string
}

// AuthHandler is a middleware for authentication
func AuthHandler(db *models.Database) negroni.HandlerFunc {
	return negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		header := r.Header.Get("Authorization")

		if header == "" {
			next.ServeHTTP(w, r)
		} else {
			splitToken := strings.Split(header, "Bearer ")
			token := splitToken[1]

			var tokenData models.Token
			var user models.User

			if db.Where("token_id = ?", token).First(&tokenData).RecordNotFound() {
				w.WriteHeader(http.StatusUnauthorized)
			} else if err := db.Where("email = ?", tokenData.UserEmail).First(&user).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				ctx := context.WithValue(r.Context(), userContextKey, &user)
				next.ServeHTTP(w, r.WithContext(ctx))
			}

		}
	})
}

// GetUserFromContext fetches the user from the context
func GetUserFromContext(ctx context.Context) *models.User {
	return ctx.Value(userContextKey).(*models.User)
}
