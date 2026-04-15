package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Database is not reachable: %v", err)
	}

	log.Println("Database connected")
	runMigrations()
}

func runMigrations() {
	query := `
	CREATE EXTENSION IF NOT EXISTS pgcrypto;

	CREATE TABLE IF NOT EXISTS users (
		id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name        VARCHAR(100) NOT NULL,
		email       VARCHAR(150) UNIQUE NOT NULL,
		password    VARCHAR(255) NOT NULL,
		balance     NUMERIC(15, 2) DEFAULT 0.00,
		created_at  TIMESTAMP DEFAULT NOW(),
		updated_at  TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		sender_id     UUID REFERENCES users(id),
		receiver_id   UUID REFERENCES users(id),
		amount        NUMERIC(15, 2) NOT NULL,
		status        VARCHAR(20) DEFAULT 'pending',
		created_at    TIMESTAMP DEFAULT NOW()
	);
	`
	_, err := DB.Exec(query)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations ran successfully")

}
