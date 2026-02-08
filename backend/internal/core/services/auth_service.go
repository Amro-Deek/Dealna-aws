package services

import (
	"context"
	"time"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type AuthService struct {
	users  ports.IUserRepository
	hasher ports.IPasswordHasher
	tokens ports.ITokenProvider
}

func NewAuthService(
	users ports.IUserRepository,
	hasher ports.IPasswordHasher,
	tokens ports.ITokenProvider,
) *AuthService {
	return &AuthService{
		users:  users,
		hasher: hasher,
		tokens: tokens,
	}
}

type AuthResult struct {
	AccessToken string
}
func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (*AuthResult, error) {

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// لا نكشف هل الإيميل موجود أو لا
		return nil, middleware.NewInvalidCredentialsError()
	}

	if err := s.hasher.Compare(user.PasswordHash, password); err != nil {
		return nil, middleware.NewInvalidCredentialsError()
	}

	token, err := s.tokens.GenerateToken(
		user.ID,
		user.Role,
		time.Now().Add(24*time.Hour),
	)
	if err != nil {
		return nil, middleware.NewInternalError(err)
	}

	return &AuthResult{AccessToken: token}, nil
}

