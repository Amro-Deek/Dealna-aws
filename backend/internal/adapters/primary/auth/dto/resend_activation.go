package dto

type ResendActivationRequest struct {
	Email string `json:"email"`
}
type ResendActivationResponse struct {
	Message string `json:"message"`
}
