package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/PhishVault/PhishVault-2/core/domain"
	"github.com/PhishVault/PhishVault-2/services/ingestion/engine"
	"github.com/PhishVault/PhishVault-2/services/intel"
	"github.com/PhishVault/PhishVault-2/services/storage"
	_ "github.com/jackc/pgx/v5/stdlib" // Postgres Driver
)

var (
	producer    *Producer
	db          *sql.DB
	neo4jClient *intel.Neo4jClient
	factory     *engine.ArtifactFactory
	storageMgr  *storage.StorageManager
)

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
		log.Printf("Incoming Request: %s %s", r.Method, r.URL.Path)
		for k, v := range r.Header {
			log.Printf("  Header %s: %v", k, v)
		}
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

	// USE ENGINE: Ingest URL
	// The engine handles canonicalization, stripping params, and redirect unwinding
	ctx := r.Context()
	artifact, err := factory.Ingest(ctx, engine.ArtifactTypeURL, []byte(req.URL), "API-USER", nil)
	if err != nil {
		log.Printf("Ingestion failed: %v", err)
		http.Error(w, "Ingestion Failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Process and Publish (Recursive)
	scanID := processAndPublish(w, artifact, "")

	// Response is handled in processAndPublish for the root item if w is passed
	_ = scanID
}

func submitEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var emailBytes []byte
	var err error

	// Try to handle as multipart form first
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		// 10MB limit
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			log.Printf("Multipart Parse Error: %v", err)
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Missing 'file' field in form", http.StatusBadRequest)
			return
		}
		defer file.Close()
		emailBytes, err = io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file part", http.StatusInternalServerError)
			return
		}
	} else {
		// Fallback to raw body (legacy or direct text submission)
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		emailBytes, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}
	}

	if len(emailBytes) == 0 {
		http.Error(w, "Empty email content", http.StatusBadRequest)
		return
	}

	// DEBUG: Log first 100 bytes and total size to check for corruption
	totalSize := len(emailBytes)
	previewLen := totalSize
	if previewLen > 100 {
		previewLen = 100
	}
	log.Printf("Email Ingestion [%s] - Received %d bytes. First %d bytes: %s", contentType, totalSize, previewLen, string(emailBytes[:previewLen]))

	// USE ENGINE: Ingest Email
	ctx := r.Context()
	artifact, err := factory.Ingest(ctx, engine.ArtifactTypeEmail, emailBytes, "API-EMAIL-USER", nil)
	if err != nil {
		log.Printf("Email Ingestion failed: %v", err)
		http.Error(w, "Processing failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Process and Publish (Recursive)
	processAndPublish(w, artifact, "")
}

func submitFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 10MB limit
	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// USE ENGINE: Ingest File
	ctx := r.Context()
	childMeta := map[string]interface{}{
		"filename":     header.Filename,
		"content_type": header.Header.Get("Content-Type"),
	}
	artifact, err := factory.Ingest(ctx, engine.ArtifactTypeFile, fileBytes, "API-FILE-USER", childMeta)
	if err != nil {
		log.Printf("File Ingestion failed: %v", err)
		http.Error(w, "File Processing Failed", http.StatusInternalServerError)
		return
	}

	// Process and Publish
	processAndPublish(w, artifact, "")
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

// processAndPublish handles recursive task creation, persistence, and queueing.
func processAndPublish(w http.ResponseWriter, artifact *engine.IngestedArtifact, parentScanID string) string {
	// Generate Unique ScanID for this Task
	// Mix content ID + time to ensure unique task ID even for same content
	taskID := fmt.Sprintf("%x", sha256.Sum256([]byte(artifact.ID+time.Now().String())))

	// Map to SAL
	task := domain.SAL{
		ScanID:          taskID,
		Timestamp:       time.Now(),
		Verdict:         "PENDING",
		IngestionSource: artifact.IngestionSource,
		ArtifactID:      artifact.ID,
		Metadata:        artifact.Metadata, // Copy all metadata (DKIM, SPF, etc.)
	}

	// Link to Parent
	if parentScanID != "" {
		if task.Metadata == nil {
			task.Metadata = make(map[string]interface{})
		}
		task.Metadata["parent_scan_id"] = parentScanID
	}

	// PROBABLE FIX: Set ArtifactType for Worker Routing
	task.ArtifactType = string(artifact.Type)

	// Type-Specific Mapping
	switch artifact.Type {
	case engine.ArtifactTypeURL:
		task.URL = artifact.Content
		if original, ok := artifact.Metadata["original_url"]; ok {
			task.Request.Headers = map[string]string{"Original-URL": original.(string)}
		}
	case engine.ArtifactTypeEmail:
		task.URL = "email://parsed"
		// Map critical email headers to Top-Level Request Headers for visibility
		headers := make(map[string]string)
		if s, ok := artifact.Metadata["Subject"]; ok {
			headers["Subject"] = s.(string)
		}
		if f, ok := artifact.Metadata["From"]; ok {
			headers["From"] = f.(string)
		}
		task.Request.Headers = headers
	case engine.ArtifactTypeFile:
		task.URL = "file://" + artifact.ID
		if fname, ok := artifact.Metadata["filename"]; ok {
			task.URL = "file://" + fname.(string)
		}
	}

	// 1. Persist
	if artifact.RawData != nil && storageMgr != nil {
		fileName := "original.blob"
		contentType := "application/octet-stream"

		switch artifact.Type {
		case engine.ArtifactTypeEmail:
			fileName = "email.eml"
			contentType = "message/rfc822"
		case engine.ArtifactTypeFile:
			if name, ok := artifact.Metadata["filename"]; ok {
				fileName = name.(string)
			}
			if ct, ok := artifact.Metadata["content_type"]; ok {
				contentType = ct.(string)
			}
		}

		path, err := storageMgr.SaveRaw(context.Background(), task.ScanID, fileName, artifact.RawData, contentType)
		if err == nil {
			task.Artifacts.RawContentPath = path
			log.Printf("Persisted raw artifact to %s", path)
		} else {
			log.Printf("Failed to persist raw artifact: %v", err)
		}
	}

	go saveScanToDB(task)

	// 2. Publish
	publishTask(nil, task, taskID) // Pass nil response writer, we handle response below if needed

	// 3. Recurse for Children
	for _, child := range artifact.Children {
		processAndPublish(nil, child, taskID)
	}

	// 4. Handle HTTP Response (Only for Root Call)
	if w != nil {
		resp := SubmitResponse{
			ScanID: taskID,
			Status: "Queued",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	return taskID
}

func publishTask(w http.ResponseWriter, task domain.SAL, scanID string) {
	body, err := json.Marshal(task)
	if err != nil {
		log.Printf("Failed to marshal task: %v", err)
		if w != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	if producer != nil {
		if err := producer.Publish(body); err != nil {
			log.Printf("Failed to publish to RabbitMQ: %v", err)
		}
	}
}

func scanDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/scans/")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	if db == nil {
		http.Error(w, "DB Disconnected", http.StatusServiceUnavailable)
		return
	}

	// Fetch Query
	var url, verdict string
	var riskScore float64
	var timestamp time.Time

	err := db.QueryRow("SELECT url, verdict, COALESCE(risk_score, 0), timestamp FROM scans WHERE scan_id = $1", id).Scan(&url, &verdict, &riskScore, &timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Not Found", http.StatusNotFound)
		} else {
			log.Printf("DB Error: %v", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		return
	}

	// Fetch Details
	var finalURL string
	var status int
	_ = db.QueryRow("SELECT final_url, response_status FROM scan_details WHERE scan_id = $1", id).Scan(&finalURL, &status)

	// Fetch Screenshot Path
	var screenshotPath string
	_ = db.QueryRow("SELECT path FROM artifacts WHERE scan_id = $1 AND artifact_type = 'screenshot'", id).Scan(&screenshotPath)

	resp := map[string]interface{}{
		"scan_id":         id,
		"url":             url,
		"verdict":         verdict,
		"risk_score":      riskScore,
		"timestamp":       timestamp,
		"final_url":       finalURL,
		"status_code":     status,
		"screenshot_path": screenshotPath, // UI can construct MinIO URL
		"screenshot_url":  "http://localhost:9000/phishvault-artifacts/" + screenshotPath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	var err error

	// Initialize Ingestion Engine
	factory = engine.NewArtifactFactory()
	factory.RegisterProcessor(engine.ArtifactTypeURL, engine.NewCanonicalizerProcessor())
	factory.RegisterProcessor(engine.ArtifactTypeEmail, engine.NewEmailProcessor())
	factory.RegisterProcessor(engine.ArtifactTypeFile, engine.NewAttachmentProcessor()) // For direct file uploads if we add endpoint

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
			// Initialize Users Table
			_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
				id SERIAL PRIMARY KEY,
				username VARCHAR(50) UNIQUE NOT NULL,
				password_hash TEXT NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`)
			if err != nil {
				log.Printf("Failed to create users table: %v", err)
			}
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

	// 5. Storage (MinIO)
	storageMgr, err = storage.NewStorageManager("localhost:9000", "minioadmin", "minioadmin", "phishvault-artifacts")
	if err != nil {
		log.Printf("Failed to connect to Storage: %v", err)
	}

	// 4. HTTP Server Setup with Graceful Shutdown
	mux := http.NewServeMux()

	// Public Endpoints
	mux.HandleFunc("/auth/login", loginHandler)
	mux.HandleFunc("/auth/register", registerHandler)
	mux.HandleFunc("/health", healthHandler) // DOCKER HEALTHCHECK

	// Protected Endpoints
	mux.HandleFunc("/submit", authMiddleware(submitHandler))
	mux.HandleFunc("/submit-email", authMiddleware(submitEmailHandler))
	mux.HandleFunc("/submit-file", authMiddleware(submitFileHandler))
	mux.HandleFunc("/scans", authMiddleware(listScansHandler))         // READ API
	mux.HandleFunc("/stats", authMiddleware(statsHandler))             // STATS API
	mux.HandleFunc("/campaigns", authMiddleware(listCampaignsHandler)) // CAMPAIGNS API
	mux.HandleFunc("/scans/", authMiddleware(scanDetailHandler))       // DETAIL API

	srv := &http.Server{
		Addr:    ":8080",
		Handler: enableCors(mux),
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		log.Println("Ingestion API server listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Channel to listen for interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking select
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case <-shutdown:
		log.Println("Main: Start shutdown")

		// Give outstanding requests a deadline for completion.
		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Asking listener to shutdown and shed load.
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Main: Graceful shutdown did not complete in %v : %v", timeout, err)
			if err := srv.Close(); err != nil {
				log.Fatalf("Main: Could not stop http server: %v", err)
			}
		}
	}

	log.Println("Main: Server stopped")

	// Cleanup Resources
	if db != nil {
		db.Close()
		log.Println("Main: DB Connection Closed")
	}
	if neo4jClient != nil {
		neo4jClient.Close()
		log.Println("Main: Neo4j Connection Closed")
	}
	if producer != nil {
		producer.Close()
		log.Println("Main: RabbitMQ Connection Closed")
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
