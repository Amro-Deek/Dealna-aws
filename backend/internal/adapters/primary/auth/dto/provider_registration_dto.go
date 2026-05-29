package dto

type RequestProviderRegistrationRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type StartProviderApplicationRequest struct {
	UniversityID string `json:"university_id" binding:"required"`
	BusinessName string `json:"business_name" binding:"required"`
	PhoneNumber  string `json:"phone_number" binding:"required"`
	BusinessType string `json:"business_type" binding:"required"`
	Address      string `json:"address" binding:"required"`
}

type GetDocumentUploadURLRequest struct {
	DocumentType     string `json:"document_type" binding:"required"`
	OriginalFilename string `json:"original_filename" binding:"required"`
	ContentType      string `json:"content_type" binding:"required"`
}

type ConfirmDocumentUploadRequest struct {
	ObjectKey        string `json:"object_key" binding:"required"`
	DocumentType     string `json:"document_type" binding:"required"`
	OriginalFilename string `json:"original_filename" binding:"required"`
	ContentType      string `json:"content_type" binding:"required"`
	SizeBytes        int64  `json:"size_bytes" binding:"required"`
}
