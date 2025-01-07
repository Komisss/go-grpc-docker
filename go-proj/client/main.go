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
	var conn *grpc.ClientConn
	var err error

	// Повторные попытки подключения
	for i := 0; i < 5; i++ {
		fmt.Println("Начало соединения")
		conn, err = grpc.Dial("server:50051", grpc.WithInsecure())
		if err == nil {
			break
		}
		log.Printf("Failed to connect, retrying... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect after retries: %v", err)
	}
	defer conn.Close()
	fmt.Println("Соединение успешно")

	client := pb.NewHelloServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.SayHello(ctx, &pb.HelloRequest{Name: "Andrey"})
	if err != nil {
		log.Fatalf("Failed to greet: %v", err)
	}
	log.Printf("Response: %s", res.Message)
}
