package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Amro-Deek/Dealna-aws/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

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
				// Enforce HMAC signing method
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				utils.WriteJSON(w, http.StatusUnauthorized, false, "Unauthorized", nil, "Invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				utils.WriteJSON(w, http.StatusUnauthorized, false, "Unauthorized", nil, "Invalid claims")
				return
			}

			userID, ok := claims["user_id"].(string)
			if !ok || userID == "" {
				utils.WriteJSON(w, http.StatusUnauthorized, false, "Unauthorized", nil, "Missing user_id")
				return
			}

			role, ok := claims["role"].(string)
			if !ok || role == "" {
				utils.WriteJSON(w, http.StatusUnauthorized, false, "Unauthorized", nil, "Missing role")
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, userID)
			ctx = context.WithValue(ctx, ContextRole, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
