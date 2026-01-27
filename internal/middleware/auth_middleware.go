package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Amro-Deek/Dealna-aws/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

// GetProfile godoc
//
//	@Summary		Get current user profile
//	@Description	Returns authenticated user's ID and role
//	@Tags			User
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	utils.APIResponse
//	@Failure		401	{object}	utils.APIResponse
//	@Router			/me [get]
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				utils.WriteJSON(w, http.StatusUnauthorized, false, "Unauthorized", nil, "Missing token")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				utils.WriteJSON(w, http.StatusUnauthorized, false, "Unauthorized", nil, "Invalid token")
				return
			}

			claims := token.Claims.(jwt.MapClaims)

			ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
			ctx = context.WithValue(ctx, "role", claims["role"])

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
