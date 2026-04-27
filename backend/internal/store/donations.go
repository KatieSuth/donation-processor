package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KatieSuth/donation-processor/backend/internal/db"
	"github.com/KatieSuth/donation-processor/backend/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq/pqerror"
)

var ErrDonationConflict = errors.New("donation uuid already exists")
var ErrDonationNotFound = errors.New("donation not found")
var ErrDonationStatusConflict = errors.New("target status is identical to current status")
var ErrInvalidDonationStatusTransition = errors.New("invalid donation status transition")

func (s *PostgresStore) CreateDonation(
	ctx context.Context,
	donationUUID uuid.UUID,
	amount int64,
	currency string,
	paymentMethod string,
	nonprofitID string,
	donorID string,
	status string,
	createdAt time.Time,
) (model.Donation, error) {
	donation, err := s.q.CreateDonation(ctx, db.CreateDonationParams{
		Uuid:          donationUUID,
		Amount:        amount,
		Currency:      currency,
		PaymentMethod: paymentMethod,
		NonprofitID:   nonprofitID,
		DonorID:       donorID,
		Status:        status,
		CreatedAt:     createdAt,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == string(pqerror.UniqueViolation) {
			return model.Donation{}, ErrDonationConflict
		}
		return model.Donation{}, fmt.Errorf("create donation: %w", err)
	}

	return model.MapDbDonationToDonation(donation), nil
}

func (s *PostgresStore) ListDonations(ctx context.Context) ([]model.Donation, error) {
	donations, err := s.q.ListDonations(ctx)
	if err != nil {
		return nil, fmt.Errorf("list donations: %w", err)
	}

	return model.MapDbDonationsToDonations(donations), nil
}

func (s *PostgresStore) GetDonation(ctx context.Context, donationUUID uuid.UUID) (model.Donation, error) {
	donation, err := s.q.GetDonation(ctx, donationUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Donation{}, ErrDonationNotFound
		}
		return model.Donation{}, fmt.Errorf("get donation: %w", err)
	}

	return model.MapDbDonationToDonation(donation), nil
}

func (s *PostgresStore) UpdateDonationStatus(ctx context.Context, donationUUID uuid.UUID, targetStatus string) (model.Donation, error) {
	currentDonation, err := s.q.GetDonation(ctx, donationUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Donation{}, ErrDonationNotFound
		}
		return model.Donation{}, fmt.Errorf("get donation before status update: %w", err)
	}

	currentStatus := currentDonation.Status
	if currentStatus == targetStatus {
		return model.Donation{}, ErrDonationStatusConflict
	}
	if !isValidStatusTransition(currentStatus, targetStatus) {
		return model.Donation{}, fmt.Errorf(
			"%w: %s -> %s",
			ErrInvalidDonationStatusTransition,
			currentStatus,
			targetStatus,
		)
	}

	updatedDonation, err := s.q.UpdateDonationStatus(ctx, db.UpdateDonationStatusParams{
		Uuid:      donationUUID,
		Status:    targetStatus,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return model.Donation{}, fmt.Errorf("update donation status: %w", err)
	}

	return model.MapDbDonationToDonation(updatedDonation), nil
}

func isValidStatusTransition(currentStatus, targetStatus string) bool {
	switch currentStatus {
	case "new":
		return targetStatus == "pending"
	case "pending":
		return targetStatus == "success" || targetStatus == "failure"
	default:
		return false
	}
}
