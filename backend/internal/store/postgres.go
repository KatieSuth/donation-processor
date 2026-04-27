// Package store is the data access layer: PostgreSQL via sqlc-generated queries, transactions,
// and store-level validation errors that handlers map to HTTP status codes.
package store

import (
	"context"
	"time"

	"github.com/KatieSuth/donation-processor/backend/internal/db"
	"github.com/KatieSuth/donation-processor/backend/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	q    *db.Queries
	pool *pgxpool.Pool // store the pool directly so we can begin transactions
}

// Store abstracts persistence for handlers and allows transactional tests to swap in a
// transaction-backed store via WithTx and NewPostgresStoreFromTx.
type Store interface {
	WithTx(ctx context.Context, fn func(Store) error) error

	CreateDonation(ctx context.Context, donationUUID uuid.UUID, amount int64, currency, paymentMethod, nonprofitID, donorID, status string, createdAt time.Time) (model.Donation, error)
	GetDonation(ctx context.Context, donationUUID uuid.UUID) (model.Donation, error)
	UpdateDonationStatus(ctx context.Context, donationUUID uuid.UUID, targetStatus string) (model.Donation, error)
	ListDonations(ctx context.Context) ([]model.Donation, error)
}

// NewPostgresStore wires a connection pool. Callers are responsible for pool lifecycle.
func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{
		q:    db.New(pool),
		pool: pool,
	}
}

func (s *PostgresStore) WithTx(ctx context.Context, fn func(Store) error) error {
	// If this store is already transaction-backed (tests commonly construct it via
	// NewPostgresStoreFromTx), execute inline so outer transaction controls commit/rollback.
	if s.pool == nil {
		return fn(s)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // no-op if already committed

	txStore := &PostgresStore{
		q:    db.New(tx),
		pool: s.pool,
	}

	if err := fn(txStore); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// For use in tests only — no pool, so WithTx executes inline against this transaction.
func NewPostgresStoreFromTx(tx pgx.Tx) *PostgresStore {
	return &PostgresStore{
		q: db.New(tx),
	}
}
