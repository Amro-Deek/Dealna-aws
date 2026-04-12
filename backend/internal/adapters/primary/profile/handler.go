package profile

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
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
// @Summary Get My Profile
// @Description Fetch and aggregate User, Profile, and Student data into a unified DTO
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
// @Summary Update Profile
// @Description Updates display name, bio, and profile picture URL with business constraints
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
// @Summary Update Student Details
// @Description Updates major and academic year of a student
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
// @Description 📱 **FLUTTER INTEGRATION WORKFLOW (3 STEPS):**\n\n**Step 1:** Call this endpoint with `{"filename": "pic.png", "content_type": "image/png"}` to secure a ticket. It returns an `upload_url` and an `object_key`.\n\n**Step 2 (The Heavy Lift):** Bypass this backend entirely and upload your binary image file directly to AWS S3. Perform an HTTP `PUT` request targeting the `upload_url` you received. You MUST attach the exact same `Content-Type` header (e.g. `image/png`) to the PUT request. Do not send FormData, send raw binary.\n\n**Step 3 (Wrap Up):** Once AWS gives you a `200 OK`, call `PUT /api/v1/profile` and set `"profile_picture_url"` to the `object_key` that was returned in Step 1.
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
