// Package models
package models

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	StatusPending PaymentStatus = "PENDING"
	StatusSuccess PaymentStatus = "SUCCESS"
	StatusFailed  PaymentStatus = "FAILED"
)

type Payment struct {
	ID        uuid.UUID     `json:"id" db:"id"`
	Amount    float64       `json:"amount" db:"amount"`
	Currency  string        `json:"currency" db:"currency"`
	Reference string        `json:"reference" db:"reference"`
	Status    PaymentStatus `json:"status" db:"status"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" db:"updated_at"`
}

type CreatePaymentRequest struct {
	Amount    float64 `json:"amount" validate:"required,gt=0"`
	Currency  string  `json:"currency" validate:"required,oneof=ETB USD"`
	Reference string  `json:"reference" validate:"required,max=255"`
}

type CreatePaymentResponse struct {
	ID     uuid.UUID     `json:"id"`
	Status PaymentStatus `json:"status"`
}

type GetPaymentResponse struct {
	ID        uuid.UUID     `json:"id"`
	Amount    float64       `json:"amount"`
	Currency  string        `json:"currency"`
	Reference string        `json:"reference"`
	Status    PaymentStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
}

type ProcessPaymentMessage struct {
	PaymentID uuid.UUID `json:"payment_id"`
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
	Code    int               `json:"code,omitempty"`
}
