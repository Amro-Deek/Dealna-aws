package persistence

import (
	"context"
	"fmt"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepository struct {
	db *pgxpool.Pool
}

func NewAdminRepository(db *pgxpool.Pool) ports.IAdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) GetDashboardMetrics(ctx context.Context, universityID string) (*domain.DashboardMetrics, error) {
	var metrics domain.DashboardMetrics

	// Users query
	err := r.db.QueryRow(ctx, `
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE role = 'STUDENT'),
			COUNT(*) FILTER (WHERE role = 'PROVIDER')
		FROM "User" 
		WHERE account_status = 'ACTIVE'
	`).Scan(&metrics.TotalActiveUsers, &metrics.TotalStudents, &metrics.TotalVerifiedProviders)
	if err != nil {
		return nil, err
	}

	// Items query
	err = r.db.QueryRow(ctx, `
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE c.name ILIKE '%product%'),
			COUNT(*) FILTER (WHERE c.name ILIKE '%apartment%'),
			COUNT(*) FILTER (WHERE c.name ILIKE '%textbook%')
		FROM item i
		LEFT JOIN category c ON i.category_id = c.category_id
		WHERE i.item_status = 'AVAILABLE'
	`).Scan(&metrics.ActiveListings, &metrics.ProductsListings, &metrics.ApartmentsListings, &metrics.TextbooksListings)
	if err != nil {
		return nil, err
	}

	// Transactions query
	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(i.price), 0)
		FROM transaction t
		JOIN item i ON t.item_id = i.item_id
		WHERE t.created_at >= NOW() - INTERVAL '1 day'
		AND t.transaction_status = 'COMPLETED'
	`).Scan(&metrics.DailyTradeVolume)
	if err != nil {
		metrics.DailyTradeVolume = 0 // default on error
	}

	// Fraud flags query
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM report
		WHERE created_at >= NOW() - INTERVAL '1 day'
	`).Scan(&metrics.FraudFlags24h)
	if err != nil {
		metrics.FraudFlags24h = 0
	}

	return &metrics, nil
}

func (r *AdminRepository) GetUsers(ctx context.Context, search string, roleFilter string, statusFilter string, limit int, offset int) ([]domain.AdminUserSnapshot, int, error) {
	query := `FROM "User" WHERE 1=1`
	var args []interface{}
	argID := 1

	if search != "" {
		query += fmt.Sprintf(` AND (email ILIKE $%d OR username ILIKE $%d)`, argID, argID)
		args = append(args, "%"+search+"%")
		argID++
	}
	if roleFilter != "" && roleFilter != "ALL" {
		query += fmt.Sprintf(` AND role = $%d`, argID)
		args = append(args, roleFilter)
		argID++
	}
	if statusFilter != "" && statusFilter != "ALL" {
		query += fmt.Sprintf(` AND account_status = $%d`, argID)
		args = append(args, statusFilter)
		argID++
	}

	var total int
	countQuery := `SELECT COUNT(*) ` + query
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := `SELECT user_id, email, role, account_status, created_at, '' ` + query + fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argID, argID+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []domain.AdminUserSnapshot
	for rows.Next() {
		var u domain.AdminUserSnapshot
		err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.Status, &u.JoinedAt, &u.AvatarURL)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	if users == nil {
		users = []domain.AdminUserSnapshot{}
	}
	return users, total, nil
}

func (r *AdminRepository) GetVerifications(ctx context.Context, status string) ([]domain.AdminProviderVerification, error) {
	rows, err := r.db.Query(ctx, `
		SELECT application_id, business_name, COALESCE(business_type, 'General'), 'Business Documents', status, submitted_at
		FROM providerapplication
		ORDER BY submitted_at DESC
		LIMIT 50
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var verifications []domain.AdminProviderVerification
	for rows.Next() {
		var v domain.AdminProviderVerification
		err := rows.Scan(&v.ID, &v.ProviderName, &v.Type, &v.ProofTypes, &v.Status, &v.SubmittedAt)
		if err != nil {
			return nil, err
		}
		verifications = append(verifications, v)
	}
	if verifications == nil {
		verifications = []domain.AdminProviderVerification{}
	}
	return verifications, nil
}

func (r *AdminRepository) ApproveVerification(ctx context.Context, applicationID string, adminID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Fetch applicant details. Note: applicant_id in providerapplication IS the User.user_id
	// (the providerapplicant table is already cleaned up during registration completion)
	var userID, businessName string
	var phone, businessType, address *string
	err = tx.QueryRow(ctx, `
		SELECT applicant_id, business_name, phone_number, business_type, address 
		FROM providerapplication WHERE application_id = $1
	`, applicationID).Scan(&userID, &businessName, &phone, &businessType, &address)
	if err != nil {
		return err
	}

	// Update application status to APPROVED
	_, err = tx.Exec(ctx, `
		UPDATE providerapplication 
		SET status = 'APPROVED', reviewed_at = NOW(), reviewed_by_admin_id = NULLIF($1, '')::uuid 
		WHERE application_id = $2
	`, adminID, applicationID)
	if err != nil {
		return err
	}

	// Insert into provider table (userID == User.user_id)
	_, err = tx.Exec(ctx, `
		INSERT INTO provider (user_id, business_name, phone_number, business_type, address, verified_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			business_name = EXCLUDED.business_name,
			phone_number = EXCLUDED.phone_number,
			business_type = EXCLUDED.business_type,
			address = EXCLUDED.address,
			verified_at = NOW()
	`, userID, businessName, phone, businessType, address)
	if err != nil {
		return err
	}

	// Update user role to PROVIDER
	_, err = tx.Exec(ctx, `
		UPDATE "User" SET role = 'PROVIDER' WHERE user_id = $1
	`, userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *AdminRepository) RejectVerification(ctx context.Context, applicationID string, adminID string, comment string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE providerapplication 
		SET status = 'REJECTED', reviewed_at = NOW(), reviewed_by_admin_id = NULLIF($1, '')::uuid, admin_comment = $2 
		WHERE application_id = $3
	`, adminID, comment, applicationID)
	return err
}

func (r *AdminRepository) WarnUser(ctx context.Context, adminID string, targetUserID string, reason string) (int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Insert warning
	_, err = tx.Exec(ctx, `
		INSERT INTO public.user_warnings (user_id, admin_id, reason)
		VALUES ($1::uuid, $2::uuid, $3)
	`, targetUserID, adminID, reason)
	if err != nil {
		return 0, err
	}

	// Count warnings
	var count int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM public.user_warnings WHERE user_id = $1::uuid
	`, targetUserID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, tx.Commit(ctx)
}

func (r *AdminRepository) BanUser(ctx context.Context, targetUserID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE "User" SET role = 'LIMITED_STUDENT' WHERE user_id = $1
	`, targetUserID)
	return err
}

func (r *AdminRepository) GetVerificationDocuments(ctx context.Context, applicationID string) ([]domain.AdminProviderDocument, error) {
	rows, err := r.db.Query(ctx, `
		SELECT document_id, file_path, COALESCE(document_type, ''), COALESCE(original_filename, ''), COALESCE(content_type, '')
		FROM providerapplicationdocument
		WHERE application_id = $1
		ORDER BY uploaded_at ASC
	`, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []domain.AdminProviderDocument
	for rows.Next() {
		var doc domain.AdminProviderDocument
		if err := rows.Scan(&doc.ID, &doc.FilePath, &doc.DocumentType, &doc.OriginalFilename, &doc.ContentType); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	if docs == nil {
		docs = []domain.AdminProviderDocument{}
	}
	return docs, nil
}

func (r *AdminRepository) GetApplicantEmail(ctx context.Context, applicationID string) (string, error) {
	var email string
	err := r.db.QueryRow(ctx, `
		SELECT u.email 
		FROM providerapplication p
		JOIN "User" u ON p.applicant_id = u.user_id
		WHERE p.application_id = $1
	`, applicationID).Scan(&email)
	return email, err
}

func (r *AdminRepository) GetKeycloakSub(ctx context.Context, userID string) (string, error) {
	var sub string
	err := r.db.QueryRow(ctx, `
		SELECT keycloak_sub FROM "User" WHERE user_id = $1
	`, userID).Scan(&sub)
	return sub, err
}

func (r *AdminRepository) GetAdminUserProfileStats(ctx context.Context, userID string) (int, int, int, error) {
	var reportsReceived, warningsReceived, totalPosts int
	err := r.db.QueryRow(ctx, `
		SELECT 
			(SELECT COUNT(*) FROM public.reports WHERE 
				(reported_entity_id = $1::uuid AND entity_type = 'USER') 
				OR 
				(reported_entity_id IN (SELECT item_id FROM public.item WHERE owner_id = $1::uuid) AND entity_type = 'ITEM')
			)::int AS reports_received,
			(SELECT COUNT(*) FROM public.user_warnings WHERE user_id = $1::uuid)::int AS warnings_received,
			(SELECT COUNT(*) FROM public.item WHERE owner_id = $1::uuid)::int AS total_posts
	`, userID).Scan(&reportsReceived, &warningsReceived, &totalPosts)
	if err != nil {
		return 0, 0, 0, err
	}
	return reportsReceived, warningsReceived, totalPosts, nil
}
