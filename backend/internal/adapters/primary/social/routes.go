package social

import (
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/profile"
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	Follow  *FollowHandler
	Profile *profile.ProfileHandler
}

func NewRoutes(followH *FollowHandler, profileH *profile.ProfileHandler) *Routes {
	return &Routes{
		Follow:  followH,
		Profile: profileH,
	}
}

func (r *Routes) Register(router chi.Router) {
	router.Route("/users/{profileId}", func(ru chi.Router) {
		// Public — anyone can view
		ru.Get("/profile", r.Profile.GetPublicProfile)
		ru.Get("/followers", r.Follow.GetFollowers)
		ru.Get("/following", r.Follow.GetFollowing)

		// Protected (auth required)
		ru.Group(func(rg chi.Router) {
			rg.Post("/follow", r.Follow.FollowUser)
			rg.Delete("/unfollow", r.Follow.UnfollowUser)
			rg.Get("/is-following", r.Follow.IsFollowing)
		})
	})
}
