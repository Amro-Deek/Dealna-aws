package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type AuthService struct {
	users    ports.IUserRepository
	hasher   ports.IPasswordHasher
	tokens   ports.ITokenProvider
	sessions ports.ISessionRepository
}

func NewAuthService(
	users ports.IUserRepository,
	hasher ports.IPasswordHasher,
	tokens ports.ITokenProvider,
	sessions ports.ISessionRepository,
) *AuthService {
	return &AuthService{
		users:    users,
		hasher:   hasher,
		tokens:   tokens,
		sessions: sessions,
	}
}

type AuthResult struct {
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (*AuthResult, error) {

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, middleware.NewInvalidCredentialsError()
	}

	if err := s.hasher.Compare(user.PasswordHash, password); err != nil {
		return nil, middleware.NewInvalidCredentialsError()
	}

	// 1️⃣ revoke ALL previous sessions (single active session)
	if err := s.sessions.RevokeAllForUser(ctx, user.ID); err != nil {
		return nil, middleware.NewInternalError(err)
	}

	jti := uuid.NewString()
	refreshExp := time.Now().Add(30 * 24 * time.Hour)

	if err := s.sessions.Create(ctx, user.ID, jti, refreshExp); err != nil {
		return nil, middleware.NewDatabaseError("create session", err)
	}

	accessExp := time.Now().Add(15 * time.Minute)

	access, err := s.tokens.GenerateAccessToken(
		user.ID,
		user.Role,
		jti,
		accessExp,
	)
	if err != nil {
		return nil, middleware.NewInternalError(err)
	}

	refresh, err := s.tokens.GenerateRefreshToken(
		user.ID,
		user.Role,
		jti,
		refreshExp,
	)
	if err != nil {
		return nil, middleware.NewInternalError(err)
	}

	return &AuthResult{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *AuthService) Refresh(
	ctx context.Context,
	refreshToken string,
) (*AuthResult, error) {

	parsed, err := s.tokens.ParseToken(refreshToken)
	if err != nil {
		return nil, err // already DomainError
	}

	if parsed.Type != ports.TokenRefresh {
		return nil, middleware.NewUnauthorizedError("invalid refresh token")
	}

	session, err := s.sessions.GetByJTI(ctx, parsed.JTI)
	if err != nil {
		return nil, middleware.NewDatabaseError("get session", err)
	}

	if session.Revoked || time.Now().After(session.ExpiresAt) {
		return nil, middleware.NewUnauthorizedError("session expired or revoked")
	}

	// rotate session
	if err := s.sessions.RevokeByJTI(ctx, parsed.JTI); err != nil {
		return nil, middleware.NewDatabaseError("revoke session", err)
	}

	newJTI := uuid.NewString()
	newRefreshExp := time.Now().Add(30 * 24 * time.Hour)

	if err := s.sessions.Create(ctx, parsed.UserID, newJTI, newRefreshExp); err != nil {
		return nil, middleware.NewDatabaseError("create session", err)
	}

	accessExp := time.Now().Add(15 * time.Minute)

	access, err := s.tokens.GenerateAccessToken(
		parsed.UserID,
		parsed.Role,
		newJTI,
		accessExp,
	)
	if err != nil {
		return nil, middleware.NewInternalError(err)
	}

	refresh, err := s.tokens.GenerateRefreshToken(
		parsed.UserID,
		parsed.Role,
		newJTI,
		newRefreshExp,
	)
	if err != nil {
		return nil, middleware.NewInternalError(err)
	}

	return &AuthResult{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, jti string) error {
	if err := s.sessions.RevokeByJTI(ctx, jti); err != nil {
		return middleware.NewDatabaseError("revoke session", err)
	}
	return nil
}
