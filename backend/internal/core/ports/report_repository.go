package ports

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type IReportRepository interface {
	CreateReport(ctx context.Context, report *domain.Report) (*domain.Report, error)
	GetReport(ctx context.Context, id string) (*domain.Report, error)
	ListReports(ctx context.Context, limit, offset int32) ([]domain.Report, error)
	UpdateReportStatus(ctx context.Context, id string, status domain.ReportStatus) (*domain.Report, error)
}

type IReportService interface {
	CreateReport(ctx context.Context, reporterID string, reportedEntityID string, entityType domain.ReportEntityType, reportType domain.ReportType, description string, attachmentURL string) (*domain.Report, error)
	GetReport(ctx context.Context, id string) (*domain.Report, error)
	ListReports(ctx context.Context, limit, offset int32) ([]domain.Report, error)
	ResolveReport(ctx context.Context, id string, adminID string, status domain.ReportStatus) (*domain.Report, error)
}
