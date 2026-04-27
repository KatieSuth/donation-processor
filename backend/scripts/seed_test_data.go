package scripts

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type seedDonation struct {
	UUID          string
	Amount        int64
	Currency      string
	PaymentMethod string
	NonprofitID   string
	DonorID       string
	Status        string
	CreatedAt     string
	UpdatedAt     string
}

var donationSeedData = []seedDonation{
	{
		UUID:          "354362d8-2080-4ca1-9ede-892e4c6d3a25",
		Amount:        5000,
		Currency:      "USD",
		PaymentMethod: "cc",
		NonprofitID:   "org1",
		DonorID:       "donor_01",
		Status:        "new",
		CreatedAt:     "2026-01-15T10:00:00Z",
		UpdatedAt:     "2026-01-15T10:00:00Z",
	},
	{
		UUID:          "385cf5cb-9d0a-4f9e-948b-412791755060",
		Amount:        10000,
		Currency:      "USD",
		PaymentMethod: "ach",
		NonprofitID:   "org2",
		DonorID:       "donor_02",
		Status:        "new",
		CreatedAt:     "2026-01-15T10:05:00Z",
		UpdatedAt:     "2026-01-15T10:05:00Z",
	},
	{
		UUID:          "86f1c108-102b-447a-9efe-67c2f3b594d8",
		Amount:        25000,
		Currency:      "USD",
		PaymentMethod: "crypto",
		NonprofitID:   "org1",
		DonorID:       "donor_03",
		Status:        "pending",
		CreatedAt:     "2026-01-15T10:10:00Z",
		UpdatedAt:     "2026-01-15T10:12:34Z",
	},
	{
		UUID:          "c9b7c4c1-2ca7-465c-bf34-2a80ee7534eb",
		Amount:        1500,
		Currency:      "USD",
		PaymentMethod: "venmo",
		NonprofitID:   "org3",
		DonorID:       "donor_04",
		Status:        "pending",
		CreatedAt:     "2026-01-15T10:15:00Z",
		UpdatedAt:     "2026-01-15T10:17:08Z",
	},
	{
		UUID:          "73aff4cc-135d-4840-96b2-9210639528c8",
		Amount:        7500,
		Currency:      "USD",
		PaymentMethod: "cc",
		NonprofitID:   "org2",
		DonorID:       "donor_05",
		Status:        "success",
		CreatedAt:     "2026-01-15T10:20:00Z",
		UpdatedAt:     "2026-01-15T10:21:47Z",
	},
	{
		UUID:          "7b789658-cb91-4ae6-bbb6-5cb90a1b1942",
		Amount:        3000,
		Currency:      "USD",
		PaymentMethod: "ach",
		NonprofitID:   "org3",
		DonorID:       "donor_06",
		Status:        "failure",
		CreatedAt:     "2026-01-15T10:25:00Z",
		UpdatedAt:     "2026-01-15T10:43:22Z",
	},
	{
		UUID:          "4619db6e-5ddf-4900-9da2-17e55e400ca4",
		Amount:        15000,
		Currency:      "USD",
		PaymentMethod: "crypto",
		NonprofitID:   "org1",
		DonorID:       "donor_07",
		Status:        "new",
		CreatedAt:     "2026-01-15T10:30:00Z",
		UpdatedAt:     "2026-01-15T10:30:00Z",
	},
	{
		UUID:          "49ce76af-3134-40e3-99d6-b6e3d7e51de5",
		Amount:        20000,
		Currency:      "USD",
		PaymentMethod: "venmo",
		NonprofitID:   "org2",
		DonorID:       "donor_08",
		Status:        "pending",
		CreatedAt:     "2026-01-15T10:35:00Z",
		UpdatedAt:     "2026-01-15T10:38:51Z",
	},
}

// SeedTestData inserts deterministic donation fixtures for debug usage only.
// Existing rows are preserved via ON CONFLICT DO NOTHING.
func SeedTestData(ctx context.Context, pool *pgxpool.Pool) error {
	const insertDonation = `
		INSERT INTO donations (
			uuid,
			amount,
			currency,
			payment_method,
			nonprofit_id,
			donor_id,
			status,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (uuid) DO NOTHING
	`

	for _, d := range donationSeedData {
		createdAt, err := time.Parse(time.RFC3339, d.CreatedAt)
		if err != nil {
			return fmt.Errorf("parse createdAt for %s: %w", d.UUID, err)
		}
		updatedAt, err := time.Parse(time.RFC3339, d.UpdatedAt)
		if err != nil {
			return fmt.Errorf("parse updatedAt for %s: %w", d.UUID, err)
		}

		if _, err := pool.Exec(
			ctx,
			insertDonation,
			d.UUID,
			d.Amount,
			d.Currency,
			d.PaymentMethod,
			d.NonprofitID,
			d.DonorID,
			d.Status,
			createdAt,
			updatedAt,
		); err != nil {
			return fmt.Errorf("seed donation %s: %w", d.UUID, err)
		}
	}

	return nil
}
