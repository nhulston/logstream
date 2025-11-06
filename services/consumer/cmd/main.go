package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nhulston/logstream/services/consumer/internal/consumer"
)

func main() {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		log.Fatal("KAFKA_BROKERS environment variable is required")
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		log.Fatal("KAFKA_GROUP_ID environment variable is required")
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

	c, err := consumer.NewConsumer(kafkaBrokers, groupID, postgresHost, postgresPort, postgresUser, postgresPassword, postgresDB)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	log.Println("Starting consumer service...")
	go c.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down consumer...")
	c.Close()
	log.Println("Consumer stopped")
}
