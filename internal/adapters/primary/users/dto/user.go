package dto

// MeResponse represents current user profile
// @Description Current authenticated user
type MeResponse struct {
	ID    string `json:"id" example:"b3c2f9a1"`
	Email string `json:"email" example:"user@bzu.edu"`
	Role  string `json:"role" example:"STUDENT"`
}
