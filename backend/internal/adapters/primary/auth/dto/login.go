package dto

// LoginRequest login payload
type LoginRequest struct {
	Email    string `json:"email" example:"user@bzu.edu"`
	Password string `json:"password" example:"secret123"`
}

// LoginResponse JWT response
type LoginResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
