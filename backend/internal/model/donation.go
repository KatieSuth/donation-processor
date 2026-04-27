package model

import (
	"time"

	"github.com/KatieSuth/donation-processor/backend/internal/db"
	"github.com/google/uuid"
)

type Donation struct {
	UUID          uuid.UUID `json:"uuid"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	PaymentMethod string    `json:"paymentMethod"`
	NonprofitID   string    `json:"nonprofitId"`
	DonorID       string    `json:"donorId"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func MapDbDonationToDonation(dbDonation db.Donation) Donation {
	return Donation{
		UUID:          dbDonation.Uuid,
		Amount:        dbDonation.Amount,
		Currency:      dbDonation.Currency,
		PaymentMethod: dbDonation.PaymentMethod,
		NonprofitID:   dbDonation.NonprofitID,
		DonorID:       dbDonation.DonorID,
		Status:        dbDonation.Status,
		CreatedAt:     dbDonation.CreatedAt,
		UpdatedAt:     dbDonation.UpdatedAt,
	}
}

func MapDbDonationsToDonations(dbDonations []db.Donation) []Donation {
	donations := make([]Donation, 0, len(dbDonations))
	for _, dbDonation := range dbDonations {
		donations = append(donations, MapDbDonationToDonation(dbDonation))
	}
	return donations
}
