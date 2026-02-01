package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PhishVault/PhishVault-2/core/domain"
	"github.com/PhishVault/PhishVault-2/services/ingestion/parser"
	"github.com/PhishVault/PhishVault-2/services/intel"
	_ "github.com/jackc/pgx/v5/stdlib" // Postgres Driver
)

var producer *Producer
var db *sql.DB
var neo4jClient *intel.Neo4jClient

type SubmitRequest struct {
	URL string `json:"url"`
}

type SubmitResponse struct {
	ScanID string `json:"scan_id"`
	Status string `json:"status"`
}

// CORS Middleware
func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
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

	scanID := fmt.Sprintf("%x", sha256.Sum256([]byte(req.URL+time.Now().String())))

	task := domain.SAL{
		ScanID:          scanID,
		URL:             req.URL,
		Timestamp:       time.Now(),
		Verdict:         "PENDING", // Initial state
		IngestionSource: "API-URL",
	}

	// 1. Write to DB (Synchronous persistence for UI visibility)
	go saveScanToDB(task)

	// 2. Publish to Queue (Async Analysis)
	publishTask(w, task, scanID)
}

func submitEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	emailData, err := parser.ParseEmail(r.Body)
	if err != nil {
		log.Printf("Failed to parse email: %v", err)
		http.Error(w, "Invalid Email Content", http.StatusBadRequest)
		return
	}

	scanID := fmt.Sprintf("%x", sha256.Sum256([]byte(emailData.Subject+time.Now().String())))
	spfResult := parser.AnalyzeSPF(emailData.Headers)

	task := domain.SAL{
		ScanID:          scanID,
		URL:             "email://source",
		Timestamp:       time.Now(),
		IngestionSource: "API-EMAIL",
		Verdict:         "PENDING",
		Request: domain.RequestDetails{
			Method: "SMTP-PARSE",
			Headers: map[string]string{
				"Subject": emailData.Subject,
				"From":    emailData.From,
				"SPF":     spfResult,
			},
		},
	}

	go saveScanToDB(task)
	publishTask(w, task, scanID)
}

// saveScanToDB inserts the initial record into Postgres
func saveScanToDB(task domain.SAL) {
	if db == nil {
		return
	}
	// Schema: scan_id, url, verdict, timestamp, metadata (JSONB)
	// Mapping: ScanID -> scan_id, URL -> url, Verdict -> verdict
	// Note: 'verdict' column is used for Verdict in this simplified schema
	query := `INSERT INTO scans (scan_id, url, verdict, timestamp) VALUES ($1, $2, $3, $4) ON CONFLICT (scan_id) DO NOTHING`
	_, err := db.Exec(query, task.ScanID, task.URL, task.Verdict, task.Timestamp)
	if err != nil {
		log.Printf("Error saving scan to DB: %v", err)
	}
}

func listScansHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if db == nil {
		http.Error(w, "Database not connected", http.StatusServiceUnavailable)
		return
	}

	rows, err := db.Query("SELECT scan_id, url, verdict, risk_score, timestamp FROM scans ORDER BY timestamp DESC LIMIT 50")
	if err != nil {
		log.Printf("Query error: %v", err)
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var scans []map[string]interface{}
	for rows.Next() {
		var id, url, verdict string
		var riskScore float64
		var createdAt time.Time

		// Use sql.NullFloat64 for risk_score as it might be null for pending scans
		var riskScoreNull sql.NullFloat64

		if err := rows.Scan(&id, &url, &verdict, &riskScoreNull, &createdAt); err != nil {
			log.Printf("Scan parsing error: %v", err)
			continue
		}

		if riskScoreNull.Valid {
			riskScore = riskScoreNull.Float64
		} else {
			riskScore = 0.0
		}

		scans = append(scans, map[string]interface{}{
			"scan_id":    id,
			"target_url": url,
			"status":     verdict,
			"verdict":    verdict,
			"risk_score": riskScore,
			"timestamp":  createdAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scans)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	// Real Stats from DB and Neo4j
	var scanCount int
	var campaignsCount int
	var nodesCount int

	if db != nil {
		// Count total scans
		db.QueryRow("SELECT COUNT(*) FROM scans").Scan(&scanCount)
	}

	if neo4jClient != nil {
		stats, err := neo4jClient.GetGraphStats()
		if err == nil {
			nodesCount = stats["nodes"]
			campaignsCount = stats["campaigns"]
		} else {
			log.Printf("Failed to get graph stats: %v", err)
		}
	}

	stats := map[string]interface{}{
		"active_scans":    scanCount,
		"campaigns":       campaignsCount,
		"threats_blocked": 1248, // Keeping this mock for now until blocking logic is implemented
		"graph_nodes":     nodesCount,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func listCampaignsHandler(w http.ResponseWriter, r *http.Request) {
	if neo4jClient == nil {
		http.Error(w, "Graph DB not connected", http.StatusServiceUnavailable)
		return
	}

	campaigns, err := neo4jClient.GetCampaigns()
	if err != nil {
		log.Printf("Failed to fetch campaigns: %v", err)
		http.Error(w, "Failed to fetch campaigns", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaigns)
}

func publishTask(w http.ResponseWriter, task domain.SAL, scanID string) {
	body, err := json.Marshal(task)
	if err != nil {
		log.Printf("Failed to marshal task: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if producer != nil {
		if err := producer.Publish(body); err != nil {
			log.Printf("Failed to publish to RabbitMQ: %v", err)
		}
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

	// Database Connection
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	// postgres://user:password@host:port/dbname
	dsn := fmt.Sprintf("postgres://phishvault:password@%s:5432/phishvault?sslmode=disable", dbHost)
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Printf("Failed to open DB: %v", err)
	} else {
		if err := db.Ping(); err != nil {
			log.Printf("Failed to ping DB: %v", err)
		} else {
			log.Println("Connected to PostgreSQL")
		}
	}

	// Neo4j Connection
	neo4jURI := "neo4j://localhost:7687"
	neo4jUser := "neo4j"
	neo4jPass := "password"
	neo4jClient, err = intel.NewNeo4jClient(neo4jURI, neo4jUser, neo4jPass)
	if err != nil {
		log.Printf("Failed to connect to Neo4j: %v", err)
	} else {
		log.Println("Connected to Neo4j")
		defer neo4jClient.Close()
	}

	// RabbitMQ
	amqpURL := "amqp://guest:guest@localhost:5672/"
	producer, err = NewProducer(amqpURL)
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v", err)
	} else {
		defer producer.Close()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/submit", submitHandler)
	mux.HandleFunc("/submit-email", submitEmailHandler)
	mux.HandleFunc("/scans", listScansHandler)         // READ API
	mux.HandleFunc("/stats", statsHandler)             // STATS API
	mux.HandleFunc("/campaigns", listCampaignsHandler) // CAMPAIGNS API

	log.Println("Ingestion API server listening on :8080")
	if err := http.ListenAndServe(":8080", enableCors(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
