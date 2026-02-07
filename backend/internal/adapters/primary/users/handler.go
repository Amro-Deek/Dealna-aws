package users

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
)

type Handler struct {
	userService *services.UserService
}

func NewHandler(userService *services.UserService) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) GetMe(
	ctx context.Context,
	userID string,
) (*domain.User, error) {
	return h.userService.GetByID(ctx, userID)
}
