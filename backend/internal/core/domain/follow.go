package domain

import "time"

// Follow represents a follow relationship between two profiles.
type Follow struct {
	FollowerProfileID  string
	FollowingProfileID string
	FollowedAt         time.Time
	DisplayName        string
	ProfilePictureURL  string
}
