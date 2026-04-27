package items

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type ItemHandler struct {
	itemService *services.ItemService
	logger      middleware.StructuredLoggerInterface
}

func NewItemHandler(service *services.ItemService, logger middleware.StructuredLoggerInterface) *ItemHandler {
	return &ItemHandler{itemService: service, logger: logger}
}

// GenerateUploadURL godoc
// @Summary      Generate Item Image Upload URL
// @Description  Creates a pre-signed S3 URL for uploading item images.
// @Tags         Items
// @Security     BearerAuth
// @Produce      json
// @Param        content_type  query    string  false  "MIME type (e.g. image/jpeg)"
// @Success      200           {object} map[string]string
// @Failure      401           {object} middleware.ErrorFrame
// @Router       /api/v1/items/picture/upload-url [post]
func (h *ItemHandler) GenerateUploadURL(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.UserIDFromContext(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewUnauthorizedError("invalid user authentication"), h.logger)
		return
	}

	contentType := r.URL.Query().Get("content_type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	url, objectKey, err := h.itemService.GenerateItemImageUploadURL(r.Context(), userID, contentType)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("upload", err.Error()), h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]string{
		"upload_url": url,
		"object_key": objectKey,
	})
}

// CreateItem godoc
// @Summary      Create Item Listing
// @Description  Mints a new item for sale. Enforces 5 listings per day limit (NFR-1).
// @Tags         Items
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        item  body     domain.CreateItemCommand  true  "Listing Details"
// @Success      201   {object} domain.Item
// @Failure      400   {object} middleware.ErrorFrame
// @Failure      401   {object} middleware.ErrorFrame
// @Failure      429   {object} middleware.ErrorFrame
// @Router       /api/v1/items [post]
func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var cmd domain.CreateItemCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("body", "invalid request body"), h.logger)
		return
	}

	userIDStr := middleware.UserIDFromContext(r.Context())
	if uID, err := uuid.Parse(userIDStr); err == nil {
		cmd.OwnerID = uID
	} else {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewUnauthorizedError("unauthorized"), h.logger)
		return
	}

	item, err := h.itemService.CreateItem(r.Context(), cmd)
	if err != nil {
		if err == services.ErrDailyItemLimitReached {
			middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("limit", err.Error()), h.logger)
			return
		}
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("item", err.Error()), h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusCreated, item)
}

// GetFeed godoc
// @Summary      Get Marketplace Feed
// @Description  Returns available items scoped automatically to the authenticated user's university. Newest first.
// @Tags         Items
// @Security     BearerAuth
// @Produce      json
// @Param        limit        query    int     false  "Items per page (default 20)"
// @Param        offset       query    int     false  "Pagination offset"
// @Param        search       query    string  false  "Full-text search on title"
// @Param        category_id  query    string  false  "Filter by category UUID"
// @Success      200          {array}  domain.FeedItem
// @Failure      401          {object} middleware.ErrorFrame
// @Router       /api/v1/items/feed [get]
func (h *ItemHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.UserIDFromContext(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewUnauthorizedError("unauthorized"), h.logger)
		return
	}

	// Auto-resolve university from the user's profile — no header needed.
	universityID, err := h.itemService.GetUserUniversityID(r.Context(), userID)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewInternalError(err), h.logger)
		return
	}

	filter := domain.ItemFilter{
		Limit:                 20,
		Offset:                0,
		RequesterUniversityID: universityID,
		ExcludedOwnerID:       userID,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = int32(l)
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = int32(o)
		}
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filter.SearchQuery = &search
	}
	if catIDStr := r.URL.Query().Get("category_id"); catIDStr != "" {
		if catID, err := uuid.Parse(catIDStr); err == nil {
			filter.CategoryID = &catID
		}
	}

	items, err := h.itemService.GetGlobalFeed(r.Context(), filter)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewInternalError(err), h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, items)
}

// GetItemDetail godoc
// @Summary      Get Item Detail
// @Description  Returns the full listing including attachments.
// @Tags         Items
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string  true  "Item UUID"
// @Success      200  {object} domain.ItemDetail
// @Failure      400  {object} middleware.ErrorFrame
// @Failure      401  {object} middleware.ErrorFrame
// @Router       /api/v1/items/{id} [get]
func (h *ItemHandler) GetItemDetail(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("id", "invalid item id"), h.logger)
		return
	}

	detail, err := h.itemService.GetItemDetail(r.Context(), itemID)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewInternalError(err), h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, detail)
}

// GetMyItems godoc
// @Summary      Get My Listings
// @Description  Returns all items owned by the currently authenticated user.
// @Tags         Items
// @Security     BearerAuth
// @Produce      json
// @Param        limit   query    int  false  "Items per page (default 20)"
// @Param        offset  query    int  false  "Pagination offset"
// @Success      200     {array}  domain.FeedItem
// @Failure      401     {object} middleware.ErrorFrame
// @Router       /api/v1/items/my [get]
func (h *ItemHandler) GetMyItems(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.UserIDFromContext(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewUnauthorizedError("unauthorized"), h.logger)
		return
	}

	limit := 20
	offset := 0
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil {
			offset = o
		}
	}

	items, err := h.itemService.GetUserStorefront(r.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewInternalError(err), h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, items)
}

// UpdateStatus godoc
// @Summary      Update Item Status
// @Description  Transitions an item status (AVAILABLE → RESERVED → SOLD). Only the owner can update.
// @Tags         Items
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path     string            true  "Item UUID"
// @Param        status  body     map[string]string true  "New status payload"
// @Success      200     {object} map[string]string
// @Failure      400     {object} middleware.ErrorFrame
// @Failure      401     {object} middleware.ErrorFrame
// @Failure      403     {object} middleware.ErrorFrame
// @Router       /api/v1/items/{id}/status [patch]
func (h *ItemHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("id", "invalid item id"), h.logger)
		return
	}

	userIDStr := middleware.UserIDFromContext(r.Context())
	userID, _ := uuid.Parse(userIDStr)

	detail, err := h.itemService.GetItemDetail(r.Context(), itemID)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewInternalError(err), h.logger)
		return
	}
	if detail.OwnerID != userID {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewForbiddenError("you do not own this item"), h.logger)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("body", "invalid body"), h.logger)
		return
	}

	err = h.itemService.UpdateStatus(r.Context(), itemID, domain.ItemStatus(req.Status))
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("status", err.Error()), h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "status updated"})
}

// DeleteItem godoc
// @Summary      Delete Item
// @Description  Soft-deletes a listing. Only the owner can delete. Preserves transaction history.
// @Tags         Items
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string  true  "Item UUID"
// @Success      200  {object} map[string]string
// @Failure      400  {object} middleware.ErrorFrame
// @Failure      401  {object} middleware.ErrorFrame
// @Failure      403  {object} middleware.ErrorFrame
// @Router       /api/v1/items/{id} [delete]
func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("id", "invalid item id"), h.logger)
		return
	}

	userIDStr := middleware.UserIDFromContext(r.Context())
	userID, _ := uuid.Parse(userIDStr)

	detail, err := h.itemService.GetItemDetail(r.Context(), itemID)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewInternalError(err), h.logger)
		return
	}
	if detail.OwnerID != userID {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewForbiddenError("you do not own this item"), h.logger)
		return
	}

	err = h.itemService.SoftDeleteItem(r.Context(), itemID)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewInternalError(err), h.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "item deleted"})
}

// GetCategories godoc
// @Summary      List Categories
// @Description  Returns all marketplace categories for dropdown selection. Public endpoint, no auth required.
// @Tags         Metadata
// @Produce      json
// @Success      200  {array}  domain.Category
// @Router       /api/v1/categories [get]
func (h *ItemHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.itemService.ListCategories(r.Context())
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewInternalError(err), h.logger)
		return
	}
	middleware.WriteJSONResponse(w, http.StatusOK, cats)
}
