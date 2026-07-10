package services

import (
	"context"

	"time"

	"github.com/google/uuid"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type AdminService struct {
	adminRepo ports.IAdminRepository
	storage   ports.IStorageProvider
	email     ports.IEmailService
	identity  ports.IIdentityProvider
	logger    middleware.StructuredLoggerInterface
	notifs    *NotificationService
	itemRepo  ports.ItemRepository
}

func NewAdminService(
	adminRepo ports.IAdminRepository,
	storage ports.IStorageProvider,
	email ports.IEmailService,
	identity ports.IIdentityProvider,
	logger middleware.StructuredLoggerInterface,
	notifs *NotificationService,
	itemRepo ports.ItemRepository,
) *AdminService {
	return &AdminService{
		adminRepo: adminRepo,
		storage:   storage,
		email:     email,
		identity:  identity,
		logger:    logger,
		notifs:    notifs,
		itemRepo:  itemRepo,
	}
}

func (s *AdminService) GetDashboardMetrics(ctx context.Context, universityID string) (*domain.DashboardMetrics, error) {
	s.logger.Info(ctx, "Fetching dashboard metrics", map[string]any{"university_id": universityID})
	return s.adminRepo.GetDashboardMetrics(ctx, universityID)
}

func (s *AdminService) GetUsers(ctx context.Context, search string, roleFilter string, statusFilter string, limit int, offset int) ([]domain.AdminUserSnapshot, int, error) {
	s.logger.Info(ctx, "Fetching users for admin", map[string]any{"search": search, "role": roleFilter, "status": statusFilter})
	return s.adminRepo.GetUsers(ctx, search, roleFilter, statusFilter, limit, offset)
}

func (s *AdminService) GetVerifications(ctx context.Context, status string) ([]domain.AdminProviderVerification, error) {
	s.logger.Info(ctx, "Fetching verifications for admin", map[string]any{"status": status})
	return s.adminRepo.GetVerifications(ctx, status)
}

func (s *AdminService) ApproveVerification(ctx context.Context, applicationID string, adminID string) error {
	s.logger.Info(ctx, "Approving verification", map[string]any{"application_id": applicationID, "admin_id": adminID})
	err := s.adminRepo.ApproveVerification(ctx, applicationID, adminID)
	if err == nil {
		if email, emailErr := s.adminRepo.GetApplicantEmail(ctx, applicationID); emailErr == nil && email != "" {
			_ = s.email.SendApplicationStatusEmail(email, "APPROVED", "")
		}
	}
	return err
}

func (s *AdminService) RejectVerification(ctx context.Context, applicationID string, adminID string, comment string) error {
	s.logger.Info(ctx, "Rejecting verification", map[string]any{"application_id": applicationID, "admin_id": adminID, "comment": comment})
	err := s.adminRepo.RejectVerification(ctx, applicationID, adminID, comment)
	if err == nil {
		if email, emailErr := s.adminRepo.GetApplicantEmail(ctx, applicationID); emailErr == nil && email != "" {
			_ = s.email.SendApplicationStatusEmail(email, "REJECTED", comment)
		}
	}
	return err
}

func (s *AdminService) WarnUser(ctx context.Context, adminID string, targetUserID string, reason string) error {
	s.logger.Info(ctx, "Warning user", map[string]any{"target_user_id": targetUserID, "admin_id": adminID, "reason": reason})
	count, err := s.adminRepo.WarnUser(ctx, adminID, targetUserID, reason)
	if err != nil {
		return err
	}

	if s.notifs != nil {
		_ = s.notifs.CreateNotification(ctx, targetUserID, domain.NotifTypeAdminWarning, NotificationContext{
			ActingUserID: &adminID,
			Reason:       &reason,
		})
	}

	if count >= 3 {
		return s.BanUser(ctx, adminID, targetUserID, "Exceeded maximum warnings")
	}
	return nil
}

func (s *AdminService) WarnItemOwner(ctx context.Context, adminID string, itemIDStr string, reason string) error {
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		return err
	}
	
	detail, err := s.itemRepo.GetItemDetail(ctx, itemID)
	if err != nil {
		return err
	}
	
	specificReason := "A report was filed against your item '" + detail.Title + "'. Please note that accumulating 3 warnings will result in an automatic account ban."
	
	return s.WarnUser(ctx, adminID, detail.OwnerID.String(), specificReason)
}

func (s *AdminService) BanUser(ctx context.Context, adminID string, targetUserID string, reason string) error {
	s.logger.Info(ctx, "Banning user (downgrade to limited-student)", map[string]any{"target_user_id": targetUserID, "admin_id": adminID, "reason": reason})

	// 1. Update DB role to LIMITED_STUDENT
	err := s.adminRepo.BanUser(ctx, targetUserID)
	if err != nil {
		return err
	}

	// 2. Fetch Keycloak Sub
	sub, err := s.adminRepo.GetKeycloakSub(ctx, targetUserID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get keycloak sub for banned user", map[string]any{"user_id": targetUserID, "error": err.Error()})
		return err // Or return nil if we want to proceed without keycloak sync
	}

	if sub != "" && s.identity != nil {
		// 3. Update Keycloak Roles
		_ = s.identity.RemoveRoleFromUser(ctx, sub, "verified-student")
		_ = s.identity.RemoveRoleFromUser(ctx, sub, "provider")
		err = s.identity.AssignRoleToUser(ctx, sub, "limited-student")
		if err != nil {
			s.logger.Error(ctx, "Failed to assign limited-student role in Keycloak", map[string]any{"user_id": targetUserID, "error": err.Error()})
		}
	}

	if s.notifs != nil {
		_ = s.notifs.CreateNotification(ctx, targetUserID, domain.NotifTypeAdminBan, NotificationContext{
			ActingUserID: &adminID,
			Reason:       &reason,
		})
	}

	return nil
}

func (s *AdminService) GetVerificationDocuments(ctx context.Context, applicationID string) ([]domain.AdminProviderDocument, error) {
	s.logger.Info(ctx, "Fetching verification documents", map[string]any{"application_id": applicationID})
	docs, err := s.adminRepo.GetVerificationDocuments(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	for i, doc := range docs {
		if doc.FilePath != "" {
			url, err := s.storage.GeneratePresignedDownloadURL(ctx, doc.FilePath, 1*time.Hour)
			if err == nil {
				docs[i].FilePath = url
			}
		}
	}
	return docs, nil
}

func (s *AdminService) GetAdminUserProfileStats(ctx context.Context, userID string) (int, int, int, error) {
	s.logger.Info(ctx, "Fetching admin user profile stats", map[string]any{"user_id": userID})
	return s.adminRepo.GetAdminUserProfileStats(ctx, userID)
}

func (s *AdminService) DeleteItemAdmin(ctx context.Context, adminID string, itemIDStr string, reason string) error {
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		return err
	}

	detail, err := s.itemRepo.GetItemDetail(ctx, itemID)
	if err != nil {
		return err
	}

	s.logger.Info(ctx, "Admin deleting item", map[string]any{"item_id": itemIDStr, "admin_id": adminID, "reason": reason})

	err = s.itemRepo.SoftDeleteItem(ctx, itemID)
	if err != nil {
		return err
	}

	if s.notifs != nil && detail != nil {
		_ = s.notifs.CreateNotification(ctx, detail.OwnerID.String(), domain.NotifTypeItemDeleted, NotificationContext{
			ActingUserID: &adminID,
			Reason:       &reason,
			ItemID:       &itemIDStr,
		})
	}

	return nil
}

