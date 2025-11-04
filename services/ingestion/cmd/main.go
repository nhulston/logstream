package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/nhulston/logstream/proto/gen"
	"github.com/nhulston/logstream/services/ingestion/internal/server"
	"google.golang.org/grpc"
)

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		log.Fatal("GRPC_PORT environment variable is required")
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		log.Fatal("KAFKA_BROKERS environment variable is required")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	ingestionServer, err := server.NewIngestionServer(kafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create ingestion server: %v", err)
	}

	pb.RegisterLogIngestionServer(grpcServer, ingestionServer)

	log.Printf("Starting gRPC server on port %s", port)
	log.Printf("Kafka brokers: %s", kafkaBrokers)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		ingestionServer.Close()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Server stopped gracefully")
	case <-ctx.Done():
		log.Println("Shutdown timeout, forcing stop")
		grpcServer.Stop()
	}
}
