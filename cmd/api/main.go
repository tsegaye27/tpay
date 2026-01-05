package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "payment-gateway/docs"
	"payment-gateway/internal/config"
	"payment-gateway/internal/database"
	"payment-gateway/internal/rabbitmq"
	"payment-gateway/internal/server"
	"payment-gateway/pkg/handlers"
)

func main() {
	cfg := config.Load()

	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	paymentRepo := database.NewPaymentRepository(db)

	rabbitClient, err := rabbitmq.NewClient(cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitClient.Close()

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	paymentHandler := handlers.NewPaymentHandler(paymentRepo, rabbitClient)

	server.SetupRoutes(e, paymentHandler)

	log.Printf("Starting API server on port %d", cfg.API.Port)
	if err := e.Start(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
