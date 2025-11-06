package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nhulston/logstream/services/query/internal/server"
)

func main() {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		log.Fatal("HTTP_PORT environment variable is required")
	}

	postgresHost := os.Getenv("POSTGRES_HOST")
	if postgresHost == "" {
		log.Fatal("POSTGRES_HOST environment variable is required")
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser == "" {
		log.Fatal("POSTGRES_USER environment variable is required")
	}

	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	if postgresPassword == "" {
		log.Fatal("POSTGRES_PASSWORD environment variable is required")
	}

	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresDB == "" {
		log.Fatal("POSTGRES_DB environment variable is required")
	}

	postgresPort := os.Getenv("POSTGRES_PORT")
	if postgresPort == "" {
		postgresPort = "5432"
	}

	srv, err := server.NewServer(postgresHost, postgresPort, postgresUser, postgresPassword, postgresDB)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: srv.Router(),
	}

	log.Printf("Starting HTTP server on port %s", port)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	srv.Close()
	log.Println("Server stopped")
}
