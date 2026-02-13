package auth


import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
)

type Handler struct {
	authService      *services.AuthService
	preRegService    *services.StudentRegistrationService
	userService      *services.UserService
}

func NewHandler(
	authService *services.AuthService,
	preRegService *services.StudentRegistrationService,
	userService *services.UserService,
) *Handler {
	return &Handler{
		authService:   authService,
		preRegService: preRegService,
		userService:   userService,
	}
}


func (h *Handler) Login(ctx context.Context, email, password string) (*services.AuthResult, error) {
	return h.authService.Login(ctx, email, password)
}

// =========================
// REFRESH
// =========================
func (h *Handler) Refresh(
	ctx context.Context,
	refreshToken string,
) (*services.AuthResult, error) {
	return h.authService.Refresh(ctx, refreshToken)
}

// =========================
// LOGOUT
// =========================
func (h *Handler) Logout(
	ctx context.Context,
	jti string,
) error {
	return h.authService.Logout(ctx, jti)
}

func (h *Handler) RequestStudentActivation(
	ctx context.Context,
	email string,
) error {
	return h.preRegService.RequestStudentActivation(ctx, email)
}

func (h *Handler) VerifyStudentActivation(
	ctx context.Context,
	token string,
) error {
	return h.preRegService.VerifyActivation(ctx, token)
}

func (h *Handler) CompleteStudentRegistration(
	ctx context.Context,
	email string,
	displayName string,
	password string,
	major *string,
	year *int,
) error {
	return h.preRegService.CompleteStudentRegistration(
		ctx,
		email,
		displayName,
		password,
		major,
		year,
	)
}

func (h *Handler) ResendStudentActivation(
	ctx context.Context,
	email string,
) error {
	return h.preRegService.ResendActivation(ctx, email)
}
