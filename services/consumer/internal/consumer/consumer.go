package consumer

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/lib/pq"
	pb "github.com/nhulston/logstream/proto/gen"
	"google.golang.org/protobuf/proto"
)

type Consumer struct {
	consumer *kafka.Consumer
	db       *sql.DB
	batcher  *Batcher
}

func NewConsumer(brokers, groupID, pgHost, pgPort, pgUser, pgPass, pgDB string) (*Consumer, error) {
	kafkaConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pgHost, pgPort, pgUser, pgPass, pgDB)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	c := &Consumer{
		consumer: kafkaConsumer,
		db:       db,
	}

	c.batcher = NewBatcher(100, 5*time.Second, c.insertBatch)

	log.Println("Consumer initialized")
	return c, nil
}

func (c *Consumer) Start() {
	if err := c.consumer.SubscribeTopics([]string{"logs.raw"}, nil); err != nil {
		log.Fatalf("Failed to subscribe to topics: %v", err)
	}

	log.Println("Consumer started, listening for messages...")
	go c.batcher.Start()

	for {
		msg, err := c.consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		logEntry := &pb.LogEntry{}
		if err := proto.Unmarshal(msg.Value, logEntry); err != nil {
			log.Printf("Failed to unmarshal log entry: %v", err)
			continue
		}

		c.batcher.Add(logEntry)
	}
}

func (c *Consumer) insertBatch(logs []*pb.LogEntry) error {
	if len(logs) == 0 {
		return nil
	}

	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO logs (timestamp, service, level, message, metadata, trace_id, span_id, host, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, entry := range logs {
		timestamp := entry.Timestamp.AsTime()

		var metadata interface{}
		if entry.Metadata != "" {
			metadata = entry.Metadata
		} else {
			metadata = nil
		}

		var traceID interface{}
		if entry.TraceId != "" {
			traceID = entry.TraceId
		} else {
			traceID = nil
		}

		var spanID interface{}
		if entry.SpanId != "" {
			spanID = entry.SpanId
		} else {
			spanID = nil
		}

		_, err = stmt.Exec(
			timestamp,
			entry.Service,
			entry.Level,
			entry.Message,
			metadata,
			traceID,
			spanID,
			entry.Host,
			pq.Array(entry.Tags),
		)
		if err != nil {
			log.Printf("Failed to insert log: %v", err)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Inserted %d logs into Postgres", len(logs))
	return nil
}

func (c *Consumer) Close() {
	c.batcher.Stop()
	c.consumer.Close()
	c.db.Close()
}
