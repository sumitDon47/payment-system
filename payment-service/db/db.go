package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func Connect() {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		envOrDefault("DB_HOST", "localhost"),
		envOrDefault("DB_PORT", "5432"),
		envOrDefault("DB_USER", "postgres"),
		envOrDefault("DB_PASSWORD", "yourpassword"),
		envOrDefault("DB_NAME", "payment_db"),
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Printf("Payment DB not reachable yet, starting gRPC server without migrations: %v", err)
		return
	}

	log.Println("Payment service DB connected")
	runMigrations()
}

func runMigrations() {
	// Share the users table from user-service (same DB, different service)
	// Add transactions table for this service
	query := `
	CREATE EXTENSION IF NOT EXISTS pgcrypto;

	CREATE TABLE IF NOT EXISTS users (
		id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name      VARCHAR(100) NOT NULL,
		email     VARCHAR(150) UNIQUE NOT NULL,
		password  VARCHAR(255) NOT NULL,
		balance   NUMERIC(15, 2) DEFAULT 0.00,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		sender_id   UUID NOT NULL REFERENCES users(id),
		receiver_id UUID NOT NULL REFERENCES users(id),
		amount      NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
		currency    VARCHAR(10) NOT NULL DEFAULT 'NPR',
		status      VARCHAR(20) NOT NULL DEFAULT 'pending',
		note        TEXT DEFAULT '',
		created_at  TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS outbox_events (
		id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		topic        VARCHAR(100) NOT NULL,
		event_key    VARCHAR(100) NOT NULL,
		payload      JSONB NOT NULL,
		status       VARCHAR(20) NOT NULL DEFAULT 'pending',
		retry_count  INT NOT NULL DEFAULT 0,
		last_error   TEXT,
		created_at   TIMESTAMP DEFAULT NOW(),
		published_at TIMESTAMP
	);

	ALTER TABLE transactions ADD COLUMN IF NOT EXISTS currency   VARCHAR(10) NOT NULL DEFAULT 'NPR';
	ALTER TABLE transactions ADD COLUMN IF NOT EXISTS status     VARCHAR(20) NOT NULL DEFAULT 'pending';
	ALTER TABLE transactions ADD COLUMN IF NOT EXISTS note       TEXT DEFAULT '';
	ALTER TABLE transactions ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT NOW();

	ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS topic        VARCHAR(100) NOT NULL DEFAULT 'payment.completed';
	ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS event_key    VARCHAR(100) NOT NULL DEFAULT '';
	ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS payload      JSONB NOT NULL DEFAULT '{}'::jsonb;
	ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS status       VARCHAR(20) NOT NULL DEFAULT 'pending';
	ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS retry_count  INT NOT NULL DEFAULT 0;
	ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS last_error   TEXT;
	ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS created_at   TIMESTAMP DEFAULT NOW();
	ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS published_at TIMESTAMP;

	CREATE INDEX IF NOT EXISTS idx_txn_sender   ON transactions(sender_id);
	CREATE INDEX IF NOT EXISTS idx_txn_receiver ON transactions(receiver_id);
	CREATE INDEX IF NOT EXISTS idx_txn_status   ON transactions(status);
	CREATE INDEX IF NOT EXISTS idx_outbox_status_created ON outbox_events(status, created_at);
	CREATE UNIQUE INDEX IF NOT EXISTS uq_outbox_topic_key ON outbox_events(topic, event_key);
	`

	if _, err := DB.Exec(query); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
	log.Println("✅ Payment service migrations complete")
}
