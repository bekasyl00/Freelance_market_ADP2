package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	paymentgrpc "payment_service/internal/delivery/grpc"
	natsdelivery "payment_service/internal/delivery/nats"
	"payment_service/internal/models"
	"payment_service/internal/repository"
	"payment_service/internal/usecase"
	"payment_service/proto"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
)

const defaultDatabaseURL = "postgresql://freelance:freelance@localhost:5433/freelance_market?sslmode=disable"

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultDatabaseURL
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	repo := repository.NewPaymentPostgresRepo(db)

	var pub models.EventPublisher
	if skipNATS() {
		log.Println("SKIP_NATS is set: payment events disabled")
	} else {
		natsURL := os.Getenv("NATS_URL")
		if natsURL == "" {
			natsURL = nats.DefaultURL
		}
		nc, err := nats.Connect(natsURL)
		if err != nil {
			log.Fatalf("nats connect (%s): %v - set SKIP_NATS=1 for local dev without NATS", natsURL, err)
		}
		defer nc.Drain()
		pub = natsdelivery.NewPublisher(nc)
	}

	uc := usecase.NewPaymentUsecase(repo, pub)
	handler := paymentgrpc.NewPaymentHandler(uc)

	lis, err := net.Listen("tcp", listenAddr())
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterPaymentServiceServer(s, handler)

	log.Printf("Payment Service gRPC on %s", listenAddr())
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func listenAddr() string {
	if p := os.Getenv("GRPC_PORT"); p != "" {
		if _, err := strconv.Atoi(p); err == nil {
			return ":" + p
		}
	}
	return ":50053"
}

func skipNATS() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("SKIP_NATS")))
	return v == "1" || v == "true" || v == "yes"
}
