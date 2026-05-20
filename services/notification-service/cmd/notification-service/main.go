package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"freelance_market_adp2/nats_worker/email"
	natsdelivery "notification_service/internal/delivery/nats"

	"github.com/nats-io/nats.go"
)

func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("nats connect (%s): %v", natsURL, err)
	}
	defer nc.Drain()

	// Initialize Email Service
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Fatalf("Invalid SMTP_PORT: %v", err)
	}
	emailCfg := email.Config{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     smtpPort,
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     os.Getenv("SMTP_FROM"),
	}
	emailSvc := email.NewSMTPService(emailCfg)

	// Initialize Worker
	// In a real app, 'uc' would be a usecase implementation.
	// Here we wrap the worker for the subscriber to consume.
	emailWorker := email.NewWorker(nc, emailSvc)

	// Assuming NewSubscriber takes a function or interface that matches HandleMessage
	// Adjust 'uc' according to your natsdelivery implementation
	sub := natsdelivery.NewSubscriber(nc, emailWorker)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := sub.Subscribe(ctx); err != nil {
		log.Fatalf("subscribe: %v", err)
	}
	log.Println("Notification Service subscribed to NATS events")
	<-ctx.Done()
	log.Println("Notification Service shutting down")
}
