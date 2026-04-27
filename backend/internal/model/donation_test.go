package model_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/donation-processor/backend/internal/db"
	"github.com/KatieSuth/donation-processor/backend/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMapDbDonationToDonation(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now().UTC().Truncate(time.Second)
	updatedAt := createdAt.Add(2 * time.Minute)

	dbDonation := db.Donation{
		Uuid:          id,
		Amount:        5000,
		Currency:      "USD",
		PaymentMethod: "cc",
		NonprofitID:   "np_123",
		DonorID:       "donor_123",
		Status:        "new",
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}

	got := model.MapDbDonationToDonation(dbDonation)

	require.Equal(t, id, got.UUID)
	require.Equal(t, int64(5000), got.Amount)
	require.Equal(t, "USD", got.Currency)
	require.Equal(t, "cc", got.PaymentMethod)
	require.Equal(t, "np_123", got.NonprofitID)
	require.Equal(t, "donor_123", got.DonorID)
	require.Equal(t, "new", got.Status)
	require.Equal(t, createdAt, got.CreatedAt)
	require.Equal(t, updatedAt, got.UpdatedAt)
}

func TestMapDbDonationsToDonations(t *testing.T) {
	first := db.Donation{
		Uuid:          uuid.New(),
		Amount:        1000,
		Currency:      "USD",
		PaymentMethod: "cc",
		NonprofitID:   "org_a",
		DonorID:       "donor_a",
		Status:        "new",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	second := db.Donation{
		Uuid:          uuid.New(),
		Amount:        2500,
		Currency:      "USD",
		PaymentMethod: "ach",
		NonprofitID:   "org_b",
		DonorID:       "donor_b",
		Status:        "pending",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	out := model.MapDbDonationsToDonations([]db.Donation{first, second})
	require.Len(t, out, 2)
	require.Equal(t, first.Uuid, out[0].UUID)
	require.Equal(t, second.Uuid, out[1].UUID)
}
