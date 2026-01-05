// Package server
package server

import (
	"github.com/labstack/echo/v4"
	"payment-gateway/pkg/handlers"
)

func SetupRoutes(e *echo.Echo, paymentHandler *handlers.PaymentHandler) {
	e.GET("/health", healthCheck)

	api := e.Group("/api/v1")

	setupPaymentRoutes(api, paymentHandler)
}

func healthCheck(c echo.Context) error {
	return c.JSON(200, map[string]string{"status": "healthy"})
}

func setupPaymentRoutes(api *echo.Group, paymentHandler *handlers.PaymentHandler) {
	payments := api.Group("/payments")
	payments.POST("", paymentHandler.CreatePayment)
	payments.GET("/:id", paymentHandler.GetPayment)
}
