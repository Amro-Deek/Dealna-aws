package models

// LoginResponse represents login success payload
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6..."`
}

type MeResponse struct {
	UserID string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role   string `json:"role" example:"Verified Student"`
}
