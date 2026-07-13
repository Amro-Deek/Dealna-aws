package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type AuthService struct {
	users    ports.IUserRepository
	identity ports.IIdentityProvider
	email    ports.IEmailService
}

func NewAuthService(
	users ports.IUserRepository,
	identity ports.IIdentityProvider,
	email ports.IEmailService,
) *AuthService {
	return &AuthService{
		users:    users,
		identity: identity,
		email:    email,
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

func generateResetCode() string {
	max := big.NewInt(1000000)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	_, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// Silent failure to prevent email enumeration
		return nil
	}

	code := generateResetCode()
	expiresAt := time.Now().Add(15 * time.Minute)

	err = s.users.CreatePasswordResetToken(ctx, email, code, expiresAt)
	if err != nil {
		return err
	}

	return s.email.SendPasswordResetEmail(email, code)
}

func (s *AuthService) ConfirmPasswordReset(ctx context.Context, email, token, newPassword string) error {
	resetToken, err := s.users.GetPasswordResetToken(ctx, email, token)
	if err != nil {
		return errors.New("invalid or expired reset code")
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return errors.New("reset code has expired")
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("user not found")
	}

	err = s.identity.ResetPassword(ctx, user.KeycloakSub, newPassword)
	if err != nil {
		return err
	}

	_ = s.users.DeletePasswordResetToken(ctx, email)

	return nil
}
