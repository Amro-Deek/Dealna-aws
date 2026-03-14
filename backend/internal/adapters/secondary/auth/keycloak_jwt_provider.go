package auth

import (
	"context"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type KeycloakJWTProvider struct {
	jwks   keyfunc.Keyfunc
	users  ports.IUserRepository
}

func NewKeycloakJWTProvider(jwksURL string, users ports.IUserRepository) (*KeycloakJWTProvider, error) {
	// Create the JWKS from Keycloak URL
	kf, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		return nil, err
	}

	return &KeycloakJWTProvider{
		jwks:  kf,
		users: users,
	}, nil
}

func (k *KeycloakJWTProvider) Authenticate(ctx context.Context, authHeader string) (*ports.AuthContext, error) {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, middleware.NewUnauthorizedError("missing bearer token")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse the JWT.
	token, err := jwt.Parse(tokenStr, k.jwks.Keyfunc)
	if err != nil {
		return nil, middleware.NewUnauthorizedError("invalid or expired token")
	}

	if !token.Valid {
		return nil, middleware.NewUnauthorizedError("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, middleware.NewUnauthorizedError("invalid token claims")
	}

	// In Keycloak, the subject is the unique user ID.
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return nil, middleware.NewUnauthorizedError("missing sub claim")
	}

	// Fetch user from database to ensure they are linked and get their internal ID/Role.
	user, err := k.users.GetByKeycloakSub(ctx, sub)
	if err != nil {
		return nil, middleware.NewUnauthorizedError("authenticated keycloak user is not linked to an internal account")
	}

	jti, _ := claims["jti"].(string)

	return &ports.AuthContext{
		UserID: user.ID,
		Role:   user.Role,
		JTI:    jti,
	}, nil
}