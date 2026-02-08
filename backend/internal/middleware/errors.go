package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"


)

//
// ==================================================
// 1️⃣ Error Codes (Domain + API Contract)
// ==================================================
//

type ErrCode string

const (
	// =========================
	// Auth / User
	// =========================
	CodeUserNotFound        ErrCode = "USER_NOT_FOUND"
	CodeEmailAlreadyUsed    ErrCode = "EMAIL_ALREADY_USED"
	CodeInvalidCredentials ErrCode = "INVALID_CREDENTIALS"
	CodeAccountSuspended   ErrCode = "ACCOUNT_SUSPENDED"
	CodeEmailNotVerified   ErrCode = "EMAIL_NOT_VERIFIED"
	CodeUnauthorized       ErrCode = "UNAUTHORIZED"
	CodeForbidden          ErrCode = "FORBIDDEN"
	CodeTokenInvalid       ErrCode = "TOKEN_INVALID"
	CodeTokenExpired       ErrCode = "TOKEN_EXPIRED"

	// =========================
	// Marketplace
	// =========================
	CodeListingNotFound     ErrCode = "LISTING_NOT_FOUND"
	CodeListingClosed       ErrCode = "LISTING_CLOSED"
	CodeListingLimitReached ErrCode = "LISTING_LIMIT_REACHED"

	// =========================
	// Reviews
	// =========================
	CodeAlreadyReviewed  ErrCode = "ALREADY_REVIEWED"
	CodeReviewNotAllowed ErrCode = "REVIEW_NOT_ALLOWED"

	// =========================
	// Validation / Generic
	// =========================
	CodeValidationFailed ErrCode = "VALIDATION_FAILED"
	CodeInvalidInput     ErrCode = "INVALID_INPUT"
	CodeConflict         ErrCode = "CONFLICT"
	CodeRateLimited      ErrCode = "RATE_LIMITED"

	// =========================
	// Infrastructure
	// =========================
	CodeDatabaseError ErrCode = "DATABASE_ERROR"
	CodeInternalError ErrCode = "INTERNAL_ERROR"
)

//
// ==================================================
// 2️⃣ Domain Error (USED INSIDE CORE)
// ==================================================
//

type DomainError struct {
	Code       ErrCode
	Message    string
	StatusCode int
	Retryable  bool
	Details    map[string]any
	Cause      error
}

func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return string(e.Code)
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

//
// ==================================================
// 3️⃣ Error Constructors (Services / Repositories)
// ==================================================
//

// ---------- Auth / User ----------

func NewUserNotFoundError(userID string) *DomainError {
	return &DomainError{
		Code:       CodeUserNotFound,
		Message:    "User not found",
		StatusCode: http.StatusNotFound,
		Retryable:  false,
		Details: map[string]any{
			"userId": userID,
		},
	}
}

func NewInvalidCredentialsError() *DomainError {
	return &DomainError{
		Code:       CodeInvalidCredentials,
		Message:    "Invalid email or password",
		StatusCode: http.StatusUnauthorized,
		Retryable:  false,
	}
}

func NewEmailAlreadyUsedError(email string) *DomainError {
	return &DomainError{
		Code:       CodeEmailAlreadyUsed,
		Message:    "Email already in use",
		StatusCode: http.StatusConflict,
		Retryable:  false,
		Details: map[string]any{
			"email": email,
		},
	}
}

func NewAccountSuspendedError() *DomainError {
	return &DomainError{
		Code:       CodeAccountSuspended,
		Message:    "Account is suspended",
		StatusCode: http.StatusForbidden,
		Retryable:  false,
	}
}

func NewUnauthorizedError(reason string) *DomainError {
	return &DomainError{
		Code:       CodeUnauthorized,
		Message:    reason,
		StatusCode: http.StatusUnauthorized,
		Retryable:  false,
	}
}

func NewForbiddenError(reason string) *DomainError {
	return &DomainError{
		Code:       CodeForbidden,
		Message:    reason,
		StatusCode: http.StatusForbidden,
		Retryable:  false,
	}
}

// ---------- Validation ----------

func NewValidationError(field, reason string) *DomainError {
	return &DomainError{
		Code:       CodeValidationFailed,
		Message:    "Validation failed",
		StatusCode: http.StatusBadRequest,
		Retryable:  false,
		Details: map[string]any{
			"field":  field,
			"reason": reason,
		},
	}
}

// ---------- Marketplace ----------

func NewListingNotFoundError(listingID string) *DomainError {
	return &DomainError{
		Code:       CodeListingNotFound,
		Message:    "Listing not found",
		StatusCode: http.StatusNotFound,
		Retryable:  false,
		Details: map[string]any{
			"listingId": listingID,
		},
	}
}

// ---------- Infrastructure ----------

func NewDatabaseError(operation string, cause error) *DomainError {
	return &DomainError{
		Code:       CodeDatabaseError,
		Message:    "Database operation failed",
		StatusCode: http.StatusInternalServerError,
		Retryable:  true,
		Details: map[string]any{
			"operation": operation,
		},
		Cause: cause,
	}
}

func NewInternalError(cause error) *DomainError {
	return &DomainError{
		Code:       CodeInternalError,
		Message:    "Internal server error",
		StatusCode: http.StatusInternalServerError,
		Retryable:  true,
		Cause:      cause,
	}
}

//
// ==================================================
// 4️⃣ Error Frame (API Response Shape)
// ==================================================
//

type ErrorFrame struct {
	Type        string         `json:"type"`
	Version     string         `json:"version"`
	Code        ErrCode        `json:"code"`
	Message     string         `json:"message"`
	StatusCode  int            `json:"statusCode"`
	Retryable   bool           `json:"retryable"`
	TS          int64          `json:"ts"`
	Correlation string         `json:"correlationId,omitempty"`
	Details     map[string]any `json:"details,omitempty"`
}

//
// ==================================================
// 5️⃣ Mapping (DomainError → ErrorFrame)
// ==================================================
//

func MapErrorToFrame(ctx context.Context, err error, correlationID string) ErrorFrame {
	frame := ErrorFrame{
		Type:        "error",
		Version:     "v1",
		TS:          time.Now().UnixMilli(),
		Correlation: correlationID,
	}

	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		frame.Code = domainErr.Code
		frame.Message = domainErr.Message
		frame.StatusCode = domainErr.StatusCode
		frame.Retryable = domainErr.Retryable
		frame.Details = domainErr.Details
		return frame
	}

	frame.Code = CodeInternalError
	frame.Message = "An unexpected error occurred"
	frame.StatusCode = http.StatusInternalServerError
	frame.Retryable = true

	return frame
}

//
// ==================================================
// 6️⃣ HTTP Helpers (Handlers use these only)
// ==================================================
//
func WriteErrorResponse(
	w http.ResponseWriter,
	ctx context.Context,
	err error,
	logger StructuredLoggerInterface,
) {
	correlationID := CorrelationIDFromContext(ctx)

	frame := MapErrorToFrame(ctx, err, correlationID)

	if logger != nil {
		logger.Error(ctx, "request.error", map[string]any{
			"code":       string(frame.Code),
			"statusCode": frame.StatusCode,
			"retryable":  frame.Retryable,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(frame.StatusCode)
	_ = json.NewEncoder(w).Encode(frame)
}


func WriteJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
