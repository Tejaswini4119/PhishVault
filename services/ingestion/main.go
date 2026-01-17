package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PhishVault/PhishVault-2/core/domain"
	"github.com/PhishVault/PhishVault-2/services/ingestion/parser"
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
		IngestionSource: "API-URL",
	}

	publishTask(w, task, scanID)
}

func submitEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size (e.g. 10MB)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	emailData, err := parser.ParseEmail(r.Body)
	if err != nil {
		log.Printf("Failed to parse email: %v", err)
		http.Error(w, "Invalid Email Content", http.StatusBadRequest)
		return
	}

	// Simple ID generation
	scanID := fmt.Sprintf("%x", sha256.Sum256([]byte(emailData.Subject+time.Now().String())))

	// Perform basic header analysis (e.g. SPF)
	spfResult := "UNKNOWN"
	if spf, ok := emailData.Headers["Received-SPF"]; ok {
		spfResult = parser.AnalyzeSPF(spf)
	}

	// Initialize SAL for Email
	// In a real system, we would upload attachments to MinIO here and populate Artifacts
	task := domain.SAL{
		ScanID:          scanID,
		URL:             "email://source", // Placeholder
		Timestamp:       time.Now(),
		IngestionSource: "API-EMAIL",
		Request: domain.RequestDetails{
			Method: "SMTP-PARSE",
			Headers: map[string]string{
				"Subject": emailData.Subject,
				"From":    emailData.From,
				"SPF":     spfResult,
			},
		},
	}

	publishTask(w, task, scanID)
}

func publishTask(w http.ResponseWriter, task domain.SAL, scanID string) {
	body, err := json.Marshal(task)
	if err != nil {
		log.Printf("Failed to marshal task: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Push task to RabbitMQ
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
	amqpURL := "amqp://guest:guest@localhost:5672/"

	// Connect to RabbitMQ
	producer, err = NewProducer(amqpURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer producer.Close()

	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/submit-email", submitEmailHandler)

	log.Println("Ingestion API server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
