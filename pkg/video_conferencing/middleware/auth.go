package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/samyak-jain/agora_backend/pkg/video_conferencing/models"
	"github.com/samyak-jain/agora_backend/utils"

	"github.com/spf13/viper"
)

type contextKey struct {
	name string
}

var userContextKey = &contextKey{"user"}

// AuthHandler is a middleware for authentication
func AuthHandler(db *models.Database, logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			if !viper.GetBool("ENABLE_OAUTH") {
				next.ServeHTTP(w, r)
				return
			}

			header := r.Header.Get("Authorization")

			if header == "" {
				logger.Debug().Msg("No Token Provided")
				next.ServeHTTP(w, r)
			} else {
				splitToken := strings.Split(header, "Bearer ")
				token := splitToken[1]

				var tokenData models.Token
				var user models.User

				if db.Where("token_id = ?", token).First(&tokenData).RecordNotFound() {
					w.WriteHeader(http.StatusUnauthorized)
					logger.Debug().Str("token", token).Msg("Passed Invalid token")
				} else if err := db.Where("email = ?", tokenData.UserEmail).First(&user).Error; err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Error().Str("email", tokenData.UserEmail).Str("token", token).Msg("Email does not exist for the provided token")
				} else {
					logger.Info().Str("token", token)
					ctx := context.WithValue(r.Context(), userContextKey, &user)
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
