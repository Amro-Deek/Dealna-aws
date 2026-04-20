package social

import (
	"encoding/json"
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type FollowHandler struct {
	svc *services.FollowService
}

func NewFollowHandler(svc *services.FollowService) *FollowHandler {
	return &FollowHandler{svc: svc}
}

// FollowUser godoc
// @Summary      Follow a user
// @Description  Authenticated user follows the profile with the given profileId
// @Tags         Social
// @Security     BearerAuth
// @Param        profileId  path  string  true  "Target profile ID"
// @Success      204
// @Failure      400  {string}  string  "cannot follow yourself"
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /users/{profileId}/follow [post]
func (h *FollowHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	targetProfileID := chi.URLParam(r, "profileId")
	if err := h.svc.FollowUser(r.Context(), userID, targetProfileID); err != nil {
		if err.Error() == "cannot follow yourself" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UnfollowUser godoc
// @Summary      Unfollow a user
// @Description  Authenticated user unfollows the profile with the given profileId
// @Tags         Social
// @Security     BearerAuth
// @Param        profileId  path  string  true  "Target profile ID"
// @Success      204
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /users/{profileId}/unfollow [delete]
func (h *FollowHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	targetProfileID := chi.URLParam(r, "profileId")
	if err := h.svc.UnfollowUser(r.Context(), userID, targetProfileID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// IsFollowing godoc
// @Summary      Check follow status
// @Description  Returns whether the authenticated user follows the given profile
// @Tags         Social
// @Security     BearerAuth
// @Param        profileId  path  string  true  "Target profile ID"
// @Success      200  {object}  map[string]bool
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /users/{profileId}/is-following [get]
func (h *FollowHandler) IsFollowing(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	targetProfileID := chi.URLParam(r, "profileId")
	result, err := h.svc.IsFollowing(r.Context(), userID, targetProfileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"isFollowing": result})
}

// GetFollowers godoc
// @Summary      Get followers
// @Description  Returns all profiles following the given profileId
// @Tags         Social
// @Param        profileId  path  string  true  "Profile ID"
// @Success      200  {array}   domain.Follow
// @Failure      500  {string}  string  "internal error"
// @Router       /users/{profileId}/followers [get]
func (h *FollowHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "profileId")
	followers, err := h.svc.GetFollowers(r.Context(), profileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(followers)
}

// GetFollowing godoc
// @Summary      Get following
// @Description  Returns all profiles that the given profileId follows
// @Tags         Social
// @Param        profileId  path  string  true  "Profile ID"
// @Success      200  {array}   domain.Follow
// @Failure      500  {string}  string  "internal error"
// @Router       /users/{profileId}/following [get]
func (h *FollowHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "profileId")
	following, err := h.svc.GetFollowing(r.Context(), profileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(following)
}
