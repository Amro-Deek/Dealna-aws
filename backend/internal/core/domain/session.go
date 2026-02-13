package domain 
import "time"

type Session struct {
	SessionID string
	UserID    string
	JTI       string
	Revoked   bool
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}