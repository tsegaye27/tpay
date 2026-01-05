// Package handlers
package handlers

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"payment-gateway/internal/database"
	"payment-gateway/internal/models"
	"payment-gateway/internal/rabbitmq"
)

type PaymentHandler struct {
	repo     *database.PaymentRepository
	rabbitMQ *rabbitmq.Client
	validate *validator.Validate
}

func NewPaymentHandler(repo *database.PaymentRepository, rabbitMQ *rabbitmq.Client) *PaymentHandler {
	return &PaymentHandler{
		repo:     repo,
		rabbitMQ: rabbitMQ,
		validate: validator.New(),
	}
}

// CreatePayment handles POST /payments - creates a new payment
// @Summary Create a new payment
// @Description Create a payment with amount, currency, and unique reference. The payment will be processed asynchronously.
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body models.CreatePaymentRequest true "Payment creation request"
// @Success 201 {object} models.CreatePaymentResponse "Payment created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request body or validation failed"
// @Failure 409 {object} models.ErrorResponse "Payment with this reference already exists"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/payments [post]
func (h *PaymentHandler) CreatePayment(c echo.Context) error {
	req := new(models.CreatePaymentRequest)

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	existing, err := h.repo.GetPaymentByReference(req.Reference)
	if err == nil && existing != nil {
		return c.JSON(http.StatusConflict, map[string]string{
			"error": "Payment with this reference already exists",
		})
	}

	payment, err := h.repo.CreatePayment(*req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create payment",
		})
	}

	if err := h.rabbitMQ.PublishPaymentProcessing(payment.ID); err != nil {
		c.Logger().Errorf("Failed to publish payment processing message: %v", err)
	}

	response := models.CreatePaymentResponse{
		ID:     payment.ID,
		Status: payment.Status,
	}

	return c.JSON(http.StatusCreated, response)
}

// GetPayment handles GET /payments/{id} - retrieves payment details
// @Summary Get payment by ID
// @Description Retrieve payment details by payment ID
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID" format(uuid)
// @Success 200 {object} models.GetPaymentResponse "Payment details"
// @Failure 400 {object} models.ErrorResponse "Invalid payment ID format"
// @Failure 404 {object} models.ErrorResponse "Payment not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/payments/{id} [get]
func (h *PaymentHandler) GetPayment(c echo.Context) error {
	idParam := c.Param("id")
	paymentID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid payment ID format",
		})
	}

	payment, err := h.repo.GetPaymentByID(paymentID)
	if err != nil {
		if err.Error() == "payment not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Payment not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve payment",
		})
	}

	response := models.GetPaymentResponse{
		ID:        payment.ID,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		Reference: payment.Reference,
		Status:    payment.Status,
		CreatedAt: payment.CreatedAt,
	}

	return c.JSON(http.StatusOK, response)
}
