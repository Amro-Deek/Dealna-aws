package postgres

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProviderRepository struct {
	db *pgxpool.Pool
	q  *generated.Queries
}

func NewProviderRepository(db *pgxpool.Pool) ports.IProviderRepository {
	return &ProviderRepository{
		db: db,
		q:  generated.New(db),
	}
}

func (r *ProviderRepository) CreateProviderApplication(ctx context.Context, applicantID, universityID, businessName string, phoneNumber, businessType, address *string, status string) (*domain.ProviderApplication, error) {
	row, err := r.q.CreateProviderApplication(ctx, generated.CreateProviderApplicationParams{
		ApplicantID:  toUUID(applicantID),
		UniversityID: toUUID(universityID),
		BusinessName: businessName,
		PhoneNumber:  toNullableText(phoneNumber),
		BusinessType: toNullableText(businessType),
		Address:      toNullableText(address),
		Status:       status,
	})
	if err != nil {
		return nil, err
	}
	return &domain.ProviderApplication{
		ID:           row.ApplicationID.String(),
		ApplicantID:  row.ApplicantID.String(),
		UniversityID: row.UniversityID.String(),
		BusinessName: row.BusinessName,
		Status:       row.Status,
	}, nil
}

func (r *ProviderRepository) GetProviderApplicationByApplicantID(ctx context.Context, applicantID string) (*domain.ProviderApplication, error) {
	row, err := r.q.GetProviderApplicationByApplicantID(ctx, toUUID(applicantID))
	if err != nil {
		return nil, err
	}
	return &domain.ProviderApplication{
		ID:           row.ApplicationID.String(),
		ApplicantID:  row.ApplicantID.String(),
		UniversityID: row.UniversityID.String(),
		BusinessName: row.BusinessName,
		Status:       row.Status,
		AdminComment: ptrFromNullableText(row.AdminComment),
	}, nil
}

func (r *ProviderRepository) UpdateProviderApplicationStatus(ctx context.Context, applicationID, status string) error {
	return r.q.UpdateProviderApplicationStatus(ctx, generated.UpdateProviderApplicationStatusParams{
		ApplicationID: toUUID(applicationID),
		Status:        status,
	})
}

func (r *ProviderRepository) UpdateProviderApplication(ctx context.Context, applicantID, universityID, businessName string, phoneNumber, businessType, address *string) (*domain.ProviderApplication, error) {
	row, err := r.q.UpdateProviderApplication(ctx, generated.UpdateProviderApplicationParams{
		ApplicantID:  toUUID(applicantID),
		UniversityID: toUUID(universityID),
		BusinessName: businessName,
		PhoneNumber:  toNullableText(phoneNumber),
		BusinessType: toNullableText(businessType),
		Address:      toNullableText(address),
	})
	if err != nil {
		return nil, err
	}
	return &domain.ProviderApplication{
		ID:           row.ApplicationID.String(),
		ApplicantID:  row.ApplicantID.String(),
		UniversityID: row.UniversityID.String(),
		BusinessName: row.BusinessName,
		Status:       row.Status,
	}, nil
}

func (r *ProviderRepository) UpdateProviderApplicationReview(ctx context.Context, applicationID, status, adminID, comment string) error {
	return r.q.UpdateProviderApplicationReview(ctx, generated.UpdateProviderApplicationReviewParams{
		ApplicationID:     toUUID(applicationID),
		Status:            status,
		ReviewedByAdminID: toUUID(adminID),
		AdminComment:      toNullableText(&comment),
	})
}

func (r *ProviderRepository) CreateProviderApplicationDocument(ctx context.Context, applicationID, filePath, documentType, originalFilename, contentType string, sizeBytes int64, uploadStatus string) (*domain.ProviderApplicationDocument, error) {
	// Remove any previously uploaded document of the same type for this application to avoid duplicates
	if documentType != "" {
		_, _ = r.db.Exec(ctx, "DELETE FROM providerapplicationdocument WHERE application_id = $1 AND document_type = $2", toUUID(applicationID), documentType)
	}

	row, err := r.q.CreateProviderApplicationDocument(ctx, generated.CreateProviderApplicationDocumentParams{
		ApplicationID:    toUUID(applicationID),
		FilePath:         filePath,
		DocumentType:     toNullableText(&documentType),
		OriginalFilename: toNullableText(&originalFilename),
		ContentType:      toNullableText(&contentType),
		SizeBytes:        toNullableInt8(&sizeBytes),
		UploadStatus:     toNullableText(&uploadStatus),
	})
	if err != nil {
		return nil, err
	}
	return &domain.ProviderApplicationDocument{
		ID:            row.DocumentID.String(),
		ApplicationID: row.ApplicationID.String(),
		FilePath:      row.FilePath,
	}, nil
}

func (r *ProviderRepository) GetProviderApplicationDocuments(ctx context.Context, applicationID string) ([]domain.ProviderApplicationDocument, error) {
	rows, err := r.q.GetProviderApplicationDocuments(ctx, toUUID(applicationID))
	if err != nil {
		return nil, err
	}
	var docs []domain.ProviderApplicationDocument
	for _, row := range rows {
		docs = append(docs, domain.ProviderApplicationDocument{
			ID:               row.DocumentID.String(),
			ApplicationID:    row.ApplicationID.String(),
			FilePath:         row.FilePath,
			DocumentType:     ptrFromNullableText(row.DocumentType),
			OriginalFilename: ptrFromNullableText(row.OriginalFilename),
			ContentType:      ptrFromNullableText(row.ContentType),
			SizeBytes:        ptrFromNullableInt8(row.SizeBytes),
			UploadStatus:     ptrFromNullableText(row.UploadStatus),
		})
	}
	return docs, nil
}

func (r *ProviderRepository) CreateProvider(ctx context.Context, userID, businessName string, phoneNumber, businessType, address *string) (*domain.Provider, error) {
	row, err := r.q.CreateProvider(ctx, generated.CreateProviderParams{
		UserID:       toUUID(userID),
		BusinessName: businessName,
		PhoneNumber:  toNullableText(phoneNumber),
		BusinessType: toNullableText(businessType),
		Address:      toNullableText(address),
	})
	if err != nil {
		return nil, err
	}
	return &domain.Provider{
		UserID:       row.UserID.String(),
		BusinessName: row.BusinessName,
	}, nil
}
