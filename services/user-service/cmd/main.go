package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	usergrpc "user_service/internal/delivery/grpc"
	"user_service/internal/repository"
	"user_service/internal/usecase"
	"user_service/proto"

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

	// 2. Инициализация слоев (Dependency Injection)
	repo := repository.NewUserPostgresRepo(db)
	logic := usecase.NewUserUsecase(repo)
	handler := usergrpc.NewUserHandler(logic)

	// 3. Запуск gRPC сервера
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterUserServiceServer(s, handler)

	log.Println("User Service is running on port :50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
