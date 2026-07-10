package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/google/uuid"
)

type ReportHandler struct {
	reportService ports.IReportService
	storage       ports.IStorageProvider
	logger        middleware.StructuredLoggerInterface
}

func NewReportHandler(reportService ports.IReportService, storage ports.IStorageProvider, logger middleware.StructuredLoggerInterface) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		storage:       storage,
		logger:        logger,
	}
}

func (h *ReportHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	reporterID, _ := r.Context().Value(middleware.ContextUserID).(string)

	var reportedEntityID, entityTypeStr, reportTypeStr, description string
	var attachmentURL string

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var req struct {
			ReportedEntityID string `json:"reported_entity_id"`
			EntityType       string `json:"entity_type"`
			Type             string `json:"type"`
			Description      string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("payload", "Invalid JSON payload"), h.logger)
			return
		}
		reportedEntityID = req.ReportedEntityID
		entityTypeStr = req.EntityType
		reportTypeStr = req.Type
		description = req.Description
	} else {
		// Enforce 3MB limit
		const maxUploadSize = 3 << 20 // 3 MB
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			h.logger.Error(r.Context(), "File too large or invalid multipart form", map[string]any{"error": err.Error()})
			middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("file", "File exceeds 3MB limit or invalid format"), h.logger)
			return
		}

		reportedEntityID = r.FormValue("reported_entity_id")
		entityTypeStr = r.FormValue("entity_type")
		reportTypeStr = r.FormValue("type")
		description = r.FormValue("description")

		file, header, err := r.FormFile("attachment")
		if err == nil {
			defer file.Close()
			ext := filepath.Ext(header.Filename)
			objectKey := fmt.Sprintf("reports/%s/%s%s", reporterID, uuid.NewString(), ext)
			contentType := header.Header.Get("Content-Type")

			s3URL, uploadErr := h.storage.UploadFile(r.Context(), objectKey, contentType, file)
			if uploadErr != nil {
				h.logger.Error(r.Context(), "Failed to upload report attachment", map[string]any{"error": uploadErr.Error()})
				middleware.WriteErrorResponse(w, r.Context(), middleware.NewInternalError(uploadErr), h.logger)
				return
			}
			attachmentURL = s3URL
		} else if err != http.ErrMissingFile {
			h.logger.Error(r.Context(), "Error retrieving file from form", map[string]any{"error": err.Error()})
			middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("attachment", "Invalid file attachment"), h.logger)
			return
		}
	}

	entityType := domain.ReportEntityType(entityTypeStr)

	if reportTypeStr == "SCAM" {
		reportTypeStr = "FRAUD"
	}
	reportType := domain.ReportType(reportTypeStr)

	if reportedEntityID == "" || entityType == "" || reportType == "" || description == "" {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("payload", "Missing required fields"), h.logger)
		return
	}

	report, err := h.reportService.CreateReport(r.Context(), reporterID, reportedEntityID, entityType, reportType, description, attachmentURL)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to create report", map[string]any{"error": err.Error()})
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusCreated, report)
}
