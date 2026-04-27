-- name: CreateDonation :one
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
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
RETURNING *;

-- name: ListDonations :many
SELECT *
FROM donations
ORDER BY created_at DESC, uuid DESC;

-- name: GetDonation :one
SELECT *
FROM donations
WHERE uuid = $1;

-- name: UpdateDonationStatus :one
UPDATE donations
SET
    status = $2,
    updated_at = $3
WHERE uuid = $1
RETURNING *;
