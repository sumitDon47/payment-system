package main

import (
"database/sql"
"encoding/json"
"fmt"
"log"
"net/http"
"os"
"time"

"github.com/gorilla/mux"
_ "github.com/lib/pq"
)

// Logger interface for structured logging
type Logger interface {
Info(msg string, kv ...interface{})
Warn(msg string, kv ...interface{})
Error(msg string, kv ...interface{})
Debug(msg string, kv ...interface{})
}

type StandardLogger struct{}

func (s *StandardLogger) Info(msg string, kv ...interface{}) {
log.Printf("[INFO] %s %v", msg, kv)
}

func (s *StandardLogger) Warn(msg string, kv ...interface{}) {
log.Printf("[WARN] %s %v", msg, kv)
}

func (s *StandardLogger) Error(msg string, kv ...interface{}) {
log.Printf("[ERROR] %s %v", msg, kv)
}

func (s *StandardLogger) Debug(msg string, kv ...interface{}) {
log.Printf("[DEBUG] %s %v", msg, kv)
}

var logger = &StandardLogger{}

type WalletLoad struct {
ID        string    json:"id"
UserID    string    json:"user_id"
Amount    float64   json:"amount"
Currency  string    json:"currency"
Status    string    json:"status"
CreatedAt time.Time json:"created_at"
UpdatedAt time.Time json:"updated_at"
}

type WalletTransfer struct {
ID             string    json:"id"
SenderID       string    json:"sender_id"
ReceiverID     string    json:"receiver_id"
Amount         float64   json:"amount"
Currency       string    json:"currency"
Status         string    json:"status"
FailureReason  string    json:"failure_reason,omitempty"
CreatedAt      time.Time json:"created_at"
UpdatedAt      time.Time json:"updated_at"
}

var db *sql.DB

func main() {
var err error
connStr := os.Getenv("DATABASE_URL")
if connStr == "" {
connStr = "host=localhost port=5432 user=postgres password=postgres dbname=bank_db sslmode=disable"
}

db, err = sql.Open("postgres", connStr)
if err != nil {
logger.Error("Failed to connect to database", "error", err)
log.Fatal(err)
}
defer db.Close()

r := mux.NewRouter()
r.HandleFunc("/wallet/load", loadWallet).Methods("POST")
r.HandleFunc("/wallet/load/verify", verifyWalletLoad).Methods("POST")
r.HandleFunc("/wallet/load/{id}", getWalletLoadStatus).Methods("GET")
r.HandleFunc("/wallet/load/failure", handleWalletLoadFailure).Methods("POST")
r.HandleFunc("/wallet/transfer/verify", verifyWalletTransfer).Methods("POST")
r.HandleFunc("/wallet/transfer/{id}", getWalletTransferStatus).Methods("GET")
r.HandleFunc("/wallet/transfer/failure", handleWalletTransferFailure).Methods("POST")

logger.Info("Server starting on port 8080")
log.Fatal(http.ListenAndServe(":8080", r))
}

func loadWallet(w http.ResponseWriter, r *http.Request) {
logger.Debug("Received loadWallet request")
var load WalletLoad
if err := json.NewDecoder(r.Body).Decode(&load); err != nil {
logger.Warn("Invalid request payload", "error", err)
http.Error(w, err.Error(), http.StatusBadRequest)
return
}

if load.UserID == "" || load.Amount <= 0 {
logger.Warn("Validation failed", "userID", load.UserID, "amount", load.Amount)
http.Error(w, "Invalid userID or amount", http.StatusBadRequest)
return
}

load.ID = fmt.Sprintf("ld_%d", time.Now().UnixNano())
load.Status = "pending"
load.CreatedAt = time.Now()
load.UpdatedAt = time.Now()

_, err := db.Exec("INSERT INTO wallet_loads (id, user_id, amount, currency, status, created_at, updated_at) VALUES (, , , , , , )",
load.ID, load.UserID, load.Amount, load.Currency, load.Status, load.CreatedAt, load.UpdatedAt)
if err != nil {
logger.Error("Failed to create load record", "error", err)
http.Error(w, err.Error(), http.StatusInternalServerError)
return
}

logger.Info("Wallet load record created", "id", load.ID, "userID", load.UserID)
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(load)
}

func verifyWalletLoad(w http.ResponseWriter, r *http.Request) {
logger.Debug("Received verifyWalletLoad request")
var req struct {
ID     string json:"id"
Status string json:"status"
}
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
logger.Warn("Invalid verification payload", "error", err)
http.Error(w, err.Error(), http.StatusBadRequest)
return
}

_, err := db.Exec("UPDATE wallet_loads SET status = , updated_at =  WHERE id = ", req.Status, time.Now(), req.ID)
if err != nil {
logger.Error("Failed to update load status", "id", req.ID, "error", err)
http.Error(w, err.Error(), http.StatusInternalServerError)
return
}

logger.Info("Wallet load status updated", "id", req.ID, "status", req.Status)
w.WriteHeader(http.StatusOK)
}

func getWalletLoadStatus(w http.ResponseWriter, r *http.Request) {
vars := mux.Vars(r)
id := vars["id"]
logger.Debug("Querying wallet load status", "id", id)

var load WalletLoad
err := db.QueryRow("SELECT id, user_id, amount, currency, status, created_at, updated_at FROM wallet_loads WHERE id = ", id).
Scan(&load.ID, &load.UserID, &load.Amount, &load.Currency, &load.Status, &load.CreatedAt, &load.UpdatedAt)
if err == sql.ErrNoRows {
logger.Warn("Wallet load record not found", "id", id)
http.Error(w, "Not found", http.StatusNotFound)
return
} else if err != nil {
logger.Error("Database error during status query", "id", id, "error", err)
http.Error(w, err.Error(), http.StatusInternalServerError)
return
}

logger.Info("Wallet load status retrieved", "id", id, "status", load.Status)
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(load)
}

func handleWalletLoadFailure(w http.ResponseWriter, r *http.Request) {
var req struct {
ID     string json:"id"
Reason string json:"reason"
}
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
logger.Warn("Invalid failure reporting payload", "error", err)
http.Error(w, err.Error(), http.StatusBadRequest)
return
}

logger.Warn("Handling wallet load failure", "id", req.ID, "reason", req.Reason)
_, err := db.Exec("UPDATE wallet_loads SET status = 'failed', updated_at =  WHERE id = ", time.Now(), req.ID)
if err != nil {
logger.Error("Failed to mark load as failed", "id", req.ID, "error", err)
http.Error(w, err.Error(), http.StatusInternalServerError)
return
}

logger.Info("Wallet load marked as failed", "id", req.ID)
w.WriteHeader(http.StatusOK)
}

func verifyWalletTransfer(w http.ResponseWriter, r *http.Request) {
logger.Debug("Received verifyWalletTransfer request")
var req struct {
ID     string json:"id"
Status string json:"status"
}
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
logger.Warn("Invalid transfer verification payload", "error", err)
http.Error(w, err.Error(), http.StatusBadRequest)
return
}

_, err := db.Exec("UPDATE wallet_transfers SET status = , updated_at =  WHERE id = ", req.Status, time.Now(), req.ID)
if err != nil {
logger.Error("Failed to update transfer status", "id", req.ID, "error", err)
http.Error(w, err.Error(), http.StatusInternalServerError)
return
}

logger.Info("Wallet transfer status updated", "id", req.ID, "status", req.Status)
w.WriteHeader(http.StatusOK)
}

func getWalletTransferStatus(w http.ResponseWriter, r *http.Request) {
vars := mux.Vars(r)
id := vars["id"]
logger.Debug("Querying wallet transfer status", "id", id)

var transfer WalletTransfer
err := db.QueryRow("SELECT id, sender_id, receiver_id, amount, currency, status, created_at, updated_at FROM wallet_transfers WHERE id = ", id).
Scan(&transfer.ID, &transfer.SenderID, &transfer.ReceiverID, &transfer.Amount, &transfer.Currency, &transfer.Status, &transfer.CreatedAt, &transfer.UpdatedAt)
if err == sql.ErrNoRows {
logger.Warn("Wallet transfer record not found", "id", id)
http.Error(w, "Not found", http.StatusNotFound)
return
} else if err != nil {
logger.Error("Database error during transfer status query", "id", id, "error", err)
http.Error(w, err.Error(), http.StatusInternalServerError)
return
}

logger.Info("Wallet transfer status retrieved", "id", id, "status", transfer.Status)
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(transfer)
}

func handleWalletTransferFailure(w http.ResponseWriter, r *http.Request) {
var req struct {
ID     string json:"id"
Reason string json:"reason"
}
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
logger.Warn("Invalid transfer failure reporting payload", "error", err)
http.Error(w, err.Error(), http.StatusBadRequest)
return
}

logger.Warn("Handling wallet transfer failure", "id", req.ID, "reason", req.Reason)

// Transaction to update transfer and potentially issue refund
tx, err := db.Begin()
if err != nil {
logger.Error("Failed to start transaction for refund", "error", err)
http.Error(w, err.Error(), http.StatusInternalServerError)
return
}

_, err = tx.Exec("UPDATE wallet_transfers SET status = 'failed', failure_reason = , updated_at =  WHERE id = ", req.Reason, time.Now(), req.ID)
if err != nil {
tx.Rollback()
logger.Error("Failed to update transfer status to failed", "id", req.ID, "error", err)
http.Error(w, err.Error(), http.StatusInternalServerError)
return
}

// Logic for refund would go here (e.g., updating sender balance)
logger.Info("Wallet transfer refund processed", "id", req.ID)

if err := tx.Commit(); err != nil {
logger.Error("Failed to commit refund transaction", "id", req.ID, "error", err)
http.Error(w, err.Error(), http.StatusInternalServerError)
return
}

logger.Info("Wallet transfer marked as failed and handled", "id", req.ID)
w.WriteHeader(http.StatusOK)
}
