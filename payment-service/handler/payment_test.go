package handler

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/sumitDon47/payment-system/payment-service/db"
	model "github.com/sumitDon47/payment-system/payment-service/models"
	pb "github.com/sumitDon47/payment-system/payment-service/proto"
)

var (
	dbInitialized = false
	dbMutex       = &sync.Mutex{}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// generateUniqueEmail creates a unique email address to avoid test conflicts
func generateUniqueEmail(baseEmail string) string {
	return fmt.Sprintf("%s+%d@test.com", baseEmail[:len(baseEmail)-8], rand.Intn(100000000))
}

// setupTestDB initializes the database connection once
func setupTestDB() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if !dbInitialized {
		db.Connect()
		dbInitialized = true
	}
	// No cleanup needed - unique emails ensure isolation
	return nil
}

// createTestUser inserts a user with unique email to avoid conflicts
func createTestUser(t *testing.T, name, email string, initialBalance float64) string {
	uniqueEmail := generateUniqueEmail(email)
	var userID string
	err := db.DB.QueryRow(
		`INSERT INTO users (name, email, password, balance) VALUES ($1, $2, 'hashed_password', $3) RETURNING id`,
		name, uniqueEmail, initialBalance,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

// ============================================================================
//  UNIT TESTS - SendPayment
// ============================================================================

func TestSendPayment_Success(t *testing.T) {
	setupTestDB()
	server := &Server{}
	senderID := createTestUser(t, "Alice", "alice@test.com", 1000)
	receiverID := createTestUser(t, "Bob", "bob@test.com", 0)

	req := &pb.SendPaymentRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Amount:     500,
		Currency:   "NPR",
		Note:       "test payment",
	}
	resp, err := server.SendPayment(context.Background(), req)

	if err != nil {
		t.Errorf("SendPayment failed: %v", err)
	}
	if resp == nil {
		t.Fatal("SendPayment returned nil response")
	}
	if resp.Status != model.StatusCompleted {
		t.Errorf("Expected status 'completed', got '%s'", resp.Status)
	}
	if resp.SenderBalance != 500 {
		t.Errorf("Expected sender balance 500, got %.2f", resp.SenderBalance)
	}
	if resp.TransactionID == "" {
		t.Error("TransactionID should not be empty")
	}

	// Verify DB state
	var senderBalance, receiverBalance float64
	db.DB.QueryRow(`SELECT balance FROM users WHERE id = $1`, senderID).Scan(&senderBalance)
	db.DB.QueryRow(`SELECT balance FROM users WHERE id = $1`, receiverID).Scan(&receiverBalance)

	if senderBalance != 500 {
		t.Errorf("Expected sender balance 500 in DB, got %.2f", senderBalance)
	}
	if receiverBalance != 500 {
		t.Errorf("Expected receiver balance 500 in DB, got %.2f", receiverBalance)
	}

	// Verify transaction record
	var txnStatus string
	db.DB.QueryRow(`SELECT status FROM transactions WHERE id = $1`, resp.TransactionID).Scan(&txnStatus)
	if txnStatus != model.StatusCompleted {
		t.Errorf("Expected transaction status 'completed', got '%s'", txnStatus)
	}

	// Verify outbox event
	var eventCount int
	db.DB.QueryRow(`SELECT COUNT(*) FROM outbox_events WHERE payload->>'transaction_id' = $1`, resp.TransactionID).Scan(&eventCount)
	if eventCount != 1 {
		t.Errorf("Expected 1 outbox event, found %d", eventCount)
	}
}

func TestSendPayment_InsufficientFunds(t *testing.T) {
	setupTestDB()
	server := &Server{}
	senderID := createTestUser(t, "Charlie", "charlie@test.com", 100)
	receiverID := createTestUser(t, "Diana", "diana@test.com", 0)

	req := &pb.SendPaymentRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Amount:     500,
		Currency:   "NPR",
	}
	resp, err := server.SendPayment(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for insufficient funds, got nil")
	}
	if resp != nil {
		t.Errorf("Expected nil response on error, got %v", resp)
	}

	// Verify balances unchanged
	var senderBalance float64
	db.DB.QueryRow(`SELECT balance FROM users WHERE id = $1`, senderID).Scan(&senderBalance)
	if senderBalance != 100 {
		t.Errorf("Sender balance should remain 100 after rollback, got %.2f", senderBalance)
	}
}

func TestSendPayment_SelfPayment(t *testing.T) {
	setupTestDB()
	server := &Server{}
	userID := createTestUser(t, "Eve", "eve@test.com", 1000)

	req := &pb.SendPaymentRequest{
		SenderID:   userID,
		ReceiverID: userID,
		Amount:     100,
		Currency:   "NPR",
	}
	resp, err := server.SendPayment(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for self-payment")
	}
	if resp != nil {
		t.Errorf("Expected nil response on error")
	}
}

func TestSendPayment_MissingReceiverID(t *testing.T) {
	setupTestDB()
	server := &Server{}
	senderID := createTestUser(t, "Frank", "frank@test.com", 1000)

	req := &pb.SendPaymentRequest{
		SenderID:   senderID,
		ReceiverID: "",
		Amount:     100,
		Currency:   "NPR",
	}
	resp, err := server.SendPayment(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for missing receiver_id")
	}
	if resp != nil {
		t.Errorf("Expected nil response on error")
	}
}

func TestSendPayment_InvalidAmount_Negative(t *testing.T) {
	setupTestDB()
	server := &Server{}
	senderID := createTestUser(t, "Grace", "grace@test.com", 1000)
	receiverID := createTestUser(t, "Henry", "henry@test.com", 0)

	req := &pb.SendPaymentRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Amount:     -100,
		Currency:   "NPR",
	}
	resp, err := server.SendPayment(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for negative amount")
	}
	if resp != nil {
		t.Errorf("Expected nil response on error")
	}
}

func TestSendPayment_InvalidAmount_Zero(t *testing.T) {
	setupTestDB()
	server := &Server{}
	senderID := createTestUser(t, "Iris", "iris@test.com", 1000)
	receiverID := createTestUser(t, "Jack", "jack@test.com", 0)

	req := &pb.SendPaymentRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Amount:     0,
		Currency:   "NPR",
	}
	resp, err := server.SendPayment(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for zero amount")
	}
	if resp != nil {
		t.Errorf("Expected nil response on error")
	}
}

func TestSendPayment_AmountExceedsLimit(t *testing.T) {
	setupTestDB()
	server := &Server{}
	senderID := createTestUser(t, "Kate", "kate@test.com", 5000000)
	receiverID := createTestUser(t, "Leo", "leo@test.com", 0)

	req := &pb.SendPaymentRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Amount:     1_000_001,
		Currency:   "NPR",
	}
	resp, err := server.SendPayment(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for amount exceeding limit")
	}
	if resp != nil {
		t.Errorf("Expected nil response on error")
	}
}

func TestSendPayment_ReceiverNotFound(t *testing.T) {
	setupTestDB()
	server := &Server{}
	senderID := createTestUser(t, "Mona", "mona@test.com", 1000)

	req := &pb.SendPaymentRequest{
		SenderID:   senderID,
		ReceiverID: "00000000-0000-0000-0000-000000000000",
		Amount:     100,
		Currency:   "NPR",
	}
	resp, err := server.SendPayment(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for non-existent receiver")
	}
	if resp != nil {
		t.Errorf("Expected nil response on error")
	}
}

func TestSendPayment_SenderNotFound(t *testing.T) {
	setupTestDB()
	server := &Server{}
	receiverID := createTestUser(t, "Nancy", "nancy@test.com", 0)

	req := &pb.SendPaymentRequest{
		SenderID:   "00000000-0000-0000-0000-000000000000",
		ReceiverID: receiverID,
		Amount:     100,
		Currency:   "NPR",
	}
	resp, err := server.SendPayment(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for non-existent sender")
	}
	if resp != nil {
		t.Errorf("Expected nil response on error")
	}
}

func TestSendPayment_DefaultCurrency(t *testing.T) {
	setupTestDB()
	server := &Server{}
	senderID := createTestUser(t, "Oscar", "oscar@test.com", 1000)
	receiverID := createTestUser(t, "Pam", "pam@test.com", 0)

	req := &pb.SendPaymentRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Amount:     100,
		Currency:   "",
		Note:       "test default currency",
	}
	resp, err := server.SendPayment(context.Background(), req)

	if err != nil {
		t.Errorf("SendPayment failed: %v", err)
	}
	if resp == nil {
		t.Fatal("SendPayment returned nil response")
	}

	// Verify transaction was created with NPR
	var currency string
	db.DB.QueryRow(`SELECT currency FROM transactions WHERE id = $1`, resp.TransactionID).Scan(&currency)
	if currency != "NPR" {
		t.Errorf("Expected currency NPR, got %s", currency)
	}
}

// ============================================================================
//  CONCURRENCY TEST - Double-Spend Prevention
// ============================================================================

func TestSendPayment_DoubleSpendPrevention(t *testing.T) {
	setupTestDB()
	server := &Server{}
	senderID := createTestUser(t, "Alice2", "alice2@test.com", 100)
	receiver1ID := createTestUser(t, "Bob2", "bob2@test.com", 0)
	receiver2ID := createTestUser(t, "Charlie2", "charlie2@test.com", 0)

	results := make(chan error, 2)

	go func() {
		req1 := &pb.SendPaymentRequest{
			SenderID:   senderID,
			ReceiverID: receiver1ID,
			Amount:     100,
		}
		_, err := server.SendPayment(context.Background(), req1)
		results <- err
	}()

	go func() {
		req2 := &pb.SendPaymentRequest{
			SenderID:   senderID,
			ReceiverID: receiver2ID,
			Amount:     100,
		}
		_, err := server.SendPayment(context.Background(), req2)
		results <- err
	}()

	err1 := <-results
	err2 := <-results

	// One should succeed, one should fail
	if err1 == nil && err2 == nil {
		t.Fatal("Both transactions succeeded; one should have failed")
	}
	if err1 != nil && err2 != nil {
		t.Fatal("Both transactions failed; one should have succeeded")
	}

	// Verify final balance is valid
	var finalBalance float64
	db.DB.QueryRow(`SELECT balance FROM users WHERE id = $1`, senderID).Scan(&finalBalance)
	if finalBalance < 0 {
		t.Errorf("Balance should never go negative; got %.2f", finalBalance)
	}
}

// ============================================================================
//  TESTS - GetTransaction
// ============================================================================

func TestGetTransaction_Success(t *testing.T) {
	setupTestDB()
	server := &Server{}
	senderID := createTestUser(t, "Quinn", "quinn@test.com", 1000)
	receiverID := createTestUser(t, "Rachel", "rachel@test.com", 0)

	sendReq := &pb.SendPaymentRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Amount:     250,
		Currency:   "NPR",
		Note:       "test memo",
	}
	sendResp, _ := server.SendPayment(context.Background(), sendReq)

	getReq := &pb.GetTransactionRequest{
		TransactionID: sendResp.TransactionID,
	}
	getResp, err := server.GetTransaction(context.Background(), getReq)

	if err != nil {
		t.Errorf("GetTransaction failed: %v", err)
	}
	if getResp == nil {
		t.Fatal("GetTransaction returned nil response")
	}
	if getResp.TransactionID != sendResp.TransactionID {
		t.Errorf("Expected txn ID %s, got %s", sendResp.TransactionID, getResp.TransactionID)
	}
	if getResp.Amount != 250 {
		t.Errorf("Expected amount 250, got %.2f", getResp.Amount)
	}
	if getResp.Status != model.StatusCompleted {
		t.Errorf("Expected status 'completed', got '%s'", getResp.Status)
	}
}

func TestGetTransaction_NotFound(t *testing.T) {
	setupTestDB()
	server := &Server{}

	getReq := &pb.GetTransactionRequest{
		TransactionID: "00000000-0000-0000-0000-000000000000",
	}
	resp, err := server.GetTransaction(context.Background(), getReq)

	if err == nil {
		t.Fatal("Expected error for non-existent transaction")
	}
	if resp != nil {
		t.Errorf("Expected nil response on error")
	}
}

// ============================================================================
//  TESTS - GetBalance
// ============================================================================

func TestGetBalance_Success(t *testing.T) {
	setupTestDB()
	server := &Server{}
	userID := createTestUser(t, "Steve", "steve@test.com", 5000)

	req := &pb.GetBalanceRequest{
		UserID: userID,
	}
	resp, err := server.GetBalance(context.Background(), req)

	if err != nil {
		t.Errorf("GetBalance failed: %v", err)
	}
	if resp == nil {
		t.Fatal("GetBalance returned nil response")
	}
	if resp.Balance != 5000 {
		t.Errorf("Expected balance 5000, got %.2f", resp.Balance)
	}
	if resp.Currency != "NPR" {
		t.Errorf("Expected currency NPR, got %s", resp.Currency)
	}
}

func TestGetBalance_UserNotFound(t *testing.T) {
	setupTestDB()
	server := &Server{}

	req := &pb.GetBalanceRequest{
		UserID: "00000000-0000-0000-0000-000000000000",
	}
	resp, err := server.GetBalance(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}
	if resp != nil {
		t.Errorf("Expected nil response on error")
	}
}
