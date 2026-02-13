package dto

type RequestActivationRequest struct {
	Email string `json:"email"`
}

type RequestActivationResponse struct {
	Message string `json:"message"`
}
