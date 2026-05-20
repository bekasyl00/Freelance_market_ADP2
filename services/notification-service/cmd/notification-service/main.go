package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"notification_service/internal/mailer"
	"notification_service/internal/usecase"

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

	// Initialize SMTP Mailer
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Fatalf("Invalid SMTP_PORT: %v", err)
	}
	m := mailer.NewSMTP(mailer.Config{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     smtpPort,
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     os.Getenv("SMTP_FROM"),
	})

	// Initialize Usecase and Subscriber
	uc := usecase.NewNotificationUsecase(m)
	sub := natsdelivery.NewSubscriber(nc, uc)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := sub.Subscribe(ctx); err != nil {
		log.Fatalf("subscribe: %v", err)
	}
	log.Println("Notification Service subscribed to NATS events")
	<-ctx.Done()
	log.Println("Notification Service shutting down")
}
