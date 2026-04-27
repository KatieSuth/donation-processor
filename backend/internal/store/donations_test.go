package store_test

import (
	"errors"
	"testing"
	"time"

	"github.com/KatieSuth/donation-processor/backend/internal/db"
	"github.com/KatieSuth/donation-processor/backend/internal/store"
	"github.com/KatieSuth/donation-processor/backend/internal/test_util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateDonation_AndConflict(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		id := uuid.New()
		createdAt := time.Now().UTC().Truncate(time.Second)

		created, err := s.CreateDonation(
			t.Context(),
			id,
			5000,
			"USD",
			"cc",
			"org_1",
			"donor_1",
			"new",
			createdAt,
		)
		require.NoError(t, err)
		require.Equal(t, id, created.UUID)
		require.Equal(t, int64(5000), created.Amount)
		require.Equal(t, "new", created.Status)

		_, err = s.CreateDonation(
			t.Context(),
			id,
			5000,
			"USD",
			"cc",
			"org_1",
			"donor_1",
			"new",
			createdAt,
		)
		require.Error(t, err)
		require.True(t, errors.Is(err, store.ErrDonationConflict))
	})
}

func TestListAndGetDonations(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		idA := uuid.New()
		idB := uuid.New()
		createdAt := time.Now().UTC().Truncate(time.Second)

		_, err := s.CreateDonation(t.Context(), idA, 1000, "USD", "ach", "org_a", "donor_a", "new", createdAt)
		require.NoError(t, err)
		_, err = s.CreateDonation(t.Context(), idB, 2000, "USD", "venmo", "org_b", "donor_b", "pending", createdAt.Add(time.Minute))
		require.NoError(t, err)

		listed, err := s.ListDonations(t.Context())
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(listed), 2)
		foundA := false
		foundB := false
		for _, donation := range listed {
			if donation.UUID == idA {
				foundA = true
			}
			if donation.UUID == idB {
				foundB = true
			}
		}
		require.True(t, foundA)
		require.True(t, foundB)

		got, err := s.GetDonation(t.Context(), idA)
		require.NoError(t, err)
		require.Equal(t, idA, got.UUID)

		_, err = s.GetDonation(t.Context(), uuid.New())
		require.Error(t, err)
		require.True(t, errors.Is(err, store.ErrDonationNotFound))
	})
}

func TestUpdateDonationStatus(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		id := uuid.New()
		createdAt := time.Now().UTC().Truncate(time.Second)
		_, err := s.CreateDonation(t.Context(), id, 1500, "USD", "cc", "org", "donor", "new", createdAt)
		require.NoError(t, err)

		updated, err := s.UpdateDonationStatus(t.Context(), id, "pending")
		require.NoError(t, err)
		require.Equal(t, "pending", updated.Status)

		_, err = s.UpdateDonationStatus(t.Context(), id, "pending")
		require.Error(t, err)
		require.True(t, errors.Is(err, store.ErrDonationStatusConflict))

		_, err = s.UpdateDonationStatus(t.Context(), id, "new")
		require.Error(t, err)
		require.True(t, errors.Is(err, store.ErrInvalidDonationStatusTransition))

		_, err = s.UpdateDonationStatus(t.Context(), uuid.New(), "pending")
		require.Error(t, err)
		require.True(t, errors.Is(err, store.ErrDonationNotFound))
	})
}

func TestIsValidStatusTransition(t *testing.T) {
	require.True(t, store.ExportIsValidStatusTransition("new", "pending"))
	require.True(t, store.ExportIsValidStatusTransition("pending", "success"))
	require.True(t, store.ExportIsValidStatusTransition("pending", "failure"))
	require.False(t, store.ExportIsValidStatusTransition("new", "success"))
	require.False(t, store.ExportIsValidStatusTransition("success", "failure"))
}
