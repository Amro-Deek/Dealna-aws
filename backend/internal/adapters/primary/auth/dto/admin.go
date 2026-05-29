package dto

type RejectProviderApplicationRequest struct {
	Comment string `json:"comment" validate:"required"`
}
