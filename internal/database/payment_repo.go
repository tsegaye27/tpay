package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"payment-gateway/internal/models"
)

type PaymentRepository struct {
	db      *sql.DB
	queries *Queries
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{
		db:      db,
		queries: New(db),
	}
}

func (r *PaymentRepository) CreatePayment(req models.CreatePaymentRequest) (*models.Payment, error) {
	ctx := context.Background()

	sqlcPayment, err := r.queries.CreatePayment(ctx, CreatePaymentParams{
		Amount:    req.Amount,
		Currency:  req.Currency,
		Reference: req.Reference,
		Status:    string(models.StatusPending),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	payment := &models.Payment{
		ID:        sqlcPayment.ID,
		Amount:    sqlcPayment.Amount,
		Currency:  sqlcPayment.Currency,
		Reference: sqlcPayment.Reference,
		Status:    models.PaymentStatus(sqlcPayment.Status),
		CreatedAt: sqlcPayment.CreatedAt.Time,
		UpdatedAt: sqlcPayment.UpdatedAt.Time,
	}

	return payment, nil
}

func (r *PaymentRepository) GetPaymentByID(id uuid.UUID) (*models.Payment, error) {
	ctx := context.Background()

	sqlcPayment, err := r.queries.GetPaymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	payment := &models.Payment{
		ID:        sqlcPayment.ID,
		Amount:    sqlcPayment.Amount,
		Currency:  sqlcPayment.Currency,
		Reference: sqlcPayment.Reference,
		Status:    models.PaymentStatus(sqlcPayment.Status),
		CreatedAt: sqlcPayment.CreatedAt.Time,
		UpdatedAt: sqlcPayment.UpdatedAt.Time,
	}

	return payment, nil
}

func (r *PaymentRepository) GetPaymentByReference(reference string) (*models.Payment, error) {
	ctx := context.Background()

	sqlcPayment, err := r.queries.GetPaymentByReference(ctx, reference)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment by reference: %w", err)
	}

	payment := &models.Payment{
		ID:        sqlcPayment.ID,
		Amount:    sqlcPayment.Amount,
		Currency:  sqlcPayment.Currency,
		Reference: sqlcPayment.Reference,
		Status:    models.PaymentStatus(sqlcPayment.Status),
		CreatedAt: sqlcPayment.CreatedAt.Time,
		UpdatedAt: sqlcPayment.UpdatedAt.Time,
	}

	return payment, nil
}

func (r *PaymentRepository) ProcessPaymentIdempotent(paymentID uuid.UUID) (*models.Payment, error) {
	ctx := context.Background()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	txQueries := r.queries.WithTx(tx)

	sqlcPayment, err := txQueries.ProcessPaymentIdempotent(ctx, paymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment for processing: %w", err)
	}

	currentStatus := models.PaymentStatus(sqlcPayment.Status)
	if currentStatus != models.StatusPending {

		payment := &models.Payment{
			ID:        sqlcPayment.ID,
			Amount:    sqlcPayment.Amount,
			Currency:  sqlcPayment.Currency,
			Reference: sqlcPayment.Reference,
			Status:    currentStatus,
			CreatedAt: sqlcPayment.CreatedAt.Time,
			UpdatedAt: sqlcPayment.UpdatedAt.Time,
		}
		return payment, nil
	}

	newStatus := models.StatusSuccess

	err = txQueries.UpdatePaymentStatus(ctx, UpdatePaymentStatusParams{
		Status: string(newStatus),
		ID:     paymentID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	payment := &models.Payment{
		ID:        sqlcPayment.ID,
		Amount:    sqlcPayment.Amount,
		Currency:  sqlcPayment.Currency,
		Reference: sqlcPayment.Reference,
		Status:    newStatus,
		CreatedAt: sqlcPayment.CreatedAt.Time,
		UpdatedAt: sqlcPayment.UpdatedAt.Time,
	}

	return payment, nil
}
