package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	pb "go-proj/example/servicepb"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedHelloServiceServer
	db *sql.DB
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("Received: %s", req.Name)

	// Записываем имя в базу данных
	_, err := s.db.Exec("INSERT INTO greetings (name) VALUES ($1)", req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to insert into database: %v", err)
	}

	return &pb.HelloResponse{Message: "Hello, " + req.Name}, nil
}

func connectToDB(connStr string, maxRetries int, retryDelay time.Duration) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("Successfully connected to the database")
				return db, nil
			}
		}

		log.Printf("Failed to connect to database, retrying (%d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("could not connect to the database after %d retries: %v", maxRetries, err)
}

func main() {
	// Настройка подключения к базе данных
	connStr := "postgres://myuser:mypassword@db:5432/mydb?sslmode=disable"
	maxRetries := 5
	retryDelay := 2 * time.Second

	// Подключаемся к базе данных с проверкой готовности
	db, err := connectToDB(connStr, maxRetries, retryDelay)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Создаём таблицу, если её ещё нет
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS greetings (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL
        )
    `)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Настройка gRPC сервера
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterHelloServiceServer(grpcServer, &server{db: db})

	log.Println("Server is listening on port 50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
