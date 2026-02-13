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
	secret   string
	sessions ports.ISessionRepository
}

func NewJWTProvider(secret string, sessions ports.ISessionRepository) *JWTProvider {
	return &JWTProvider{secret: secret, sessions: sessions}
}

func (j *JWTProvider) GenerateAccessToken(userID, role, jti string, exp time.Time) (string, error) {
	return j.generate(userID, role, jti, ports.TokenAccess, exp)
}

func (j *JWTProvider) GenerateRefreshToken(userID, role, jti string, exp time.Time) (string, error) {
	return j.generate(userID, role, jti, ports.TokenRefresh, exp)
}

func (j *JWTProvider) generate(userID, role, jti string, typ ports.TokenType, exp time.Time) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"jti":     jti,
		"typ":     string(typ),
		"exp":     exp.Unix(),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(j.secret))
}

func (j *JWTProvider) ParseToken(tokenStr string) (*ports.ParsedToken, error) {
	tokenStr = strings.TrimSpace(tokenStr)
	if tokenStr == "" {
		return nil, middleware.NewUnauthorizedError("missing token")
	}

	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, middleware.NewUnauthorizedError("invalid signing method")
		}
		return []byte(j.secret), nil
	})

	if err != nil || !tok.Valid {
		return nil, middleware.NewUnauthorizedError("invalid or expired token")
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, middleware.NewUnauthorizedError("invalid token claims")
	}

	userID, _ := claims["user_id"].(string)
	role, _ := claims["role"].(string)
	jti, _ := claims["jti"].(string)
	typStr, _ := claims["typ"].(string)

	if userID == "" || jti == "" || typStr == "" {
		return nil, middleware.NewUnauthorizedError("missing token claims")
	}

	exp := time.Unix(int64(claims["exp"].(float64)), 0)

	return &ports.ParsedToken{
		UserID: userID,
		Role:   role,
		JTI:    jti,
		Type:   ports.TokenType(typStr),
		Exp:    exp,
	}, nil
}

func (j *JWTProvider) Authenticate(ctx context.Context, authHeader string) (*ports.AuthContext, error) {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, middleware.NewUnauthorizedError("missing bearer token")
	}

	raw := strings.TrimPrefix(authHeader, "Bearer ")
	parsed, err := j.ParseToken(raw)
	if err != nil {
		return nil, err
	}

	if parsed.Type != ports.TokenAccess {
		return nil, middleware.NewUnauthorizedError("invalid access token")
	}

	session, err := j.sessions.GetByJTI(ctx, parsed.JTI)
	if err != nil {
		return nil, middleware.NewDatabaseError("get session", err)
	}

	if session.Revoked || time.Now().After(session.ExpiresAt) {
		return nil, middleware.NewUnauthorizedError("session expired or revoked")
	}

	return &ports.AuthContext{
		UserID: parsed.UserID,
		Role:   parsed.Role,
		JTI:    parsed.JTI,
	}, nil
}
