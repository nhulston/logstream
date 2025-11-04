package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	pb "github.com/nhulston/logstream/proto/gen"
	"google.golang.org/protobuf/proto"
)

type IngestionServer struct {
	pb.UnimplementedLogIngestionServer
	producer *kafka.Producer
	batcher  *Batcher
}

func NewIngestionServer(kafkaBrokers string) (*IngestionServer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":                     kafkaBrokers,
		"acks":                                  "1",
		"retries":                               3,
		"max.in.flight.requests.per.connection": 5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	server := &IngestionServer{
		producer: producer,
	}

	server.batcher = NewBatcher(100, 5*time.Second, server.sendToKafka)

	go server.batcher.Start()

	log.Println("Ingestion server initialized")
	return server, nil
}

func (s *IngestionServer) Ingest(ctx context.Context, req *pb.IngestRequest) (*pb.IngestResponse, error) {
	// TODO: Add rate limiting (token bucket algorithm) per service

	accepted := 0
	rejected := 0

	for _, logEntry := range req.Logs {
		if err := s.validateLog(logEntry); err != nil {
			log.Printf("Invalid log entry: %v", err)
			rejected++
			continue
		}

		s.batcher.Add(logEntry)
		accepted++
	}

	return &pb.IngestResponse{
		Accepted: int32(accepted),
		Rejected: int32(rejected),
	}, nil
}

func (s *IngestionServer) validateLog(entry *pb.LogEntry) error {
	if entry.Service == "" {
		return fmt.Errorf("service is required")
	}
	if entry.Level == "" {
		return fmt.Errorf("level is required")
	}
	if entry.Message == "" {
		return fmt.Errorf("message is required")
	}
	return nil
}

func (s *IngestionServer) sendToKafka(logs []*pb.LogEntry) error {
	topic := "logs.raw"

	for _, logEntry := range logs {
		value := serializeLog(logEntry)

		err := s.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value: value,
			Key:   []byte(logEntry.Service),
		}, nil)

		if err != nil {
			log.Printf("Failed to produce message: %v", err)
			return err
		}
	}

	s.producer.Flush(5000)
	log.Printf("Sent %d logs to Kafka topic %s", len(logs), topic)
	return nil
}

func (s *IngestionServer) Close() {
	log.Println("Closing ingestion server...")
	s.batcher.Stop()
	s.producer.Close()
}

func serializeLog(entry *pb.LogEntry) []byte {
	data, err := proto.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return nil
	}
	return data
}
