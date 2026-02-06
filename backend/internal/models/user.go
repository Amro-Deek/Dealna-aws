package models

import "github.com/google/uuid"

type UserRole string

const (
	RoleAdmin            UserRole = "Admin"
	RoleVerifiedStudent  UserRole = "Verified Student"
	RoleLimitedStudent   UserRole = "Limited Student"
	RoleVerifiedProvider UserRole = "Verified Provider"
)

type User struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Password string    `json:"-"` // Hidden in JSON
	Role     UserRole  `json:"role"`
}
