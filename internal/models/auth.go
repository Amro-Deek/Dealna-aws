package models

// LoginResponse represents login success payload
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6..."`
}
