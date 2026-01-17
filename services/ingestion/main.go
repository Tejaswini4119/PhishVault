package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"phishvault/core/domain" // Assumes module name 'phishvault'
)

var producer *Producer

type SubmitRequest struct {
	URL string `json:"url"`
}

type SubmitResponse struct {
	ScanID string `json:"scan_id"`
	Status string `json:"status"`
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Simple ID generation for MVP
	scanID := fmt.Sprintf("%x", sha256.Sum256([]byte(req.URL+time.Now().String())))

	// Initialize SAL with Ingestion Context
	task := domain.SAL{
		ScanID:          scanID,
		URL:             req.URL,
		Timestamp:       time.Now(),
		IngestionSource: "API",
	}

	body, err := json.Marshal(task)
	if err != nil {
		log.Printf("Failed to marshal task: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Push key extraction task to RabbitMQ
	if err := producer.Publish(body); err != nil {
		log.Printf("Failed to publish to RabbitMQ: %v", err)
		http.Error(w, "Failed to queue task", http.StatusInternalServerError)
		return
	}

	resp := SubmitResponse{
		ScanID: scanID,
		Status: "Queued",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func main() {
	var err error
	// Connection string for RabbitMQ.
	// In production, use environment variables.
	// For local dev with Docker, potentially "amqp://guest:guest@localhost:5672/"
	amqpURL := "amqp://guest:guest@localhost:5672/"

	// Connect to RabbitMQ
	producer, err = NewProducer(amqpURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer producer.Close()

	http.HandleFunc("/submit", submitHandler)

	log.Println("Ingestion API server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
