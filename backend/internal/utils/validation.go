package utils

import (
	"net/mail"
	"regexp"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// IsValidEmail validates an email address using both standard net/mail and a strict regex.
func IsValidEmail(e string) bool {
	if len(e) < 3 || len(e) > 254 {
		return false
	}

	// Parse using standard Go library
	_, err := mail.ParseAddress(e)
	if err != nil {
		return false
	}

	// Enforce strict format (must have domain extension)
	return emailRegex.MatchString(e)
}
