package auth


import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
)

type Handler struct {
	authService      *services.AuthService
	preRegService    *services.StudentRegistrationService
	providerRegSvc   *services.ProviderRegistrationService
	userService      *services.UserService
}

func NewHandler(
	authService *services.AuthService,
	preRegService *services.StudentRegistrationService,
	providerRegSvc *services.ProviderRegistrationService,
	userService *services.UserService,
) *Handler {
	return &Handler{
		authService:    authService,
		preRegService:  preRegService,
		providerRegSvc: providerRegSvc,
		userService:    userService,
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
	refreshToken string,
) error {
	return h.authService.Logout(ctx, refreshToken)
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

func (h *Handler) GetStudentRegistrationStatus(
	ctx context.Context,
	email string,
) (*domain.StudentPreRegistration, error) {
	return h.preRegService.GetRegistrationStatus(ctx, email)
}

func (h *Handler) RequestProviderActivation(
	ctx context.Context,
	email string,
) error {
	return h.providerRegSvc.RequestProviderActivation(ctx, email)
}

func (h *Handler) VerifyProviderActivation(
	ctx context.Context,
	token string,
) error {
	return h.providerRegSvc.VerifyProviderActivation(ctx, token)
}

func (h *Handler) CompleteProviderRegistration(
	ctx context.Context,
	email string,
	password string,
) error {
	return h.providerRegSvc.CompleteProviderRegistration(ctx, email, password)
}

func (h *Handler) ResendProviderActivation(
	ctx context.Context,
	email string,
) error {
	return h.providerRegSvc.ResendProviderActivation(ctx, email)
}
func (h *Handler) StartProviderApplication(
	ctx context.Context,
	userID string,
	universityID string,
	businessName string,
	phoneNumber string,
	businessType string,
	address string,
) (*domain.ProviderApplication, error) {
	return h.providerRegSvc.StartApplication(ctx, userID, universityID, businessName, phoneNumber, businessType, address)
}

func (h *Handler) GetDocumentUploadURL(
	ctx context.Context,
	userID string,
	documentType string,
	originalFilename string,
	contentType string,
) (string, string, error) {
	return h.providerRegSvc.GetDocumentUploadURL(ctx, userID, documentType, originalFilename, contentType)
}

func (h *Handler) ConfirmDocumentUpload(
	ctx context.Context,
	userID string,
	objectKey string,
	documentType string,
	originalFilename string,
	contentType string,
	sizeBytes int64,
) error {
	return h.providerRegSvc.ConfirmDocumentUpload(ctx, userID, objectKey, documentType, originalFilename, contentType, sizeBytes)
}

func (h *Handler) SubmitProviderApplication(ctx context.Context, userID string) error {
	return h.providerRegSvc.SubmitApplication(ctx, userID)
}

func (h *Handler) ApproveProviderApplication(ctx context.Context, adminID, applicantID string) error {
	return h.providerRegSvc.ApproveApplication(ctx, adminID, applicantID)
}

func (h *Handler) RejectProviderApplication(ctx context.Context, adminID, applicantID, comment string) error {
	return h.providerRegSvc.RejectApplication(ctx, adminID, applicantID, comment)
}

func (h *Handler) GetProviderApplicationStatus(ctx context.Context, applicantID string) (*domain.ProviderApplication, error) {
	return h.providerRegSvc.GetApplicationStatus(ctx, applicantID)
}
