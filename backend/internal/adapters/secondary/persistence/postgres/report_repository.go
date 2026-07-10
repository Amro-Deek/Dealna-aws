package postgres

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
	"github.com/jackc/pgx/v5/pgtype"
)

type ReportRepository struct {
	queries *generated.Queries
}

func NewReportRepository(queries *generated.Queries) ports.IReportRepository {
	return &ReportRepository{queries: queries}
}

func (r *ReportRepository) CreateReport(ctx context.Context, report *domain.Report) (*domain.Report, error) {
	var reportedEntityID pgtype.UUID
	_ = reportedEntityID.Scan(report.ReportedEntityID)

	var reporterID pgtype.UUID
	_ = reporterID.Scan(report.ReporterID)

	res, err := r.queries.CreateReport(ctx, generated.CreateReportParams{
		ReporterID:       reporterID,
		ReportedEntityID: reportedEntityID,
		EntityType:       generated.ReportEntityType(report.EntityType),
		Type:             generated.ReportType(report.Type),
		Description:      pgtype.Text{String: report.Description, Valid: true},
		AttachmentUrl:    pgtype.Text{String: report.AttachmentURL, Valid: report.AttachmentURL != ""},
	})
	if err != nil {
		return nil, err
	}

	return mapToDomainReport(res), nil
}

func (r *ReportRepository) GetReport(ctx context.Context, id string) (*domain.Report, error) {
	var uuid pgtype.UUID
	_ = uuid.Scan(id)

	res, err := r.queries.GetReport(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return mapToDomainReport(res), nil
}

func (r *ReportRepository) ListReports(ctx context.Context, limit, offset int32) ([]domain.Report, error) {
	reports, err := r.queries.ListReports(ctx, generated.ListReportsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	domainReports := make([]domain.Report, len(reports))
	for i, row := range reports {
		// Create a domain report from the base report data
		domainRep := mapToDomainReport(generated.Reports{
			ID:               row.ID,
			ReporterID:       row.ReporterID,
			ReportedEntityID: row.ReportedEntityID,
			EntityType:       row.EntityType,
			Type:             row.Type,
			Description:      row.Description,
			AttachmentUrl:    row.AttachmentUrl,
			Status:           row.Status,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
		})
		
		// Attach joined names
		if row.ReporterName.Valid {
			domainRep.ReporterName = row.ReporterName.String
		}
		if row.ReportedEntityName != "" {
			domainRep.ReportedEntityName = row.ReportedEntityName
		}
		
		domainReports[i] = *domainRep
	}

	return domainReports, nil
}

func (r *ReportRepository) UpdateReportStatus(ctx context.Context, id string, status domain.ReportStatus) (*domain.Report, error) {
	var uuid pgtype.UUID
	_ = uuid.Scan(id)

	res, err := r.queries.UpdateReportStatus(ctx, generated.UpdateReportStatusParams{
		ID:     uuid,
		Status: generated.ReportStatus(status),
	})
	if err != nil {
		return nil, err
	}

	return mapToDomainReport(res), nil
}

func mapToDomainReport(r generated.Reports) *domain.Report {
	// Workaround for pgtype to string if value() returns [16]byte
	// Just use custom helper since Value() returns driver.Value which is usually a string for UUIDs in pgx v5, but sometimes [16]byte.
	return &domain.Report{
		ID:               pgtypeUUIDToStr(r.ID),
		ReporterID:       pgtypeUUIDToStr(r.ReporterID),
		ReportedEntityID: pgtypeUUIDToStr(r.ReportedEntityID),
		EntityType:       domain.ReportEntityType(r.EntityType),
		Type:             domain.ReportType(r.Type),
		Description:      r.Description.String,
		AttachmentURL:    r.AttachmentUrl.String,
		Status:           domain.ReportStatus(r.Status),
		CreatedAt:        r.CreatedAt.Time,
		UpdatedAt:        r.UpdatedAt.Time,
	}
}

func pgtypeUUIDToStr(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	// UUID formatting
	b := u.Bytes
	return string([]byte{
		hex[b[0]>>4], hex[b[0]&0x0f],
		hex[b[1]>>4], hex[b[1]&0x0f],
		hex[b[2]>>4], hex[b[2]&0x0f],
		hex[b[3]>>4], hex[b[3]&0x0f],
		'-',
		hex[b[4]>>4], hex[b[4]&0x0f],
		hex[b[5]>>4], hex[b[5]&0x0f],
		'-',
		hex[b[6]>>4], hex[b[6]&0x0f],
		hex[b[7]>>4], hex[b[7]&0x0f],
		'-',
		hex[b[8]>>4], hex[b[8]&0x0f],
		hex[b[9]>>4], hex[b[9]&0x0f],
		'-',
		hex[b[10]>>4], hex[b[10]&0x0f],
		hex[b[11]>>4], hex[b[11]&0x0f],
		hex[b[12]>>4], hex[b[12]&0x0f],
		hex[b[13]>>4], hex[b[13]&0x0f],
		hex[b[14]>>4], hex[b[14]&0x0f],
		hex[b[15]>>4], hex[b[15]&0x0f],
	})
}

const hex = "0123456789abcdef"
