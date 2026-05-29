package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"time"
)

type ProviderRegistrationService struct {
	users     ports.IUserRepository
	providers ports.IProviderRepository
	email     ports.IEmailService
	identity  ports.IIdentityProvider
	storage   ports.IStorageProvider
	notifs    *NotificationService
}

func NewProviderRegistrationService(
	users ports.IUserRepository,
	providers ports.IProviderRepository,
	email ports.IEmailService,
	identity ports.IIdentityProvider,
	storage ports.IStorageProvider,
	notifs *NotificationService,
) *ProviderRegistrationService {
	return &ProviderRegistrationService{
		users:     users,
		providers: providers,
		email:     email,
		identity:  identity,
		storage:   storage,
		notifs:    notifs,
	}
}

func (s *ProviderRegistrationService) RequestProviderRegistration(
	ctx context.Context,
	email string,
	password string,
) error {
	email = strings.ToLower(strings.TrimSpace(email))

	// 1. Check if user already exists locally
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return middleware.NewEmailAlreadyUsedError(email)
	}

	nameParts := strings.Split(email, "@")
	firstName := nameParts[0]

	// 2. Register in Keycloak (emailVerified=false)
	keycloakSub, err := s.identity.RegisterUser(ctx, email, password, firstName, firstName)
	if err != nil {
		return err
	}

	// 3. Create local user with APPLICANT role
	user, err := s.users.CreateApplicantUser(ctx, email, keycloakSub)
	if err != nil {
		_ = s.identity.DeleteUser(ctx, keycloakSub)
		return middleware.NewDatabaseError("create applicant user", err)
	}

	// 4. Create providerapplication with status EMAIL_UNVERIFIED
	// We need a dummy university ID for now, or just use the BZU university ID
	// Let's assume the user selects university in Step 3, but the schema requires university_id NOT NULL in providerapplication!
	// Wait, we can either make university_id nullable or pass a default one. Let's make it nullable or provide a default.
	// Actually, wait, does schema.sql require university_id in providerapplication?
	// Yes: university_id uuid NOT NULL.
	// That's a problem if we don't know the university_id at step 1.
	
	// We can fetch the Birzeit University ID since it's the only one right now
	// Or we can just trigger Keycloak's VERIFY_EMAIL and wait until they login to create the providerapplication!
	// YES! We don't even need to create providerapplication in Step 1.
	// They just become an APPLICANT User.
	// Step 2 is native Keycloak email verification.
	// Step 3 (Submit details) is where we create the providerapplication.

	// Trigger Keycloak verification email
	err = s.identity.ExecuteActionsEmail(ctx, keycloakSub, []string{"VERIFY_EMAIL"})
	if err != nil {
		// Log error, but user can still request it later
	}

	_ = user
	return nil
}

func (s *ProviderRegistrationService) GetDocumentUploadURL(
	ctx context.Context,
	userID string,
	documentType string,
	originalFilename string,
	contentType string,
) (string, string, error) {
	// Use S3 provider to get a presigned URL
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

	// 3. Create Provider profile
	_, err = s.providers.CreateProvider(ctx, applicantID, app.BusinessName, nil, nil, nil)
	if err != nil {
		return middleware.NewDatabaseError("create provider", err)
	}

	// 4. Update user role in DB
	err = s.users.UpdateUserRole(ctx, applicantID, "PROVIDER")
	if err != nil {
		return middleware.NewDatabaseError("update user role", err)
	}

	// 5. Update user role in Keycloak
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
		return nil, middleware.NewDatabaseError("application not found", err)
	}
	return app, nil
}

