package dto

type VerifyActivationRequest struct {
	Token string `json:"token"`
}
type VerifyActivationResponse struct {
	Message string `json:"message"`
}
