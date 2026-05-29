package domain

import "time"

type ProviderApplication struct {
	ID                 string     `json:"id"`
	ApplicantID        string     `json:"applicant_id"`
	UniversityID       string     `json:"university_id"`
	BusinessName       string     `json:"business_name"`
	PhoneNumber        *string    `json:"phone_number"`
	BusinessType       *string    `json:"business_type"`
	Address            *string    `json:"address"`
	Status             string     `json:"status"`
	SubmittedAt        time.Time  `json:"submitted_at"`
	ReviewedAt         *time.Time `json:"reviewed_at"`
	AdminComment       *string    `json:"admin_comment"`
	ReviewedByAdminID  *string    `json:"reviewed_by_admin_id"`
}

type ProviderApplicationDocument struct {
	ID               string
	ApplicationID    string
	FilePath         string
	DocumentType     *string
	OriginalFilename *string
	ContentType      *string
	SizeBytes        *int64
	UploadStatus     *string
	UploadedAt       time.Time
}

type Provider struct {
	UserID       string
	BusinessName string
	PhoneNumber  *string
	BusinessType *string
	Address      *string
	VerifiedAt   *time.Time
}
