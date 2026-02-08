package auth

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type JWTProvider struct {
	secret string
}

func NewJWTProvider(secret string) *JWTProvider {
	return &JWTProvider{secret: secret}
}

//
// =========================
// TokenProvider
// =========================
//

func (j *JWTProvider) GenerateToken(
	userID string,
	role string,
	expiresAt time.Time,
) (string, error) {

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(j.secret))
}

//
// =========================
// AuthContextProvider
// =========================
//

var _ ports.IAuthContextProvider = (*JWTProvider)(nil)

func (j *JWTProvider) Authenticate(
	ctx context.Context,
	authHeader string,
) (*ports.AuthContext, error) {

	if strings.TrimSpace(authHeader) == "" {
		return nil, middleware.NewUnauthorizedError("missing authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, middleware.NewUnauthorizedError("invalid authorization header")
	}

	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenStr == "" {
		return nil, middleware.NewUnauthorizedError("missing bearer token")
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// ⛔ Security: enforce signing method
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, middleware.NewUnauthorizedError("invalid signing method")
		}
		return []byte(j.secret), nil
	})

	if err != nil || !token.Valid {
		return nil, middleware.NewUnauthorizedError("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, middleware.NewUnauthorizedError("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return nil, middleware.NewUnauthorizedError("invalid token user")
	}

	role, _ := claims["role"].(string)

	return &ports.AuthContext{
		UserID: userID,
		Role:   role,
	}, nil
}
