package profile

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type ProfileHandler struct {
	profileService *services.ProfileService
	storageService *services.StorageService
	logger         middleware.StructuredLoggerInterface
}

func NewProfileHandler(
	profileService *services.ProfileService,
	storageService *services.StorageService,
	logger middleware.StructuredLoggerInterface,
) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		storageService: storageService,
		logger:         logger,
	}
}


// GetMyProfile retrieves the authenticated user's profile
// @Summary [Self] Get My Profile
// @Description Primary entry point for mobile apps. Returns aggregated user info, student data, and the unique `profile_id` required for all social/follow actions.
// @Tags Profile
// @Security BearerAuth
// @Produce json
// @Success 200 {object} services.ProfileDTO
// @Failure 401 {object} middleware.ErrorFrame
// @Failure 500 {object} middleware.ErrorFrame
// @Router /api/v1/profile [get]
func (h *ProfileHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewUnauthorizedError("missing user ID"), nil)
		return
	}

	profile, err := h.profileService.GetMyProfile(r.Context(), userID)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, profile)
}

type UpdateProfileReq struct {
	DisplayName       *string `json:"display_name"`
	Bio               *string `json:"bio"`
	ProfilePictureURL *string `json:"profile_picture_url"`
}

// UpdateProfile handles profile updates
// @Summary [Self] Update Base Profile
// @Description Updates social identity fields like display name, bio, and finishing the S3 profile picture upload sync.
// @Tags Profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body profile.UpdateProfileReq true "Profile Update Payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} middleware.ErrorFrame
// @Failure 401 {object} middleware.ErrorFrame
// @Failure 500 {object} middleware.ErrorFrame
// @Router /api/v1/profile [put]
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewUnauthorizedError("missing user ID"), nil)
		return
	}

	var req UpdateProfileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("body", "invalid json"), nil)
		return
	}

	err := h.profileService.UpdateProfile(r.Context(), userID, req.DisplayName, req.Bio, req.ProfilePictureURL)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Profile updated successfully"})
}

type UpdateStudentReq struct {
	Major        *string `json:"major"`
	AcademicYear *int    `json:"academic_year"`
}

// UpdateStudent handles student academic info updates
// @Summary [Self] Update Academic Info
// @Description Specifically for student-only data. Updates the `major` and `academic_year`. This is decoupled from the base profile to separate academic status from social identity.
// @Tags Profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body profile.UpdateStudentReq true "Student Update Payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} middleware.ErrorFrame
// @Failure 401 {object} middleware.ErrorFrame
// @Failure 500 {object} middleware.ErrorFrame
// @Router /api/v1/profile/student [put]
func (h *ProfileHandler) UpdateStudent(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewUnauthorizedError("missing user ID"), nil)
		return
	}

	var req UpdateStudentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("body", "invalid json"), nil)
		return
	}

	err := h.profileService.UpdateStudentDetails(r.Context(), userID, req.Major, req.AcademicYear)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Student details updated successfully"})
}

type GenerateUploadURLReq struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

// GenerateUploadURL generates an S3 presigned URL for profile pictures
// @Summary Generate S3 Upload URL
// @Description 📱 **FLUTTER INTEGRATION WORKFLOW (3 STEPS):**\n\n**Step 1 (The Secure Handshake):** Call this endpoint with `{"filename": "pic.png", "content_type": "image/png"}` to secure a ticket. It returns an `upload_url` and an `object_key`. Note: Ensure you store your `profile_id` returned from `GET /api/v1/profile` for all future social interactions.\n\n**Step 2 (The Direct Upload):** Bypass this backend entirely and upload your binary image file directly to AWS S3. Perform an HTTP `PUT` request targeting the `upload_url` you received. You MUST attach the exact same `Content-Type` header (e.g. `image/png`) to the `PUT` request. Do not send `FormData`, send raw binary data.\n\n**Step 3 (Final Sync):** Once AWS gives you a `200 OK`, call `PUT /api/v1/profile` and set `"profile_picture_url"` to the `object_key` that was returned in Step 1. Your updated profile is now ready to be discovered via the public profile lookup.
// @Tags Profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body profile.GenerateUploadURLReq true "File Info"
// @Success 200 {object} map[string]string
// @Failure 400 {object} middleware.ErrorFrame
// @Failure 401 {object} middleware.ErrorFrame
// @Failure 500 {object} middleware.ErrorFrame
// @Router /api/v1/profile/picture/upload-url [post]
func (h *ProfileHandler) GenerateUploadURL(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewUnauthorizedError("missing user ID"), nil)
		return
	}

	var req GenerateUploadURLReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("body", "invalid json"), nil)
		return
	}

	if req.Filename == "" || req.ContentType == "" {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("body", "filename and content_type are required"), nil)
		return
	}

	url, objectKey, err := h.storageService.GenerateProfilePictureUploadURL(r.Context(), userID, req.Filename, req.ContentType)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]string{
		"upload_url": url,
		"object_key": objectKey,
		"expires_in": (15 * time.Minute).String(),
	})
}

// GetPublicProfile retrieves a user's public profile info
// @Summary [Public] Discover User
// @Description Fetch public-facing info of any user (display name, bio, counters). Use this to render the profile page of others before a follow action.
// @Tags Social
// @Security BearerAuth
// @Produce json
// @Param profileId path string true "Profile ID"
// @Success 200 {object} services.ProfileDTO
// @Failure 401 {object} middleware.ErrorFrame
// @Failure 404 {object} middleware.ErrorFrame
// @Failure 500 {object} middleware.ErrorFrame
// @Router /api/v1/users/{profileId}/profile [get]
func (h *ProfileHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "profileId")
	if profileID == "" {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("path", "profileId is required"), nil)
		return
	}

	profile, err := h.profileService.GetPublicProfile(r.Context(), profileID)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, profile)
}

// GetPublicProfileByUserID retrieves a user's public profile using their user ID (owner_id from feed)
// @Summary [Public] Get Owner Profile by User ID
// @Description Fetch the public profile of an item owner using their user ID (the owner_id field returned in the feed). Use this when clicking on an item to show the seller's profile.
// @Tags Social
// @Security BearerAuth
// @Produce json
// @Param profileId path string true "User ID (owner_id from the item feed)"
// @Success 200 {object} services.ProfileDTO
// @Failure 401 {object} middleware.ErrorFrame
// @Failure 404 {object} middleware.ErrorFrame
// @Failure 500 {object} middleware.ErrorFrame
// @Router /api/v1/users/{profileId}/profile-by-user [get]
func (h *ProfileHandler) GetPublicProfileByUserID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "profileId")
	if userID == "" {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("path", "profileId is required"), nil)
		return
	}

	profile, err := h.profileService.GetPublicProfileByUserID(r.Context(), userID)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, profile)
}
