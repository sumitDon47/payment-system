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
		"host=%s port=%s user=%s password=%s dbname-%s sslmode-disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DBNAME"),
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("DB not reachable:%v", err)
	}

	log.Println("Payment service DB connected")
	runMigrations()
}

func runMigrations() {
	// Share the users table from user-service (same DB, different service)
	// Add transactions table for this service
	query := `
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

	CREATE INDEX IF NOT EXISTS idx_txn_sender   ON transactions(sender_id);
	CREATE INDEX IF NOT EXISTS idx_txn_receiver ON transactions(receiver_id);
	CREATE INDEX IF NOT EXISTS idx_txn_status   ON transactions(status);
	`

	if _, err := DB.Exec(query); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
	log.Println("✅ Payment service migrations complete")
}
