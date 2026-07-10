package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/google/uuid"
)

type ProviderRegistrationService struct {
	users     ports.IUserRepository
	providers ports.IProviderRepository
	preRegs   ports.IProviderPreRegistrationRepository
	email     ports.IEmailService
	identity  ports.IIdentityProvider
	storage   ports.IStorageProvider
	notifs    *NotificationService
}

func NewProviderRegistrationService(
	users ports.IUserRepository,
	providers ports.IProviderRepository,
	preRegs ports.IProviderPreRegistrationRepository,
	email ports.IEmailService,
	identity ports.IIdentityProvider,
	storage ports.IStorageProvider,
	notifs *NotificationService,
) *ProviderRegistrationService {
	return &ProviderRegistrationService{
		users:     users,
		providers: providers,
		preRegs:   preRegs,
		email:     email,
		identity:  identity,
		storage:   storage,
		notifs:    notifs,
	}
}

func (s *ProviderRegistrationService) RequestProviderActivation(
	ctx context.Context,
	email string,
) error {
	email = strings.ToLower(strings.TrimSpace(email))

	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return middleware.NewEmailAlreadyUsedError(email)
	}

	if existingPre, err := s.preRegs.GetByEmail(ctx, email); err == nil && existingPre != nil {
		if existingPre.UsedAt == nil {
			token := uuid.NewString()
			existingPre.Token = token
			existingPre.ExpiresAt = time.Now().Add(24 * time.Hour)
			existingPre.VerifiedAt = nil
			if err := s.preRegs.Update(ctx, existingPre); err != nil {
				return middleware.NewDatabaseError("update prereg", err)
			}
			link := fmt.Sprintf("http://98.92.82.224:8080/api/v1/auth/providers/activate?token=%s", token)
			return s.email.SendActivationLink(email, link, "provider")
		}
		return middleware.NewEmailAlreadyUsedError(email)
	}

	token := uuid.NewString()

	pre := &domain.ProviderPreRegistration{
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.preRegs.Create(ctx, pre); err != nil {
		return middleware.NewDatabaseError("create prereg", err)
	}

	link := fmt.Sprintf("http://98.92.82.224:8080/api/v1/auth/providers/activate?token=%s", token)
	return s.email.SendActivationLink(email, link, "provider")
}

func (s *ProviderRegistrationService) VerifyProviderActivation(
	ctx context.Context,
	token string,
) error {
	pre, err := s.preRegs.GetByToken(ctx, token)
	if err != nil {
		return middleware.NewUnauthorizedError("invalid token")
	}

	if pre.VerifiedAt != nil {
		return middleware.NewUnauthorizedError("already verified")
	}

	if time.Now().After(pre.ExpiresAt) {
		return middleware.NewUnauthorizedError("token expired")
	}

	now := time.Now()
	pre.VerifiedAt = &now

	if err := s.preRegs.Update(ctx, pre); err != nil {
		return middleware.NewDatabaseError("update prereg", err)
	}

	return nil
}

func (s *ProviderRegistrationService) GetRegistrationStatus(
	ctx context.Context,
	email string,
) (*domain.ProviderPreRegistration, error) {
	return s.preRegs.GetByEmail(ctx, email)
}

func (s *ProviderRegistrationService) CompleteProviderRegistration(
	ctx context.Context,
	email string,
	password string,
) (*AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	pre, err := s.preRegs.GetByEmail(ctx, email)
	if err != nil {
		return nil, middleware.NewUnauthorizedError("activation not requested")
	}

	if pre.VerifiedAt == nil {
		return nil, middleware.NewUnauthorizedError("email not verified")
	}

	if pre.UsedAt != nil {
		return nil, middleware.NewUnauthorizedError("registration already completed")
	}

	if time.Now().After(pre.ExpiresAt) {
		return nil, middleware.NewUnauthorizedError("activation expired")
	}

	nameParts := strings.Split(email, "@")
	firstName := nameParts[0]

	// Register in Keycloak (emailVerified=true because they clicked our token)
	keycloakSub, err := s.identity.RegisterUser(ctx, email, password, firstName, firstName)
	if err != nil {
		fmt.Printf("Error registering user in Keycloak: %v\n", err)
		return nil, err
	}

	// Make sure Keycloak sets email verified
	_ = s.identity.ExecuteActionsEmail(ctx, keycloakSub, []string{}) // clear any actions if needed, though they shouldn't exist since we didn't trigger VERIFY_EMAIL

	err = s.identity.AssignRoleToUser(ctx, keycloakSub, "APPLICANT")
	if err != nil {
		fmt.Printf("Error assigning APPLICANT role in Keycloak: %v\n", err)
		// We can continue, but it might cause issues. Let's log it.
	}

	// Create local user with APPLICANT role
	user, err := s.users.CreateApplicantUser(ctx, email, keycloakSub)
	if err != nil {
		fmt.Printf("Error creating local applicant user: %v\n", err)
		_ = s.identity.DeleteUser(ctx, keycloakSub)
		return nil, middleware.NewDatabaseError("create applicant user", err)
	}

	// Mark token as used
	now := time.Now()
	pre.UsedAt = &now
	if err := s.preRegs.Update(ctx, pre); err != nil {
		// non-fatal
	}

	// Log the user in to get JWT tokens
	loginRes, err := s.identity.Login(ctx, email, password)
	if err != nil {
		fmt.Printf("Error logging in after registration: %v\n", err)
		// Account was created but auto-login failed - still a success
		return &AuthResult{
			User: user,
		}, nil
	}

	return &AuthResult{
		AccessToken:  loginRes.AccessToken,
		RefreshToken: loginRes.RefreshToken,
		ExpiresIn:    loginRes.ExpiresIn,
		User:         user,
	}, nil
}

func (s *ProviderRegistrationService) ResendProviderActivation(
	ctx context.Context,
	email string,
) error {
	email = strings.ToLower(strings.TrimSpace(email))

	pre, err := s.preRegs.GetByEmail(ctx, email)
	if err != nil {
		return middleware.NewUnauthorizedError("no pending activation")
	}

	if pre.UsedAt != nil {
		return middleware.NewUnauthorizedError("registration already completed")
	}

	if pre.ResendCount >= 3 &&
		pre.ResendWindowStart != nil &&
		time.Since(*pre.ResendWindowStart) < 30*time.Minute {
		return middleware.NewValidationError("email", "resend limit exceeded")
	}

	newToken := uuid.NewString()
	exp := time.Now().Add(24 * time.Hour)

	pre.Token = newToken
	pre.ExpiresAt = exp

	if pre.ResendWindowStart == nil || time.Since(*pre.ResendWindowStart) > 30*time.Minute {
		now := time.Now()
		pre.ResendWindowStart = &now
		pre.ResendCount = 1
	} else {
		pre.ResendCount++
	}

	if err := s.preRegs.Update(ctx, pre); err != nil {
		return middleware.NewDatabaseError("update prereg", err)
	}

	link := fmt.Sprintf("http://dealna-web-hosting.s3-website-us-east-1.amazonaws.com/provider/activate/index.html?token=%s", newToken)
	return s.email.SendActivationLink(email, link, "provider")
}

func (s *ProviderRegistrationService) GetDocumentUploadURL(
	ctx context.Context,
	userID string,
	documentType string,
	originalFilename string,
	contentType string,
) (string, string, error) {
	// We generate a unique object key using userID and documentType
	// Or we can just use a UUID.

	// Create application if it doesn't exist
	app, err := s.providers.GetProviderApplicationByApplicantID(ctx, userID)
	if err != nil {
		// If no rows, create it
		// We'll insert it with a dummy university for now since we don't have it yet, or the user passes it.
		// Wait, the applicant needs to provide university ID. Let's make university_id optional or we fetch the default one.
		return "", "", middleware.NewDatabaseError("application not found", err)
	}

	objectKey := "providers/" + app.ID + "/documents/" + documentType + "_" + originalFilename

	url, err := s.storage.GeneratePresignedUploadURL(ctx, objectKey, contentType, 15*time.Minute)
	if err != nil {
		return "", "", middleware.NewInternalError(err)
	}

	return url, objectKey, nil
}

func (s *ProviderRegistrationService) StartApplication(
	ctx context.Context,
	userID string,
	universityID string,
	businessName string,
	phoneNumber string,
	businessType string,
	address string,
) (*domain.ProviderApplication, error) {
	// Check if an application already exists
	app, err := s.providers.GetProviderApplicationByApplicantID(ctx, userID)
	if err == nil {
		if app.Status == "DRAFT" {
			// Update the draft with new values
			return s.providers.UpdateProviderApplication(ctx, userID, universityID, businessName, &phoneNumber, &businessType, &address)
		}
		return app, nil
	}

	app, err = s.providers.CreateProviderApplication(
		ctx,
		userID,
		universityID,
		businessName,
		&phoneNumber,
		&businessType,
		&address,
		"DRAFT",
	)
	if err != nil {
		return nil, middleware.NewDatabaseError("create provider application", err)
	}
	return app, nil
}

func (s *ProviderRegistrationService) ConfirmDocumentUpload(
	ctx context.Context,
	userID string,
	objectKey string,
	documentType string,
	originalFilename string,
	contentType string,
	sizeBytes int64,
) error {
	app, err := s.providers.GetProviderApplicationByApplicantID(ctx, userID)
	if err != nil {
		return middleware.NewDatabaseError("application not found", err)
	}

	_, err = s.providers.CreateProviderApplicationDocument(
		ctx,
		app.ID,
		objectKey, // Assuming objectKey is stored as file_path
		documentType,
		originalFilename,
		contentType,
		sizeBytes,
		"UPLOADED",
	)
	if err != nil {
		return middleware.NewDatabaseError("create document record", err)
	}

	return nil
}

func (s *ProviderRegistrationService) SubmitApplication(ctx context.Context, userID string) error {
	app, err := s.providers.GetProviderApplicationByApplicantID(ctx, userID)
	if err != nil {
		return middleware.NewDatabaseError("application not found", err)
	}

	docs, err := s.providers.GetProviderApplicationDocuments(ctx, app.ID)
	if err != nil {
		return middleware.NewDatabaseError("get documents", err)
	}

	hasNationalID := false
	hasProofOfOwnership := false

	for _, doc := range docs {
		if doc.DocumentType != nil {
			if *doc.DocumentType == "NATIONAL_ID" {
				hasNationalID = true
			}
			if *doc.DocumentType == "PROOF_OF_OWNERSHIP" {
				hasProofOfOwnership = true
			}
		}
	}

	if !hasNationalID || !hasProofOfOwnership {
		return middleware.NewValidationError("documents", "NATIONAL_ID and PROOF_OF_OWNERSHIP are both required")
	}

	err = s.providers.UpdateProviderApplicationStatus(ctx, app.ID, "PENDING_REVIEW")
	if err != nil {
		return middleware.NewDatabaseError("update status", err)
	}

	return nil
}

func (s *ProviderRegistrationService) ApproveApplication(ctx context.Context, adminID, applicantID string) error {
	// 1. Get application
	app, err := s.providers.GetProviderApplicationByApplicantID(ctx, applicantID)
	if err != nil {
		return middleware.NewDatabaseError("application not found", err)
	}

	if app.Status != "PENDING_REVIEW" {
		return middleware.NewValidationError("status", "application is not in PENDING_REVIEW state")
	}

	// 2. Update status to APPROVED and log admin
	err = s.providers.UpdateProviderApplicationReview(ctx, app.ID, "APPROVED", adminID, "Approved by admin")
	if err != nil {
		return middleware.NewDatabaseError("update status", err)
	}

	// 3. Create Provider profile in provider table
	_, err = s.providers.CreateProvider(ctx, applicantID, app.BusinessName, nil, nil, nil)
	if err != nil {
		return middleware.NewDatabaseError("create provider", err)
	}

	// 4. Update user role and status in DB
	err = s.users.UpdateUserRole(ctx, applicantID, "PROVIDER")
	if err != nil {
		return middleware.NewDatabaseError("update user role", err)
	}

	err = s.users.UpdateUserStatus(ctx, applicantID, "ACTIVE")
	if err != nil {
		return middleware.NewDatabaseError("update user status", err)
	}

	// 5. Create or update Profile row with Business Name as display name
	existingProfile, _, profileErr := s.users.GetProfile(ctx, applicantID)
	if profileErr != nil || existingProfile == nil {
		// Profile doesn't exist yet — create it
		if createErr := s.users.CreateProfileForUser(ctx, applicantID, app.BusinessName); createErr != nil {
			fmt.Printf("[WARNING] CreateProfileForUser failed: %v\n", createErr)
		}
	} else {
		// Profile exists — update display name to business name
		now := time.Now()
		s.users.UpdateProfile(ctx, applicantID, &app.BusinessName, &existingProfile.Bio, &existingProfile.ProfilePictureURL, &now)
	}

	// 6. Update user role in Keycloak
	user, err := s.users.GetByID(ctx, applicantID)
	if err != nil {
		return middleware.NewDatabaseError("get user", err)
	}

	err = s.identity.AssignRoleToUser(ctx, user.KeycloakSub, "PROVIDER")
	if err != nil {
		// Log the error but don't fail the approval process
		fmt.Printf("[WARNING] AssignRoleToUser failed: %v\n", err)
	}

	// Send notification
	_ = s.notifs.CreateNotification(ctx, applicantID, domain.NotifTypeApplicationApproved, NotificationContext{
		ActingUserID: &adminID,
	})

	return nil
}

func (s *ProviderRegistrationService) RejectApplication(ctx context.Context, adminID, applicantID, comment string) error {
	app, err := s.providers.GetProviderApplicationByApplicantID(ctx, applicantID)
	if err != nil {
		return middleware.NewDatabaseError("application not found", err)
	}

	if app.Status != "PENDING_REVIEW" {
		return middleware.NewValidationError("status", "application is not in PENDING_REVIEW state")
	}

	err = s.providers.UpdateProviderApplicationReview(ctx, app.ID, "REJECTED", adminID, comment)
	if err != nil {
		return middleware.NewDatabaseError("update status", err)
	}

	// Send notification
	_ = s.notifs.CreateNotification(ctx, applicantID, domain.NotifTypeApplicationRejected, NotificationContext{
		ActingUserID: &adminID,
	})

	return nil
}

func (s *ProviderRegistrationService) GetApplicationStatus(ctx context.Context, applicantID string) (*domain.ProviderApplication, error) {
	app, err := s.providers.GetProviderApplicationByApplicantID(ctx, applicantID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, middleware.NewUserNotFoundError("Application not found")
		}
		return nil, middleware.NewDatabaseError("application not found", err)
	}
	return app, nil
}
