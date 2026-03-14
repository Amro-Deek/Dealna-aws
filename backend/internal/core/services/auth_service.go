package services

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type AuthService struct {
	users    ports.IUserRepository
	identity ports.IIdentityProvider
}

func NewAuthService(
	users ports.IUserRepository,
	identity ports.IIdentityProvider,
) *AuthService {
	return &AuthService{
		users:    users,
		identity: identity,
	}
}

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	User         *domain.User
}

func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (*AuthResult, error) {

	loginRes, err := s.identity.Login(ctx, email, password)
	if err != nil {
		return nil, err
	}

	user, err := s.users.GetByKeycloakSub(ctx, loginRes.Subject)
	if err != nil {
		return nil, middleware.NewUnauthorizedError("authenticated keycloak user is not linked to an internal account")
	}

	return &AuthResult{
		AccessToken:  loginRes.AccessToken,
		RefreshToken: loginRes.RefreshToken,
		ExpiresIn:    loginRes.ExpiresIn,
		User:         user,
	}, nil
}

func (s *AuthService) Refresh(
	ctx context.Context,
	refreshToken string,
) (*AuthResult, error) {
	loginRes, err := s.identity.Refresh(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.users.GetByKeycloakSub(ctx, loginRes.Subject)
	if err != nil {
		return nil, middleware.NewUnauthorizedError("authenticated keycloak user is not linked to an internal account")
	}

	return &AuthResult{
		AccessToken:  loginRes.AccessToken,
		RefreshToken: loginRes.RefreshToken,
		ExpiresIn:    loginRes.ExpiresIn,
		User:         user,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.identity.Logout(ctx, refreshToken)
}