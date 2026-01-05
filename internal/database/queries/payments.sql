-- name: CreatePayment :one
INSERT INTO payments (amount, currency, reference, status)
VALUES ($1, $2, $3, $4)
RETURNING id, amount::float8, currency, reference, status, created_at, updated_at;

-- name: GetPaymentByID :one
SELECT id, amount::float8, currency, reference, status, created_at, updated_at
FROM payments
WHERE id = $1;

-- name: GetPaymentByReference :one
SELECT id, amount::float8, currency, reference, status, created_at, updated_at
FROM payments
WHERE reference = $1;

-- name: ProcessPaymentIdempotent :one
SELECT id, amount::float8, currency, reference, status, created_at, updated_at
FROM payments
WHERE id = $1
FOR UPDATE;

-- name: UpdatePaymentStatus :exec
UPDATE payments
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
