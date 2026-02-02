package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dbHost := "localhost"
	dsn := fmt.Sprintf("postgres://phishvault:password@%s:5432/phishvault?sslmode=disable", dbHost)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Printf("Ping failed: %v. Assuming DB might not be reachable from here, but try query anyway if possible.", err)
	}

	query := `
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'scans' AND column_name = 'risk_score';
	`
	var colName, dataType string
	err = db.QueryRow(query).Scan(&colName, &dataType)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	fmt.Printf("Column: %s, Type: %s\n", colName, dataType)
}
