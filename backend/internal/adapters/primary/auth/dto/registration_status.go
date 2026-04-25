package dto

import "time"

type StudentRegistrationStatusResponse struct {
	Email                   string     `json:"email"`
	IsVerified              bool       `json:"is_verified"`
	IsUsed                  bool       `json:"is_used"`
	ExpiresAt               time.Time  `json:"expires_at"`
	VerifiedAt              *time.Time `json:"verified_at,omitempty"`
	CanCompleteRegistration bool       `json:"can_complete_registration"`
}
