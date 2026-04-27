// Package handler_test is a black-box test package for the HTTP API surface.
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KatieSuth/donation-processor/backend/internal/db"
	"github.com/KatieSuth/donation-processor/backend/internal/handler"
	"github.com/KatieSuth/donation-processor/backend/internal/model"
	"github.com/KatieSuth/donation-processor/backend/internal/store"
	"github.com/KatieSuth/donation-processor/backend/internal/test_util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealth checks that the health endpoint returns a 200 OK
// and the expected status message.
func TestHealth(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		h := newTestHandler(t, s)

		// Create a Gin test context for a GET request
		c, w := test_util.NewGinContext(http.MethodGet, "/health")

		// Call the handler
		h.Health(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		body := test_util.DecodeJSON[map[string]string](t, w)
		assert.Equal(t, "ok", body["status"])
		assert.Equal(t, "API is running", body["message"])
		assert.NotEmpty(t, body["timestamp"], "timestamp should be present in health check")
	})
}

func TestCreateDonation(t *testing.T) {
	type donationResponse struct {
		UUID          string `json:"uuid"`
		Amount        int64  `json:"amount"`
		Currency      string `json:"currency"`
		PaymentMethod string `json:"paymentMethod"`
		NonprofitID   string `json:"nonprofitId"`
		DonorID       string `json:"donorId"`
		Status        string `json:"status"`
		CreatedAt     string `json:"createdAt"`
		UpdatedAt     string `json:"updatedAt"`
	}

	type errorResponse struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		h := newTestHandler(t, s)

		t.Run("returns 201 when payload is valid", func(t *testing.T) {
			id := uuid.New()
			createdAt := time.Now().UTC().Truncate(time.Second)
			body := map[string]any{
				"uuid":          id.String(),
				"amount":        5000,
				"currency":      "USD",
				"paymentMethod": "cc",
				"nonprofitId":   "np_001",
				"donorId":       "donor_001",
				"status":        "new",
				"createdAt":     createdAt.Format(time.RFC3339),
			}

			w := performDonationRequest(t, h, body)
			require.Equal(t, http.StatusCreated, w.Code)

			resp := test_util.DecodeJSON[donationResponse](t, w)
			require.Equal(t, id.String(), resp.UUID)
			require.Equal(t, int64(5000), resp.Amount)
			require.Equal(t, "USD", resp.Currency)
			require.Equal(t, "cc", resp.PaymentMethod)
			require.Equal(t, "np_001", resp.NonprofitID)
			require.Equal(t, "donor_001", resp.DonorID)
			require.Equal(t, "new", resp.Status)
			gotCreatedAt, err := time.Parse(time.RFC3339, resp.CreatedAt)
			require.NoError(t, err)
			require.True(t, gotCreatedAt.Equal(createdAt))
			require.Equal(t, resp.CreatedAt, resp.UpdatedAt)
		})

		t.Run("returns 409 for duplicate uuid", func(t *testing.T) {
			id := uuid.New()
			body := map[string]any{
				"uuid":          id.String(),
				"amount":        1200,
				"currency":      "USD",
				"paymentMethod": "ach",
				"nonprofitId":   "np_002",
				"donorId":       "donor_002",
				"status":        "pending",
				"createdAt":     time.Now().UTC().Format(time.RFC3339),
			}

			first := performDonationRequest(t, h, body)
			require.Equal(t, http.StatusCreated, first.Code)

			second := performDonationRequest(t, h, body)
			require.Equal(t, http.StatusConflict, second.Code)
			errResp := test_util.DecodeJSON[errorResponse](t, second)
			require.Equal(t, "error", errResp.Status)
		})

		t.Run("returns 400 for invalid payload", func(t *testing.T) {
			body := map[string]any{
				"uuid":          uuid.New().String(),
				"amount":        1200,
				"currency":      "EUR",
				"paymentMethod": "ach",
				"nonprofitId":   "np_003",
				"donorId":       "donor_003",
				"status":        "pending",
				"createdAt":     time.Now().UTC().Format(time.RFC3339),
			}

			resp := performDonationRequest(t, h, body)
			require.Equal(t, http.StatusBadRequest, resp.Code)
			errResp := test_util.DecodeJSON[errorResponse](t, resp)
			require.Equal(t, "error", errResp.Status)
		})

		t.Run("returns 400 for invalid payment method", func(t *testing.T) {
			body := map[string]any{
				"uuid":          uuid.New().String(),
				"amount":        1200,
				"currency":      "USD",
				"paymentMethod": "wire",
				"nonprofitId":   "np_004",
				"donorId":       "donor_004",
				"status":        "pending",
				"createdAt":     time.Now().UTC().Format(time.RFC3339),
			}

			resp := performDonationRequest(t, h, body)
			require.Equal(t, http.StatusBadRequest, resp.Code)
			errResp := test_util.DecodeJSON[errorResponse](t, resp)
			require.Equal(t, "error", errResp.Status)
			require.Equal(t, "paymentMethod must be one of: cc, ach, crypto, venmo", errResp.Message)
		})
	})
}

func TestListDonations(t *testing.T) {
	type donationResponse struct {
		UUID string `json:"uuid"`
	}
	type listResponse struct {
		Donations []donationResponse `json:"donations"`
	}

	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		h := newTestHandler(t, s)

		donationA := map[string]any{
			"uuid":          uuid.New().String(),
			"amount":        5000,
			"currency":      "USD",
			"paymentMethod": "cc",
			"nonprofitId":   "np_101",
			"donorId":       "donor_101",
			"status":        "new",
			"createdAt":     time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339),
		}
		donationB := map[string]any{
			"uuid":          uuid.New().String(),
			"amount":        9000,
			"currency":      "USD",
			"paymentMethod": "ach",
			"nonprofitId":   "np_102",
			"donorId":       "donor_102",
			"status":        "pending",
			"createdAt":     time.Now().UTC().Format(time.RFC3339),
		}

		require.Equal(t, http.StatusCreated, performDonationRequest(t, h, donationA).Code)
		require.Equal(t, http.StatusCreated, performDonationRequest(t, h, donationB).Code)

		c, w := test_util.NewGinContext(http.MethodGet, "/donations")
		h.ListDonations(c)

		require.Equal(t, http.StatusOK, w.Code)
		resp := test_util.DecodeJSON[listResponse](t, w)
		require.GreaterOrEqual(t, len(resp.Donations), 2)
	})
}

func TestGetDonation(t *testing.T) {
	type donationResponse struct {
		UUID string `json:"uuid"`
	}
	type errorResponse struct {
		Status string `json:"status"`
	}

	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		h := newTestHandler(t, s)

		t.Run("returns 200 when donation exists", func(t *testing.T) {
			id := uuid.New()
			body := map[string]any{
				"uuid":          id.String(),
				"amount":        4000,
				"currency":      "USD",
				"paymentMethod": "cc",
				"nonprofitId":   "np_201",
				"donorId":       "donor_201",
				"status":        "success",
				"createdAt":     time.Now().UTC().Format(time.RFC3339),
			}
			require.Equal(t, http.StatusCreated, performDonationRequest(t, h, body).Code)

			c, w := test_util.NewGinContext(http.MethodGet, "/donations/"+id.String())
			c.Params = gin.Params{{Key: "uuid", Value: id.String()}}
			h.GetDonation(c)

			require.Equal(t, http.StatusOK, w.Code)
			resp := test_util.DecodeJSON[donationResponse](t, w)
			require.Equal(t, id.String(), resp.UUID)
		})

		t.Run("returns 404 when donation is not found", func(t *testing.T) {
			missingID := uuid.New()
			c, w := test_util.NewGinContext(http.MethodGet, "/donations/"+missingID.String())
			c.Params = gin.Params{{Key: "uuid", Value: missingID.String()}}
			h.GetDonation(c)

			require.Equal(t, http.StatusNotFound, w.Code)
			resp := test_util.DecodeJSON[errorResponse](t, w)
			require.Equal(t, "error", resp.Status)
		})
	})
}

func TestUpdateDonationStatus(t *testing.T) {
	type donationResponse struct {
		Status string `json:"status"`
	}
	type errorResponse struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		h := newTestHandler(t, s)

		t.Run("returns 200 for valid transition new to pending", func(t *testing.T) {
			id := uuid.New()
			createBody := map[string]any{
				"uuid":          id.String(),
				"amount":        2500,
				"currency":      "USD",
				"paymentMethod": "cc",
				"nonprofitId":   "np_301",
				"donorId":       "donor_301",
				"status":        "new",
				"createdAt":     time.Now().UTC().Format(time.RFC3339),
			}
			require.Equal(t, http.StatusCreated, performDonationRequest(t, h, createBody).Code)

			w := performUpdateDonationStatusRequest(t, h, id.String(), map[string]any{"status": "pending"})
			require.Equal(t, http.StatusOK, w.Code)
			resp := test_util.DecodeJSON[donationResponse](t, w)
			require.Equal(t, "pending", resp.Status)
		})

		t.Run("returns 409 when target status matches current", func(t *testing.T) {
			id := uuid.New()
			createBody := map[string]any{
				"uuid":          id.String(),
				"amount":        2500,
				"currency":      "USD",
				"paymentMethod": "ach",
				"nonprofitId":   "np_302",
				"donorId":       "donor_302",
				"status":        "pending",
				"createdAt":     time.Now().UTC().Format(time.RFC3339),
			}
			require.Equal(t, http.StatusCreated, performDonationRequest(t, h, createBody).Code)

			w := performUpdateDonationStatusRequest(t, h, id.String(), map[string]any{"status": "pending"})
			require.Equal(t, http.StatusConflict, w.Code)
			resp := test_util.DecodeJSON[errorResponse](t, w)
			require.Equal(t, "error", resp.Status)
		})

		t.Run("returns 422 for invalid status transition", func(t *testing.T) {
			id := uuid.New()
			createBody := map[string]any{
				"uuid":          id.String(),
				"amount":        2500,
				"currency":      "USD",
				"paymentMethod": "crypto",
				"nonprofitId":   "np_303",
				"donorId":       "donor_303",
				"status":        "new",
				"createdAt":     time.Now().UTC().Format(time.RFC3339),
			}
			require.Equal(t, http.StatusCreated, performDonationRequest(t, h, createBody).Code)

			w := performUpdateDonationStatusRequest(t, h, id.String(), map[string]any{"status": "success"})
			require.Equal(t, http.StatusUnprocessableEntity, w.Code)
			resp := test_util.DecodeJSON[errorResponse](t, w)
			require.Equal(t, "error", resp.Status)
			require.Contains(t, resp.Message, "invalid donation status transition")
		})

		t.Run("returns 404 when donation does not exist", func(t *testing.T) {
			w := performUpdateDonationStatusRequest(t, h, uuid.New().String(), map[string]any{"status": "pending"})
			require.Equal(t, http.StatusNotFound, w.Code)
			resp := test_util.DecodeJSON[errorResponse](t, w)
			require.Equal(t, "error", resp.Status)
		})
	})
}

type mockStore struct {
	listDonationsFunc       func() ([]model.Donation, error)
	getDonationFunc         func(uuid.UUID) (model.Donation, error)
	createDonationFunc      func() (model.Donation, error)
	updateDonationStatusFunc func(uuid.UUID, string) (model.Donation, error)
}

func (m mockStore) WithTx(_ context.Context, fn func(store.Store) error) error {
	return fn(m)
}

func (m mockStore) CreateDonation(_ context.Context, _ uuid.UUID, _ int64, _, _, _, _, _ string, _ time.Time) (model.Donation, error) {
	if m.createDonationFunc != nil {
		return m.createDonationFunc()
	}
	return model.Donation{}, nil
}

func (m mockStore) GetDonation(_ context.Context, id uuid.UUID) (model.Donation, error) {
	if m.getDonationFunc != nil {
		return m.getDonationFunc(id)
	}
	return model.Donation{}, nil
}

func (m mockStore) UpdateDonationStatus(_ context.Context, id uuid.UUID, status string) (model.Donation, error) {
	if m.updateDonationStatusFunc != nil {
		return m.updateDonationStatusFunc(id, status)
	}
	return model.Donation{}, nil
}

func (m mockStore) ListDonations(_ context.Context) ([]model.Donation, error) {
	if m.listDonationsFunc != nil {
		return m.listDonationsFunc()
	}
	return []model.Donation{}, nil
}

func TestHandler_ErrorAndValidationBranches(t *testing.T) {
	t.Run("ListDonations returns 500 on store error", func(t *testing.T) {
		h := handler.New("http://localhost:3000", mockStore{
			listDonationsFunc: func() ([]model.Donation, error) {
				return nil, errors.New("db down")
			},
		})
		c, w := test_util.NewGinContext(http.MethodGet, "/donations")
		h.ListDonations(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("GetDonation returns 400 on invalid uuid", func(t *testing.T) {
		h := handler.New("http://localhost:3000", mockStore{})
		c, w := test_util.NewGinContext(http.MethodGet, "/donations/bad")
		c.Params = gin.Params{{Key: "uuid", Value: "not-a-uuid"}}
		h.GetDonation(c)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetDonation returns 500 on unknown store error", func(t *testing.T) {
		h := handler.New("http://localhost:3000", mockStore{
			getDonationFunc: func(_ uuid.UUID) (model.Donation, error) {
				return model.Donation{}, errors.New("db boom")
			},
		})
		id := uuid.New()
		c, w := test_util.NewGinContext(http.MethodGet, "/donations/"+id.String())
		c.Params = gin.Params{{Key: "uuid", Value: id.String()}}
		h.GetDonation(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("CreateDonation returns 400 on malformed json", func(t *testing.T) {
		h := handler.New("http://localhost:3000", mockStore{})
		c, w := test_util.NewGinContext(http.MethodPost, "/donations")
		c.Request = httptest.NewRequest(http.MethodPost, "/donations", bytes.NewReader([]byte("{")))
		c.Request.Header.Set("Content-Type", "application/json")
		h.CreateDonation(c)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateDonation returns 400 on invalid createdAt", func(t *testing.T) {
		h := handler.New("http://localhost:3000", mockStore{})
		w := performDonationRequest(t, h, map[string]any{
			"uuid":          uuid.New().String(),
			"amount":        5000,
			"currency":      "USD",
			"paymentMethod": "cc",
			"nonprofitId":   "np",
			"donorId":       "donor",
			"status":        "new",
			"createdAt":     "not-a-date",
		})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateDonation returns 500 on unknown store error", func(t *testing.T) {
		h := handler.New("http://localhost:3000", mockStore{
			createDonationFunc: func() (model.Donation, error) {
				return model.Donation{}, errors.New("db boom")
			},
		})
		w := performDonationRequest(t, h, map[string]any{
			"uuid":          uuid.New().String(),
			"amount":        5000,
			"currency":      "USD",
			"paymentMethod": "cc",
			"nonprofitId":   "np",
			"donorId":       "donor",
			"status":        "new",
			"createdAt":     time.Now().UTC().Format(time.RFC3339),
		})
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("UpdateDonationStatus returns 400 on invalid uuid", func(t *testing.T) {
		h := handler.New("http://localhost:3000", mockStore{})
		w := performUpdateDonationStatusRequest(t, h, "invalid", map[string]any{"status": "pending"})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdateDonationStatus returns 400 on malformed json", func(t *testing.T) {
		h := handler.New("http://localhost:3000", mockStore{})
		id := uuid.New().String()
		c, w := test_util.NewGinContext(http.MethodPatch, "/donations/"+id+"/status")
		c.Request = httptest.NewRequest(http.MethodPatch, "/donations/"+id+"/status", bytes.NewReader([]byte("{")))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "uuid", Value: id}}
		h.UpdateDonationStatus(c)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdateDonationStatus returns 400 on invalid status", func(t *testing.T) {
		h := handler.New("http://localhost:3000", mockStore{})
		w := performUpdateDonationStatusRequest(t, h, uuid.New().String(), map[string]any{"status": "invalid"})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdateDonationStatus returns 500 on unknown store error", func(t *testing.T) {
		h := handler.New("http://localhost:3000", mockStore{
			updateDonationStatusFunc: func(_ uuid.UUID, _ string) (model.Donation, error) {
				return model.Donation{}, errors.New("db boom")
			},
		})
		w := performUpdateDonationStatusRequest(t, h, uuid.New().String(), map[string]any{"status": "pending"})
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func performDonationRequest(t *testing.T, h *handler.Handler, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	c, w := test_util.NewGinContext(http.MethodPost, "/donations")
	c.Request = httptest.NewRequest(http.MethodPost, "/donations", bytes.NewReader(payload))
	c.Request.Header.Set("Content-Type", "application/json")

	h.CreateDonation(c)
	return w
}

func performUpdateDonationStatusRequest(t *testing.T, h *handler.Handler, donationUUID string, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	c, w := test_util.NewGinContext(http.MethodPatch, "/donations/"+donationUUID+"/status")
	c.Request = httptest.NewRequest(http.MethodPatch, "/donations/"+donationUUID+"/status", bytes.NewReader(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "uuid", Value: donationUUID}}

	h.UpdateDonationStatus(c)
	return w
}
