package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"payment-gateway/internal/config"
	"payment-gateway/internal/database"
	"payment-gateway/internal/rabbitmq"
	"payment-gateway/pkg/worker"
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

	paymentWorker := worker.NewPaymentWorker(paymentRepo, rabbitClient)

	go func() {
		if err := paymentWorker.Start(); err != nil {
			log.Fatalf("Failed to start worker: %v", err)
		}
	}()

	log.Println("Payment worker started. Press Ctrl+C to exit.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Println("Shutting down worker...")
}
