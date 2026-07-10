package http

import (
	"encoding/json"
	"math"
	"net/http"

	"strconv"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	adminService  ports.IAdminService
	reportService ports.IReportService
	logger        middleware.StructuredLoggerInterface
}

func NewAdminHandler(adminService ports.IAdminService, reportService ports.IReportService, logger middleware.StructuredLoggerInterface) *AdminHandler {
	return &AdminHandler{
		adminService:  adminService,
		reportService: reportService,
		logger:        logger,
	}
}

func (h *AdminHandler) Register(r chi.Router) {
	r.Get("/me", h.GetMe)
	r.Get("/dashboard", h.GetDashboard)
	r.Get("/users", h.GetUsers)
	r.Get("/users/{id}/stats", h.GetUserStats)
	r.Get("/verifications", h.GetVerifications)
	r.Get("/verifications/{id}/documents", h.GetVerificationDocuments)
	r.Post("/verifications/{id}/approve", h.ApproveVerification)
	r.Post("/verifications/{id}/reject", h.RejectVerification)
	r.Post("/users/{id}/warn", h.WarnUser)
	r.Post("/users/{id}/ban", h.BanUser)
	r.Get("/reports", h.GetReports)
	r.Post("/reports/{id}/resolve", h.ResolveReport)
}

func (h *AdminHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.ContextUserID).(string)
	role, _ := r.Context().Value(middleware.ContextRole).(string)

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"role":    role,
	})
}

func (h *AdminHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	// For testing, grab university_id from query if provided.
	universityID := r.URL.Query().Get("university_id")

	metrics, err := h.adminService.GetDashboardMetrics(r.Context(), universityID)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to get dashboard metrics", map[string]any{"error": err.Error()})
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, metrics)
}

func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	roleFilter := r.URL.Query().Get("role")
	statusFilter := r.URL.Query().Get("status")
	
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	
	offset := (page - 1) * limit

	users, totalCount, err := h.adminService.GetUsers(r.Context(), search, roleFilter, statusFilter, limit, offset)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to get users", map[string]any{"error": err.Error()})
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"users":       users,
		"total":       totalCount,
		"page":        page,
		"total_pages": totalPages,
	})
}

func (h *AdminHandler) GetVerifications(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	verifications, err := h.adminService.GetVerifications(r.Context(), status)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to get verifications", map[string]any{"error": err.Error()})
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"verifications": verifications,
	})
}

func (h *AdminHandler) ApproveVerification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	adminID, _ := r.Context().Value(middleware.ContextUserID).(string)
	
	if err := h.adminService.ApproveVerification(r.Context(), id, adminID); err != nil {
		h.logger.Error(r.Context(), "Failed to approve verification", map[string]any{"error": err.Error()})
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}
	middleware.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *AdminHandler) RejectVerification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	adminID, _ := r.Context().Value(middleware.ContextUserID).(string)
	
	var req struct {
		Comment string `json:"comment"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(r.Context(), "Invalid JSON payload", map[string]any{"error": err.Error()})
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := h.adminService.RejectVerification(r.Context(), id, adminID, req.Comment); err != nil {
		h.logger.Error(r.Context(), "Failed to reject verification", map[string]any{"error": err.Error()})
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}
	middleware.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *AdminHandler) WarnUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	adminID, _ := r.Context().Value(middleware.ContextUserID).(string)

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(r.Context(), "Invalid JSON payload", map[string]any{"error": err.Error()})
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := h.adminService.WarnUser(r.Context(), adminID, id, req.Reason); err != nil {
		h.logger.Error(r.Context(), "Failed to warn user", map[string]any{"error": err.Error()})
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}
	middleware.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *AdminHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	adminID, _ := r.Context().Value(middleware.ContextUserID).(string)

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(r.Context(), "Invalid JSON payload", map[string]any{"error": err.Error()})
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := h.adminService.BanUser(r.Context(), adminID, id, req.Reason); err != nil {
		h.logger.Error(r.Context(), "Failed to ban user", map[string]any{"error": err.Error()})
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}
	middleware.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *AdminHandler) GetVerificationDocuments(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	docs, err := h.adminService.GetVerificationDocuments(r.Context(), id)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to get verification documents", map[string]any{"error": err.Error()})
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}
	middleware.WriteJSONResponse(w, http.StatusOK, map[string]any{"documents": docs})
}

func (h *AdminHandler) GetReports(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit <= 0 {
		limit = 50
	}

	reports, err := h.reportService.ListReports(r.Context(), int32(limit), int32(offset))
	if err != nil {
		h.logger.Error(r.Context(), "Failed to get reports", map[string]any{"error": err.Error()})
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]any{"reports": reports})
}

func (h *AdminHandler) ResolveReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	adminID, _ := r.Context().Value(middleware.ContextUserID).(string)

	var req struct {
		ActionTaken string               `json:"action_taken"`
		Status      domain.ReportStatus  `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(r.Context(), "Invalid JSON payload", map[string]any{"error": err.Error()})
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Derive status from action_taken if status not set directly
	status := req.Status
	if status == "" {
		switch req.ActionTaken {
		case "dismissed":
			status = domain.ReportStatusDismissed
		default:
			status = domain.ReportStatusResolved
		}
	}

	report, err := h.reportService.ResolveReport(r.Context(), id, adminID, status)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to resolve report", map[string]any{"error": err.Error()})
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	// Execute side effects based on action taken
	if req.ActionTaken == "warned_user" {
		if report.EntityType == "USER" {
			_ = h.adminService.WarnUser(r.Context(), adminID, report.ReportedEntityID, "A report was filed against your profile. Please note that accumulating 3 warnings will result in an automatic account ban.")
		} else if report.EntityType == "ITEM" {
			err := h.adminService.WarnItemOwner(r.Context(), adminID, report.ReportedEntityID, "")
			if err != nil {
				h.logger.Error(r.Context(), "Failed to warn item owner", map[string]any{"error": err.Error(), "item_id": report.ReportedEntityID})
			}
		}
	} else if req.ActionTaken == "deleted_item" && report.EntityType == "ITEM" {
		_ = h.adminService.DeleteItemAdmin(r.Context(), adminID, report.ReportedEntityID, "Deleted due to reports.")
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]any{"report": report})
}

func (h *AdminHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	reportsReceived, warningsReceived, totalPosts, err := h.adminService.GetAdminUserProfileStats(r.Context(), userID)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to fetch user stats", map[string]any{"error": err.Error(), "user_id": userID})
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"reports_received":  reportsReceived,
		"warnings_received": warningsReceived,
		"total_posts":       totalPosts,
	})
}
