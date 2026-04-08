package domain

import "time"

type Profile struct {
	UserID                   string
	DisplayName              string
	Bio                      string
	ProfilePictureURL        string
	DisplayNameLastChangedAt time.Time
	RatingCount              int
	TotalReviewsCount        int
	SoldItemsCount           int
	FollowerCount            int
	FollowingCount           int
}
