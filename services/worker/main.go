package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/PhishVault/PhishVault-2/core/domain"
	"github.com/PhishVault/PhishVault-2/services/analysis"
	"github.com/PhishVault/PhishVault-2/services/analysis/scanner"
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
		"analysis_tasks", // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
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
			// Process each message in a safe wrapper to prevent worker crash
			func(delivery amqp091.Delivery) {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("CRITICAL: Worker panic recovered for scan: %v", r)
						// Optionally: Nack the message or update DB with "FAILED" status
					}
				}()

				log.Printf("Received a task: %s", delivery.Body)

				var task domain.SAL
				if err := json.Unmarshal(delivery.Body, &task); err != nil {
					log.Printf("Error decoding task: %v", err)
					return // Skip malformed tasks
				}

				// 3.5 Run Scanner (Headless Browser)
				if scanEngine != nil {
					log.Printf("TxID: %s | Scanning URL: %s", task.ScanID, task.URL)
					artifacts, err := scanEngine.Scan(context.Background(), task.URL)
					if err != nil {
						log.Printf("TxID: %s | Scanner failed: %v", task.ScanID, err)
						// We proceed even if scan fails, to let Orchestrator decide (e.g. valid empty result)
						// or maybe Orchestrator analyzes the error.
						// For now, we update artifacts with error info if possible or just log.
					} else {
						task.Artifacts = artifacts
						log.Printf("TxID: %s | Scan successful. Content Size: %d", task.ScanID, len(artifacts.RawContent))
					}
				}

				// 4. Process Task (Real ETE Logic)
				// Orchestrator handles AI, Graph, Verdict
				result, err := orchestrator.ProcessArtifact(context.Background(), task)
				if err != nil {
					log.Printf("TxID: %s | Analysis failed: %v", task.ScanID, err)
					return
				}

				// 5. Update Database with Result
				_, err = db.Exec("UPDATE scans SET verdict = $1, risk_score = $2 WHERE scan_id = $3",
					result.Verdict, result.RiskScore, result.ScanID)

				if err != nil {
					log.Printf("TxID: %s | Failed to update DB: %v", task.ScanID, err)
				} else {
					log.Printf("TxID: %s | Pipeline Completed. Verdict: %s", task.ScanID, result.Verdict)
				}
			}(d)
		}
	}()

	<-forever
}
