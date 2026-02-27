package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/PhishVault/PhishVault-2/core/domain"
	"github.com/PhishVault/PhishVault-2/services/analysis"
	"github.com/PhishVault/PhishVault-2/services/analysis/scanner"
	"github.com/PhishVault/PhishVault-2/services/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	log.Println("Starting Analysis Worker...")

	// 1. Database Connection
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dsn := fmt.Sprintf("postgres://phishvault:password@%s:5432/phishvault?sslmode=disable", dbHost)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// Self-Heal: Ensure risk_score is FLOAT and updated_at exists
	_, err = db.Exec("ALTER TABLE scans ALTER COLUMN risk_score TYPE FLOAT USING risk_score::double precision")
	if err != nil {
		log.Printf("Schema check warning (risk_score): %v", err)
	}

	_, err = db.Exec("ALTER TABLE scans ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP")
	if err != nil {
		log.Printf("Schema check error (updated_at): %v", err)
	} else {
		log.Println("Ensured scans.updated_at exists")
	}

	_, err = db.Exec("ALTER TABLE scans ADD COLUMN IF NOT EXISTS signals JSONB")
	if err != nil {
		log.Printf("Schema check error (signals): %v", err)
	} else {
		log.Println("Ensured scans.signals exists")
	}

	// 1.5 Initialize Storage (MinIO)
	storageManager, err := storage.NewStorageManager(
		"localhost:9000", "minioadmin", "minioadmin", "phishvault-artifacts",
	)
	if err != nil {
		log.Printf("Failed to init storage: %v", err)
	}

	// 2. Initialize Orchestrator and Scanner
	orchestrator := analysis.NewOrchestrator()
	if orchestrator == nil {
		log.Fatalf("Failed to initialize Orchestrator")
	}

	scanEngine, err := scanner.NewScanner()
	if err != nil {
		log.Printf("Failed to initialize Scanner (Playwright): %v", err)
		// We can continue without scanner if needed, or fail.
		// For ETE, let's try to continue but log heavily.
	} else {
		defer scanEngine.Close()
	}

	// 3. Connect to RabbitMQ
	amqpURL := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"scan_tasks", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack (Set to false for reliability)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	log.Println("Worker is running. Waiting for messages.")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a task: %s", d.Body)

			var task domain.SAL
			if err := json.Unmarshal(d.Body, &task); err != nil {
				log.Printf("Error decoding task: %v", err)
				continue
			}

			// 3.5 Run Scanner (Headless Browser)
			if task.ArtifactType == "URL" && scanEngine != nil {
				log.Printf("Scanning URL: %s", task.URL)
				artifacts, err := scanEngine.Scan(context.Background(), task.URL)
				if err != nil {
					log.Printf("Scanner failed: %v", err)
				} else {
					task.Artifacts = artifacts
					log.Printf("Scan successful. Content Size: %d", len(artifacts.RawContent))
				}
			} else if task.ArtifactType == "EMAIL" || task.ArtifactType == "FILE" {
				log.Printf("Processing %s artifact: %s (Skipping Playwright)", task.ArtifactType, task.ArtifactID)
				// Phase 2: Insert logic for deep header/file analysis here
				task.Verdict = "SCANNED"
				task.RiskScore = 0 // Placeholder
			}

			// 4. Process Task (Real ETE Logic)
			result, err := orchestrator.ProcessArtifact(context.Background(), task)
			if err != nil {
				log.Printf("Analysis failed: %v", err)
				// Update DB to FAILED state so UI doesn't hang
				_, dbErr := db.Exec("UPDATE scans SET verdict = $1 WHERE scan_id = $2", "FAILED", task.ScanID)
				if dbErr != nil {
					log.Printf("Failed to update status to FAILED: %v", dbErr)
				}
				continue
			}

			// 4.5 Persist Artifacts & Details
			var screenshotPath string
			if storageManager != nil && result.Artifacts.Screenshot != "" {
				// Save temp file
				data, decodeErr := base64.StdEncoding.DecodeString(result.Artifacts.Screenshot)
				if decodeErr == nil {
					tmpPath := fmt.Sprintf("/tmp/%s.png", result.ScanID)
					if wErr := os.WriteFile(tmpPath, data, 0644); wErr == nil {
						// Upload
						objName := fmt.Sprintf("scans/%s/screenshot.png", result.ScanID)
						if upErr := storageManager.UploadFile(context.Background(), objName, tmpPath, "image/png"); upErr == nil {
							screenshotPath = objName
							log.Printf("Uploaded screenshot to %s", objName)
						} else {
							log.Printf("MinIO upload failed: %v", upErr)
						}
						os.Remove(tmpPath) // Cleanup
					}
				}
			}

			// Insert details (Basic info for now)
			// Note: We might want accurate IP/Status from Scanner later
			_, dErr := db.Exec(`INSERT INTO scan_details (scan_id, final_url, response_status) VALUES ($1, $2, $3) ON CONFLICT (scan_id) DO NOTHING`,
				result.ScanID, result.FinalURL, 200)
			if dErr != nil {
				log.Printf("Failed to insert scan_details: %v", dErr)
			}

			// Update artifacts table
			if screenshotPath != "" {
				_, aErr := db.Exec(`INSERT INTO artifacts (scan_id, artifact_type, path) VALUES ($1, 'screenshot', $2)`, result.ScanID, screenshotPath)
				if aErr != nil {
					log.Printf("Failed to insert artifact record: %v", aErr)
				}
			}

			// 5. Update Database with Result
			// Note: Orchestrator sets Verdict to MALICIOUS/SAFE
			log.Printf("Updating DB for %s: Verdict=%s, RiskScore=%f", result.ScanID, result.Verdict, result.RiskScore)
			
			signalsJSON, _ := json.Marshal(result.Signals)
			
			_, err = db.Exec("UPDATE scans SET verdict = $1, risk_score = $2, signals = $3, updated_at = NOW() WHERE scan_id = $4",
				result.Verdict, result.RiskScore, signalsJSON, result.ScanID)

			if err != nil {
				log.Printf("Failed to update DB: %v", err)
			} else {
				log.Printf("Scan %s completed. Verdict: %s", result.ScanID, result.Verdict)
				// 6. Manual Ack only after DB persistence
				d.Ack(false)
			}
		}
	}()

	<-forever
}
