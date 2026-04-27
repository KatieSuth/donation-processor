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

func TestWithTx_CommitsOnSuccess(t *testing.T) {
	test_util.WithTestPool(t, func(s *store.PostgresStore) {
		err := s.WithTx(t.Context(), func(txStore store.Store) error {
			_, createErr := txStore.CreateDonation(
				t.Context(),
				uuid.New(),
				1200,
				"USD",
				"cc",
				"org_tx",
				"donor_tx",
				"new",
				time.Now().UTC(),
			)
			return createErr
		})
		require.NoError(t, err)
	})
}

func TestWithTx_RollsBackOnError(t *testing.T) {
	test_util.WithTestPool(t, func(s *store.PostgresStore) {
		sentinel := errors.New("boom")
		err := s.WithTx(t.Context(), func(_ store.Store) error {
			return sentinel
		})
		require.Error(t, err)
		require.ErrorIs(t, err, sentinel)
	})
}

func TestWithTx_InlineOnTransactionBackedStore(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		called := false
		err := s.WithTx(t.Context(), func(_ store.Store) error {
			called = true
			return nil
		})
		require.NoError(t, err)
		require.True(t, called)
	})
}

func TestNewPostgresStoreFromTx_ReturnsStore(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		require.NotNil(t, s)
	})
}
