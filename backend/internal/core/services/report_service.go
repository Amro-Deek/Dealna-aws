package services

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type ReportService struct {
	reportRepo ports.IReportRepository
	logger     middleware.StructuredLoggerInterface
}

func NewReportService(reportRepo ports.IReportRepository, logger middleware.StructuredLoggerInterface) ports.IReportService {
	return &ReportService{
		reportRepo: reportRepo,
		logger:     logger,
	}
}

func (s *ReportService) CreateReport(ctx context.Context, reporterID string, reportedEntityID string, entityType domain.ReportEntityType, reportType domain.ReportType, description string, attachmentURL string) (*domain.Report, error) {
	s.logger.Info(ctx, "Creating report", map[string]any{"reporter_id": reporterID, "entity_type": entityType, "reported_entity_id": reportedEntityID})

	report := &domain.Report{
		ReporterID:       reporterID,
		ReportedEntityID: reportedEntityID,
		EntityType:       entityType,
		Type:             reportType,
		Description:      description,
		AttachmentURL:    attachmentURL,
	}

	return s.reportRepo.CreateReport(ctx, report)
}

func (s *ReportService) GetReport(ctx context.Context, id string) (*domain.Report, error) {
	return s.reportRepo.GetReport(ctx, id)
}

func (s *ReportService) ListReports(ctx context.Context, limit, offset int32) ([]domain.Report, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.reportRepo.ListReports(ctx, limit, offset)
}

func (s *ReportService) ResolveReport(ctx context.Context, id string, adminID string, status domain.ReportStatus) (*domain.Report, error) {
	s.logger.Info(ctx, "Resolving report", map[string]any{"report_id": id, "admin_id": adminID, "status": status})
	return s.reportRepo.UpdateReportStatus(ctx, id, status)
}
