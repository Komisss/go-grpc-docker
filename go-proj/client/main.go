package main

import (
	"context"
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

	client := pb.NewHelloServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.SayHello(ctx, &pb.HelloRequest{Name: "World"})
	if err != nil {
		log.Fatalf("Failed to greet: %v", err)
	}
	log.Printf("Response: %s", res.Message)
}
