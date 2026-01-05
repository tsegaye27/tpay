// Package worker
package worker

import (
	"log"

	"payment-gateway/internal/database"
	"payment-gateway/internal/models"
	"payment-gateway/internal/rabbitmq"
)

type PaymentWorker struct {
	repo     *database.PaymentRepository
	rabbitMQ *rabbitmq.Client
}

func NewPaymentWorker(repo *database.PaymentRepository, rabbitMQ *rabbitmq.Client) *PaymentWorker {
	return &PaymentWorker{
		repo:     repo,
		rabbitMQ: rabbitMQ,
	}
}

func (w *PaymentWorker) Start() error {
	log.Println("Starting payment worker...")

	return w.rabbitMQ.ConsumePaymentProcessing(w.processPayment)
}

func (w *PaymentWorker) processPayment(message models.ProcessPaymentMessage) error {
	log.Printf("Processing payment: %s", message.PaymentID)

	payment, err := w.repo.ProcessPaymentIdempotent(message.PaymentID)
	if err != nil {
		log.Printf("Failed to process payment %s: %v", message.PaymentID, err)
		return err
	}

	log.Printf("Successfully processed payment %s with status: %s", payment.ID, payment.Status)
	return nil
}
