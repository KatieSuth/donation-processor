// Package handler implements the boilerplate REST API handlers.
package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/KatieSuth/donation-processor/backend/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler holds shared dependencies for HTTP handlers.
type Handler struct {
	frontendURL string
	store       store.Store
}

// New builds a Handler.
func New(frontendURL string, s store.Store) *Handler {
	return &Handler{
		frontendURL: frontendURL,
		store:       s,
	}
}

// GET /health
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"message":   "API is running",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// GET /donations
func (h *Handler) ListDonations(c *gin.Context) {
	donations, err := h.store.ListDonations(c.Request.Context())
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "list donations failed", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to list donations",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"donations": donations,
	})
}

// GET /donations/:uuid
func (h *Handler) GetDonation(c *gin.Context) {
	donationUUID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		writeValidationError(c, "uuid must be a valid UUID")
		return
	}

	donation, err := h.store.GetDonation(c.Request.Context(), donationUUID)
	if err != nil {
		if errors.Is(err, store.ErrDonationNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Donation not found",
			})
			return
		}

		slog.ErrorContext(c.Request.Context(), "get donation failed", "error", err, "uuid", donationUUID.String())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to retrieve donation",
		})
		return
	}

	c.JSON(http.StatusOK, donation)
}

type createDonationRequest struct {
	UUID          string `json:"uuid"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	PaymentMethod string `json:"paymentMethod"`
	NonprofitID   string `json:"nonprofitId"`
	DonorID       string `json:"donorId"`
	Status        string `json:"status"`
	CreatedAt     string `json:"createdAt"`
}

type updateDonationStatusRequest struct {
	Status string `json:"status"`
}

// POST /donations
func (h *Handler) CreateDonation(c *gin.Context) {
	var req createDonationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Improper json or json value types",
		})
		return
	}

	donationUUID, err := uuid.Parse(req.UUID)
	if err != nil {
		writeValidationError(c, "uuid must be a valid UUID")
		return
	}
	if req.Amount < 0 {
		writeValidationError(c, "amount must be greater than or equal to 0")
		return
	}
	if req.Currency != "USD" {
		writeValidationError(c, "currency must be USD")
		return
	}
	if !isAllowedPaymentMethod(req.PaymentMethod) {
		writeValidationError(c, "paymentMethod must be one of: cc, ach, crypto, venmo")
		return
	}
	if strings.TrimSpace(req.NonprofitID) == "" {
		writeValidationError(c, "nonprofitId is required")
		return
	}
	if strings.TrimSpace(req.DonorID) == "" {
		writeValidationError(c, "donorId is required")
		return
	}
	if !isAllowedStatus(req.Status) {
		writeValidationError(c, "status must be one of: new, pending, success, failure")
		return
	}
	createdAt, err := time.Parse(time.RFC3339, req.CreatedAt)
	if err != nil {
		writeValidationError(c, "createdAt must be a valid RFC3339 datetime")
		return
	}

	donation, err := h.store.CreateDonation(
		c.Request.Context(),
		donationUUID,
		req.Amount,
		req.Currency,
		req.PaymentMethod,
		req.NonprofitID,
		req.DonorID,
		req.Status,
		createdAt.UTC(),
	)
	if err != nil {
		if errors.Is(err, store.ErrDonationConflict) {
			slog.WarnContext(
				c.Request.Context(),
				"donation already exists",
				"uuid", req.UUID,
				"error", err,
			)
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"status":  "error",
				"message": "Donation with this uuid already exists",
			})
			return
		}

		slog.ErrorContext(c.Request.Context(), "create donation failed", "error", err, "uuid", req.UUID)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create donation",
		})
		return
	}

	c.JSON(http.StatusCreated, donation)
}

// PATCH /donations/:uuid/status
func (h *Handler) UpdateDonationStatus(c *gin.Context) {
	donationUUID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		writeValidationError(c, "uuid must be a valid UUID")
		return
	}

	var req updateDonationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Improper json or json value types",
		})
		return
	}
	if !isAllowedStatus(req.Status) {
		writeValidationError(c, "status must be one of: new, pending, success, failure")
		return
	}

	donation, err := h.store.UpdateDonationStatus(c.Request.Context(), donationUUID, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrDonationNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Donation not found",
			})
			return
		case errors.Is(err, store.ErrDonationStatusConflict):
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"status":  "error",
				"message": "Target status is identical to current status",
			})
			return
		case errors.Is(err, store.ErrInvalidDonationStatusTransition):
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		default:
			slog.ErrorContext(c.Request.Context(), "update donation status failed", "error", err, "uuid", donationUUID.String())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to update donation status",
			})
			return
		}
	}

	c.JSON(http.StatusOK, donation)
}

func writeValidationError(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"status":  "error",
		"message": message,
	})
}

func isAllowedPaymentMethod(paymentMethod string) bool {
	switch paymentMethod {
	case "cc", "ach", "crypto", "venmo":
		return true
	default:
		return false
	}
}

func isAllowedStatus(status string) bool {
	switch status {
	case "new", "pending", "success", "failure":
		return true
	default:
		return false
	}
}
