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
	pb.UnimplementedMovieServiceServer
	db *sql.DB
}

// Метод CreateMovie: создание фильма
func (s *server) CreateMovie(ctx context.Context, req *pb.CreateMovieRequest) (*pb.MovieResponse, error) {
	log.Printf("Creating movie: %s (%d)", req.Title, req.Year)

	// Сохраняем фильм в базе данных
	var id int
	err := s.db.QueryRow(
		"INSERT INTO movies (title, description, year) VALUES ($1, $2, $3) RETURNING id",
		req.Title, req.Description, req.Year,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to insert movie: %v", err)
	}

	// Возвращаем информацию о созданном фильме
	return &pb.MovieResponse{
		Id:          int32(id),
		Title:       req.Title,
		Description: req.Description,
		Year:        req.Year,
	}, nil
}

// Метод GetMovie: получение фильма по ID
func (s *server) GetMovie(ctx context.Context, req *pb.GetMovieRequest) (*pb.MovieResponse, error) {
	log.Printf("Fetching movie with ID: %d", req.Id)

	// Достаем фильм из базы данных
	var title, description string
	var year int
	err := s.db.QueryRow(
		"SELECT title, description, year FROM movies WHERE id = $1", req.Id,
	).Scan(&title, &description, &year)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("movie with ID %d not found", req.Id)
	} else if err != nil {
		return nil, fmt.Errorf("failed to fetch movie: %v", err)
	}

	// Возвращаем информацию о фильме
	return &pb.MovieResponse{
		Id:          req.Id,
		Title:       title,
		Description: description,
		Year:        int32(year),
	}, nil
}

// Метод ListMovies: получение списка всех фильмов
func (s *server) ListMovies(ctx context.Context, req *pb.Empty) (*pb.MovieListResponse, error) {
	log.Println("Fetching all movies")

	// Достаем список фильмов из базы данных
	rows, err := s.db.Query("SELECT id, title, description, year FROM movies")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movies: %v", err)
	}
	defer rows.Close()

	// Формируем список фильмов
	var movies []*pb.MovieResponse
	for rows.Next() {
		var id int
		var title, description string
		var year int
		if err := rows.Scan(&id, &title, &description, &year); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		movies = append(movies, &pb.MovieResponse{
			Id:          int32(id),
			Title:       title,
			Description: description,
			Year:        int32(year),
		})
	}

	// Возвращаем список фильмов
	return &pb.MovieListResponse{Movies: movies}, nil
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
        CREATE TABLE IF NOT EXISTS movies (
            id SERIAL PRIMARY KEY,
            title TEXT NOT NULL,
            description TEXT,
            year INT NOT NULL
        )
    `)
	if err != nil {
		log.Fatalf("Failed to create movies table: %v", err)
	}

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMovieServiceServer(grpcServer, &server{db: db})

	log.Println("Server is listening on port 50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
