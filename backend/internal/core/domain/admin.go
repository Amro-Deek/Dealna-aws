package domain

import "time"

type DashboardMetrics struct {
	TotalActiveUsers       int     `json:"total_active_users"`
	TotalStudents          int     `json:"total_students"`
	TotalVerifiedProviders int     `json:"total_verified_providers"`
	ActiveListings         int     `json:"active_listings"`
	ProductsListings       int     `json:"products_listings"`
	ApartmentsListings     int     `json:"apartments_listings"`
	TextbooksListings      int     `json:"textbooks_listings"`
	DailyTradeVolume       float64 `json:"daily_trade_volume"`
	FraudFlags24h          int     `json:"fraud_flags_24h"`
}

type AdminUserSnapshot struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Status    string    `json:"status"` // Verified, Pending, Flagged
	JoinedAt  time.Time `json:"joined_at"`
	AvatarURL string    `json:"avatar_url"`
}

type AdminProviderVerification struct {
	ID           string    `json:"id"`
	ProviderName string    `json:"provider_name"`
	Type         string    `json:"type"` // e.g. Non-student Trust
	ProofTypes   string    `json:"proof_types"`
	Status       string    `json:"status"`
	SubmittedAt  time.Time `json:"submitted_at"`
}

type AdminProviderDocument struct {
	ID               string `json:"id"`
	FilePath         string `json:"file_path"`
	DocumentType     string `json:"document_type"`
	OriginalFilename string `json:"original_filename"`
	ContentType      string `json:"content_type"`
}

type AdminUserProfile struct {
	User             Profile `json:"user"`
	ReportsReceived  int     `json:"reports_received"`
	WarningsReceived int     `json:"warnings_received"`
	TotalPosts       int     `json:"total_posts"`
	Items            []Item  `json:"items"`
}
