package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	natsdelivery "notification_service/internal/delivery/nats"
	// Create later
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
