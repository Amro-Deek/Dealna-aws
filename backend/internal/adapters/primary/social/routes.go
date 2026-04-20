package social

import "github.com/go-chi/chi/v5"

type Routes struct {
	Follow *FollowHandler
}

func NewRoutes(followH *FollowHandler) *Routes {
	return &Routes{Follow: followH}
}

func (r *Routes) Register(router chi.Router) {
	router.Route("/users/{profileId}", func(ru chi.Router) {
		// Protected (auth required)
		ru.Post("/follow", r.Follow.FollowUser)
		ru.Delete("/unfollow", r.Follow.UnfollowUser)
		ru.Get("/is-following", r.Follow.IsFollowing)

		// Public — anyone can view
		ru.Get("/followers", r.Follow.GetFollowers)
		ru.Get("/following", r.Follow.GetFollowing)
	})
}
