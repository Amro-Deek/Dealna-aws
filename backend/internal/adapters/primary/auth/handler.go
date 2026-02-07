package auth

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
)

type Handler struct {
	authService *services.AuthService
}

func NewHandler(authService *services.AuthService) *Handler {
	return &Handler{
		authService: authService,
	}
}

func (h *Handler) Login(ctx context.Context, email, password string) (*services.AuthResult, error) {
	return h.authService.Login(ctx, email, password)
}
