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

	CREATE TABLE IF NOT EXISTS password_reset_tokens (
		id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token         VARCHAR(255) NOT NULL UNIQUE,
		expires_at    TIMESTAMP NOT NULL,
		created_at    TIMESTAMP DEFAULT NOW(),
		CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_password_reset_token ON password_reset_tokens(token);
	CREATE INDEX IF NOT EXISTS idx_password_reset_expires ON password_reset_tokens(expires_at);

	CREATE TABLE IF NOT EXISTS otp_codes (
		id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email         VARCHAR(150) NOT NULL,
		code          VARCHAR(6) NOT NULL,
		name          VARCHAR(100) NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		expires_at    TIMESTAMP NOT NULL,
		attempts      INTEGER DEFAULT 0,
		verified      BOOLEAN DEFAULT FALSE,
		created_at    TIMESTAMP DEFAULT NOW(),
		UNIQUE(email, code)
	);

	CREATE INDEX IF NOT EXISTS idx_otp_email ON otp_codes(email);
	CREATE INDEX IF NOT EXISTS idx_otp_expires ON otp_codes(expires_at);
	`
	_, err := DB.Exec(query)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations ran successfully")

}
