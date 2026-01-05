// Package rabbitmq
package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"payment-gateway/internal/config"
	"payment-gateway/internal/models"
)

type Client struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  config.RabbitMQConfig
}

func NewClient(cfg config.RabbitMQConfig) (*Client, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.VHost)

	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	client := &Client{
		conn:    conn,
		channel: channel,
		config:  cfg,
	}

	if err := client.declareQueue(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	log.Println("Successfully connected to RabbitMQ")
	return client, nil
}

func (c *Client) declareQueue() error {
	queueName := "payment_processing"

	_, err := c.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	log.Printf("Declared queue: %s", queueName)
	return nil
}

func (c *Client) PublishPaymentProcessing(paymentID uuid.UUID) error {
	queueName := "payment_processing"

	message := models.ProcessPaymentMessage{
		PaymentID: paymentID,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = c.channel.Publish(
		"",
		queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published payment processing message for payment ID: %s", paymentID)
	return nil
}

func (c *Client) ConsumePaymentProcessing(handler func(models.ProcessPaymentMessage) error) error {
	queueName := "payment_processing"

	msgs, err := c.channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("Started consuming messages from queue: %s", queueName)

	go func() {
		for d := range msgs {
			var message models.ProcessPaymentMessage
			if err := json.Unmarshal(d.Body, &message); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				d.Nack(false, false)
				continue
			}

			if err := handler(message); err != nil {
				log.Printf("Failed to process payment %s: %v", message.PaymentID, err)

				time.Sleep(time.Second * 5)
				d.Nack(false, true)
				continue
			}

			d.Ack(false)
			log.Printf("Successfully processed payment: %s", message.PaymentID)
		}
	}()

	return nil
}

func (c *Client) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}
	log.Println("RabbitMQ connection closed")
	return nil
}
