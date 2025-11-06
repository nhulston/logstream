package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

type Server struct {
	db     *sql.DB
	router *mux.Router
}

type LogEntry struct {
	ID         int64     `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Service    string    `json:"service"`
	Level      string    `json:"level"`
	Message    string    `json:"message"`
	Metadata   *string   `json:"metadata,omitempty"`
	TraceID    *string   `json:"trace_id,omitempty"`
	SpanID     *string   `json:"span_id,omitempty"`
	Host       *string   `json:"host,omitempty"`
	Tags       []string  `json:"tags,omitempty"`
	IngestedAt time.Time `json:"ingested_at"`
}

type QueryResponse struct {
	Logs  []LogEntry `json:"logs"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

func NewServer(pgHost, pgPort, pgUser, pgPass, pgDB string) (*Server, error) {
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

	s := &Server{
		db: db,
	}

	s.setupRoutes()

	log.Println("Query server initialized")
	return s, nil
}

func (s *Server) setupRoutes() {
	s.router = mux.NewRouter()
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
	s.router.HandleFunc("/query", s.handleQuery).Methods("GET")
	s.router.HandleFunc("/query/{trace_id}", s.handleQueryByTrace).Methods("GET")
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	level := r.URL.Query().Get("level")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	offset := (page - 1) * limit

	query := "SELECT id, timestamp, service, level, message, metadata, trace_id, span_id, host, tags, ingested_at FROM logs WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if service != "" {
		query += fmt.Sprintf(" AND service = $%d", argCount)
		args = append(args, service)
		argCount++
	}

	if level != "" {
		query += fmt.Sprintf(" AND level = $%d", argCount)
		args = append(args, level)
		argCount++
	}

	if from != "" {
		query += fmt.Sprintf(" AND timestamp >= $%d", argCount)
		args = append(args, from)
		argCount++
	}

	if to != "" {
		query += fmt.Sprintf(" AND timestamp <= $%d", argCount)
		args = append(args, to)
		argCount++
	}

	query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("Query error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var entry LogEntry
		var tags pq.StringArray
		err := rows.Scan(
			&entry.ID,
			&entry.Timestamp,
			&entry.Service,
			&entry.Level,
			&entry.Message,
			&entry.Metadata,
			&entry.TraceID,
			&entry.SpanID,
			&entry.Host,
			&tags,
			&entry.IngestedAt,
		)
		entry.Tags = []string(tags)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		logs = append(logs, entry)
	}

	response := QueryResponse{
		Logs:  logs,
		Total: len(logs),
		Page:  page,
		Limit: limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleQueryByTrace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	traceID := vars["trace_id"]

	rows, err := s.db.Query("SELECT id, timestamp, service, level, message, metadata, trace_id, span_id, host, tags, ingested_at FROM logs WHERE trace_id = $1 ORDER BY timestamp", traceID)
	if err != nil {
		log.Printf("Query error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var entry LogEntry
		var tags pq.StringArray
		err := rows.Scan(
			&entry.ID,
			&entry.Timestamp,
			&entry.Service,
			&entry.Level,
			&entry.Message,
			&entry.Metadata,
			&entry.TraceID,
			&entry.SpanID,
			&entry.Host,
			&tags,
			&entry.IngestedAt,
		)
		entry.Tags = []string(tags)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		logs = append(logs, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (s *Server) Close() {
	s.db.Close()
}
