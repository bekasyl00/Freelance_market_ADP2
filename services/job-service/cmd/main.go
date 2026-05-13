package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	jobgrpc "job_service/internal/delivery/grpc"
	natsdelivery "job_service/internal/delivery/nats"
	"job_service/internal/domain"
	"job_service/internal/repository"
	"job_service/internal/usecase"
	"job_service/proto"

	"github.com/nats-io/nats.go"
	_ "github.com/lib/pq"
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

	repo := repository.NewJobPostgresRepo(db)

	var pub domain.ApplicationPublisher
	if skipNATS() {
		log.Println("SKIP_NATS is set: NATS publishing disabled (dev mode)")
		pub = nil
	} else {
		natsURL := os.Getenv("NATS_URL")
		if natsURL == "" {
			natsURL = nats.DefaultURL
		}
		nc, err := nats.Connect(natsURL)
		if err != nil {
			log.Fatalf("nats connect (%s): %v — set SKIP_NATS=1 for local dev without NATS", natsURL, err)
		}
		defer nc.Drain()
		pub = natsdelivery.NewPublisher(nc)
	}

	uc := usecase.NewJobUsecase(repo, pub)
	handler := jobgrpc.NewJobHandler(uc)

	addr := listenAddr()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterJobServiceServer(s, handler)

	log.Printf("Job Service gRPC on %s", addr)
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
	return ":50052"
}

func skipNATS() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("SKIP_NATS")))
	return v == "1" || v == "true" || v == "yes"
}
