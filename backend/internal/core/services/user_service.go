package services

import (
    "context"

    "github.com/Amro-Deek/Dealna-aws/internal/core/domain"
    "github.com/Amro-Deek/Dealna-aws/internal/core/ports"
)

type UserService struct {
    users ports.IUserRepository
}

func NewUserService(users ports.IUserRepository) *UserService {
    return &UserService{users: users}
}

func (s *UserService) GetByID(
    ctx context.Context,
    id string,
) (*domain.User, error) {

    return s.users.GetByID(ctx, id)
}

func (s *UserService) GetByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {

	return s.users.GetByEmail(ctx, email)
}


