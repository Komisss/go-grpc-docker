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

// Метод SayHello, который возвращает приветствие на выбранном языке
func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("Received: Name=%s, Language=%s", req.Name, req.Language)

	// Получаем приветствие из базы данных
	var greeting string
	err := s.db.QueryRow("SELECT greeting FROM greetings WHERE language = $1", req.Language).Scan(&greeting)
	if err == sql.ErrNoRows {
		// Если язык не найден, возвращаем приветствие на русском
		greeting = "Привет"
	} else if err != nil {
		return nil, fmt.Errorf("failed to query database: %v", err)
	}

	// Формируем ответ
	message := fmt.Sprintf("%s, %s!", greeting, req.Name)
	return &pb.HelloResponse{Message: message}, nil
}

// Ожидание доступности базы данных
func waitForDB(db *sql.DB) error {
	for i := 0; i < 10; i++ { // До 10 попыток подключения
		err := db.Ping()
		if err == nil {
			return nil
		}
		log.Printf("Database not ready, retrying in 2 seconds...")
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("database not reachable")
}

func main() {
	// Настройка подключения к базе данных
	connStr := "postgres://myuser:mypassword@db:5432/mydb?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ждём доступности базы данных
	if err := waitForDB(db); err != nil {
		log.Fatalf("Database not reachable: %v", err)
	}

	// Создаём таблицу, если её ещё нет
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS greetings (
            id SERIAL PRIMARY KEY,
            language TEXT NOT NULL,
            greeting TEXT NOT NULL
        )
    `)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Заполняем базу данными, если она пустая
	_, err = db.Exec(`
        INSERT INTO greetings (language, greeting) VALUES
        ('English', 'Hello'),
        ('Russian', 'Привет'),
        ('Spanish', 'Hola'),
        ('French', 'Bonjour')
        ON CONFLICT DO NOTHING
    `)
	if err != nil {
		log.Fatalf("Failed to seed greetings: %v", err)
	}

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
