package models

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Invalid credentials"`
	Error   string `json:"error" example:"Authentication failed"`
}
