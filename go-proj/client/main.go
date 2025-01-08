package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "go-proj/example/servicepb"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("server:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewMovieServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Создаем фильм
	createReq := &pb.CreateMovieRequest{
		Title:       "Унесенные призраками",
		Description: "Тихиро с мамой и папой переезжает в новый дом. Заблудившись по дороге, они оказываются в странном пустынном городе, где их ждет великолепный пир. Родители с жадностью набрасываются на еду и к ужасу девочки превращаются в свиней, став пленниками злой колдуньи Юбабы. Теперь, оказавшись одна среди волшебных существ и загадочных видений, Тихиро должна придумать, как избавить своих родителей от чар коварной старухи.",
		Year:        2001,
	}
	createResp, err := client.CreateMovie(ctx, createReq)
	if err != nil {
		log.Fatalf("Failed to create movie: %v", err)
	}
	fmt.Printf("Created Movie: ID=%d, Title=%s, Description=%s, Year=%d\n",
		createResp.Id, createResp.Title, createResp.Description, createResp.Year)

	// Получаем один фильм
	getReq := &pb.GetMovieRequest{Id: createResp.Id}
	getResp, err := client.GetMovie(ctx, getReq)
	if err != nil {
		log.Fatalf("Failed to get movie: %v", err)
	}
	fmt.Printf("Fetched Movie: ID=%d, Title=%s, Description=%s, Year=%d\n",
		getResp.Id, getResp.Title, getResp.Description, getResp.Year)

	// Получаем список всех фильмов
	listResp, err := client.ListMovies(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("Failed to list movies: %v", err)
	}
	fmt.Println("All Movies:")
	for _, movie := range listResp.Movies {
		fmt.Printf("ID=%d, Title=%s, Description=%s, Year=%d\n",
			movie.Id, movie.Title, movie.Description, movie.Year)
	}
}
