package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/sumitDon47/payment-system/user-service/models"
)

var DB *sql.DB
var validBankCodes = map[string]string{
	"IME": "ime-api-key-placeholder",
	"NMB": "nmb-api-key-placeholder",
	"SCB": "scb-api-key-placeholder",
}

const BANK_API_PORT = ":8082"

func init() {
	// Initialize database connection
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

	log.Println("Bank API - Database connected")
}

func main() {
	// Setup routes
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/bank-api/v1/wallet/load", validateBankAuth(loadWallet))
	http.HandleFunc("/bank-api/v1/wallet/verify", verifyWalletLoad)
	http.HandleFunc("/bank-api/v1/wallet/status", getWalletLoadStatus)
	http.HandleFunc("/bank-api/v1/wallet/failure", handleWalletLoadFailure)

	log.Printf("Bank API service starting on %s\n", BANK_API_PORT)
	if err := http.ListenAndServe(BANK_API_PORT, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// healthCheck - Health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "bank-api",
	})
}

// validateBankAuth - Middleware to validate bank API key and signature
func validateBankAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-Bank-API-Key")
		bankCode := r.Header.Get("X-Bank-Code")
		signature := r.Header.Get("X-Signature")

		if apiKey == "" || bankCode == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(models.BankAPIError{
				Error: "Missing authentication headers",
				Code:  "AUTH_001",
			})
			return
		}

		// Validate bank code
		expectedKey, ok := validBankCodes[bankCode]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(models.BankAPIError{
				Error: "Invalid bank code",
				Code:  "AUTH_002",
			})
			return
		}

		// Verify API key (in production, this should be more secure)
		if apiKey != expectedKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(models.BankAPIError{
				Error: "Invalid API key",
				Code:  "AUTH_003",
			})
			return
		}

		// TODO: Verify HMAC signature on request body
		_ = signature

		next(w, r)
	}
}

// loadWallet - POST /bank-api/v1/wallet/load
// Bank initiates wallet load request
func loadWallet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.BankAPIError{Error: "Method not allowed", Code: "METHOD_001"})
		return
	}

	var req models.BankWalletLoadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.BankAPIError{Error: "Invalid request body", Code: "REQ_001"})
		return
	}

	// Validate request
	if req.PhoneNumber == "" || req.Amount <= 0 || req.BankReference == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.BankAPIError{
			Error: "Missing required fields",
			Code:  "REQ_002",
		})
		return
	}

	// Look up user by phone number
	var userID string
	err := DB.QueryRow(
		`SELECT id FROM users WHERE phone_number = $1 AND phone_verified = true`,
		req.PhoneNumber,
	).Scan(&userID)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.BankAPIError{
			Error:   "User not found",
			Code:    "USER_001",
			Details: "No verified user found with this phone number",
		})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.BankAPIError{Error: "Database error", Code: "DB_001"})
		return
	}

	// Check for duplicate bank reference (idempotency)
	var existingTxID string
	err = DB.QueryRow(
		`SELECT id FROM bank_wallet_loads WHERE bank_reference = $1`,
		req.BankReference,
	).Scan(&existingTxID)

	if err == nil {
		// Duplicate request - return existing transaction
		var existingLoad models.BankWalletLoad
		DB.QueryRow(
			`SELECT id, user_id, phone_number, amount, bank_reference, bank_code, status, created_at, updated_at, completed_at
			 FROM bank_wallet_loads WHERE id = $1`,
			existingTxID,
		).Scan(&existingLoad.ID, &existingLoad.UserID, &existingLoad.PhoneNumber, &existingLoad.Amount,
			&existingLoad.BankReference, &existingLoad.BankCode, &existingLoad.Status,
			&existingLoad.CreatedAt, &existingLoad.UpdatedAt, &existingLoad.CompletedAt)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.BankWalletLoadResponse{
			Status:        existingLoad.Status,
			TransactionID: existingLoad.ID,
			PhoneNumber:   req.PhoneNumber,
			Amount:        req.Amount,
			BankReference: req.BankReference,
			Timestamp:     existingLoad.CreatedAt,
		})
		return
	}

	// Create bank_wallet_loads record with status = pending
	transactionID := generateUUID()
	bankCode := r.Header.Get("X-Bank-Code")

	_, err = DB.Exec(
		`INSERT INTO bank_wallet_loads (id, user_id, phone_number, amount, bank_reference, bank_code, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'pending', NOW(), NOW())`,
		transactionID, userID, req.PhoneNumber, req.Amount, req.BankReference, bankCode,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.BankAPIError{Error: "Failed to create wallet load", Code: "DB_002"})
		return
	}

	// Get current wallet balance
	var balance float64
	DB.QueryRow(`SELECT balance FROM users WHERE id = $1`, userID).Scan(&balance)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.BankWalletLoadResponse{
		Status:        "pending",
		TransactionID: transactionID,
		PhoneNumber:   req.PhoneNumber,
		Amount:        req.Amount,
		WalletBalance: balance,
		BankReference: req.BankReference,
		Timestamp:     time.Now(),
	})

	log.Printf("Wallet load initiated: TxID=%s, Phone=%s, Amount=%.2f\n", transactionID, req.PhoneNumber, req.Amount)
}

// verifyWalletLoad - POST /bank-api/v1/wallet/verify
// Bank calls this callback to verify wallet load completion
func verifyWalletLoad(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.BankAPIError{Error: "Method not allowed", Code: "METHOD_001"})
		return
	}

	var req models.BankWalletVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.BankAPIError{Error: "Invalid request body", Code: "REQ_001"})
		return
	}

	// Verify signature (TODO: implement HMAC verification)
	_ = req.Signature

	// Update bank_wallet_loads status
	if req.Status == "completed" {
		// Get the transaction and user details
		var userID string
		var amount float64
		err := DB.QueryRow(
			`SELECT user_id, amount FROM bank_wallet_loads WHERE bank_reference = $1`,
			req.BankReference,
		).Scan(&userID, &amount)

		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.BankAPIError{Error: "Transaction not found", Code: "TXN_001"})
			return
		}

		// Update user balance and mark wallet load as completed
		tx, err := DB.Begin()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.BankAPIError{Error: "Transaction error", Code: "DB_003"})
			return
		}

		// Update user balance
		_, err = tx.Exec(
			`UPDATE users SET balance = balance + $1 WHERE id = $2`,
			amount, userID,
		)
		if err != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.BankAPIError{Error: "Failed to update balance", Code: "DB_004"})
			return
		}

		// Update bank_wallet_loads status
		_, err = tx.Exec(
			`UPDATE bank_wallet_loads SET status = 'completed', updated_at = NOW(), completed_at = NOW()
			 WHERE bank_reference = $1`,
			req.BankReference,
		)
		if err != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.BankAPIError{Error: "Failed to update transaction", Code: "DB_005"})
			return
		}

		tx.Commit()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":         "verified",
			"bank_reference": req.BankReference,
			"message":        "Wallet load completed successfully",
		})

		log.Printf("Wallet load verified and completed: BankRef=%s, Amount=%.2f\n", req.BankReference, amount)
	}
}

// getWalletLoadStatus - GET /bank-api/v1/wallet/status
// Check status of a wallet load
func getWalletLoadStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	transactionID := r.URL.Query().Get("transaction_id")
	if transactionID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.BankAPIError{Error: "transaction_id parameter required", Code: "REQ_003"})
		return
	}

	var load models.BankWalletLoad
	err := DB.QueryRow(
		`SELECT id, user_id, phone_number, amount, bank_reference, bank_code, status, created_at, updated_at, completed_at
		 FROM bank_wallet_loads WHERE id = $1`,
		transactionID,
	).Scan(&load.ID, &load.UserID, &load.PhoneNumber, &load.Amount, &load.BankReference,
		&load.BankCode, &load.Status, &load.CreatedAt, &load.UpdatedAt, &load.CompletedAt)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.BankAPIError{Error: "Transaction not found", Code: "TXN_001"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.BankWalletStatusResponse{
		TransactionID: load.ID,
		Status:        load.Status,
		Amount:        load.Amount,
		PhoneNumber:   load.PhoneNumber,
		BankReference: load.BankReference,
		CreatedAt:     load.CreatedAt,
		CompletedAt:   load.CompletedAt,
	})
}

// handleWalletLoadFailure - POST /bank-api/v1/wallet/failure
// Handle wallet load failure notification from bank
func handleWalletLoadFailure(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.BankAPIError{Error: "Method not allowed", Code: "METHOD_001"})
		return
	}

	var req models.BankWalletFailureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.BankAPIError{Error: "Invalid request body", Code: "REQ_001"})
		return
	}

	// Update transaction status to failed
	_, err := DB.Exec(
		`UPDATE bank_wallet_loads SET status = 'failed', updated_at = NOW()
		 WHERE bank_reference = $1`,
		req.BankReference,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.BankAPIError{Error: "Failed to update transaction", Code: "DB_006"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":         "recorded",
		"bank_reference": req.BankReference,
		"message":        "Failure notification recorded",
	})

	log.Printf("Wallet load failed: BankRef=%s, Reason=%s\n", req.BankReference, req.Reason)
}

// Helper function to generate UUID (simplified)
func generateUUID() string {
	return fmt.Sprintf("%x", hmac.New(sha256.New, []byte(time.Now().String())).Sum(nil))[:16]
}
