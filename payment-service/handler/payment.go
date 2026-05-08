package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sumitDon47/payment-system/payment-service/db"
	"github.com/sumitDon47/payment-system/payment-service/metrics"
	model "github.com/sumitDon47/payment-system/payment-service/models"
	pb "github.com/sumitDon47/payment-system/payment-service/proto"
	"github.com/sumitDon47/payment-system/payment-service/utils"
)

// Server implements the gRPC PaymentServiceServer interface
type Server struct{}

func (s *Server) SendPayment(ctx context.Context, req *pb.SendPaymentRequest) (*pb.SendPaymentResponse, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.PaymentDuration.WithLabelValues("send_payment").Observe(duration)
	}()

	// Validation
	if req.SenderID == "" || req.ReceiverID == "" {
		metrics.ErrorCounter.WithLabelValues("missing_fields").Inc()
		utils.Warn("SendPayment validation failed", map[string]interface{}{
			"error": "sender_id and receiver_id are required",
		})
		return nil, fmt.Errorf("sender_id and receiver_id are required")
	}
	if req.SenderID == req.ReceiverID {
		metrics.ErrorCounter.WithLabelValues("self_transfer").Inc()
		utils.Warn("SendPayment validation failed", map[string]interface{}{
			"sender_id": req.SenderID,
			"error":     "cannot send payment to yourself",
		})
		return nil, fmt.Errorf("cannot send payment to yourself")
	}
	if req.Amount <= 0 {
		metrics.ErrorCounter.WithLabelValues("invalid_amount").Inc()
		utils.Warn("SendPayment validation failed", map[string]interface{}{
			"sender_id": req.SenderID,
			"amount":    req.Amount,
			"error":     "amount must be greater than zero",
		})
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	if req.Amount > 1_000_000 {
		metrics.ErrorCounter.WithLabelValues("amount_limit_exceeded").Inc()
		utils.Warn("SendPayment validation failed", map[string]interface{}{
			"sender_id": req.SenderID,
			"amount":    req.Amount,
			"error":     "amount exceeds single-transaction limit",
		})
		return nil, fmt.Errorf("amount exceeds single-transaction limit")
	}

	currency := req.Currency
	if currency == "" {
		currency = "NPR"
	}

	utils.Info("SendPayment initiated", map[string]interface{}{
		"sender_id":   req.SenderID,
		"receiver_id": req.ReceiverID,
		"amount":      req.Amount,
		"currency":    currency,
	})

	// Check receiver exists
	var receiverExists bool
	err := db.DB.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`,
		req.ReceiverID,
	).Scan(&receiverExists)
	if err != nil || !receiverExists {
		metrics.ErrorCounter.WithLabelValues("receiver_not_found").Inc()
		utils.Error("Receiver validation failed", fmt.Errorf("receiver not found"), map[string]interface{}{
			"receiver_id": req.ReceiverID,
		})
		return nil, fmt.Errorf("receiver not found")
	}

	// Begin transaction with SERIALIZABLE isolation
	tx, err := db.DB.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		metrics.ErrorCounter.WithLabelValues("transaction_begin_failed").Inc()
		utils.Error("Failed to begin transaction", err, map[string]interface{}{
			"sender_id": req.SenderID,
		})
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			utils.Info("Transaction rolled back", map[string]interface{}{
				"sender_id":   req.SenderID,
				"receiver_id": req.ReceiverID,
				"error":       err.Error(),
			})
		}
	}()

	// Get sender balance with FOR UPDATE lock
	var senderBalance float64
	err = tx.QueryRowContext(ctx,
		`SELECT balance FROM users WHERE id = $1 FOR UPDATE`,
		req.SenderID,
	).Scan(&senderBalance)
	if err == sql.ErrNoRows {
		metrics.ErrorCounter.WithLabelValues("sender_not_found").Inc()
		utils.Error("Sender not found", err, map[string]interface{}{
			"sender_id": req.SenderID,
		})
		return nil, fmt.Errorf("sender not found")
	}
	if err != nil {
		metrics.ErrorCounter.WithLabelValues("balance_fetch_failed").Inc()
		utils.Error("Failed to fetch sender balance", err, map[string]interface{}{
			"sender_id": req.SenderID,
		})
		return nil, fmt.Errorf("failed to fetch sender: %w", err)
	}

	// Check balance
	if senderBalance < req.Amount {
		metrics.ErrorCounter.WithLabelValues("insufficient_funds").Inc()
		utils.Warn("Insufficient funds", map[string]interface{}{
			"sender_id": req.SenderID,
			"balance":   senderBalance,
			"amount":    req.Amount,
		})
		return nil, fmt.Errorf("insufficient funds: have %.2f, need %.2f", senderBalance, req.Amount)
	}
	newSenderBalance := senderBalance - req.Amount

	// Create transaction record
	var txnID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO transactions (sender_id, receiver_id, amount, currency, status, note)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		req.SenderID, req.ReceiverID, req.Amount, currency, model.StatusPending, req.Note,
	).Scan(&txnID)
	if err != nil {
		metrics.ErrorCounter.WithLabelValues("transaction_creation_failed").Inc()
		utils.Error("Failed to create transaction record", err, map[string]interface{}{
			"sender_id":   req.SenderID,
			"receiver_id": req.ReceiverID,
		})
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	// Debit sender
	_, err = tx.ExecContext(ctx,
		`UPDATE users SET balance = balance - $1, updated_at = NOW() WHERE id = $2`,
		req.Amount, req.SenderID,
	)
	if err != nil {
		metrics.ErrorCounter.WithLabelValues("debit_failed").Inc()
		utils.Error("Failed to debit sender", err, map[string]interface{}{
			"sender_id": req.SenderID,
			"amount":    req.Amount,
		})
		return nil, fmt.Errorf("failed to debit sender: %w", err)
	}

	// Credit receiver
	_, err = tx.ExecContext(ctx,
		`UPDATE users SET balance = balance + $1, updated_at = NOW() WHERE id = $2`,
		req.Amount, req.ReceiverID,
	)
	if err != nil {
		metrics.ErrorCounter.WithLabelValues("credit_failed").Inc()
		utils.Error("Failed to credit receiver", err, map[string]interface{}{
			"receiver_id": req.ReceiverID,
			"amount":      req.Amount,
		})
		return nil, fmt.Errorf("failed to credit receiver: %w", err)
	}

	// Update transaction status
	_, err = tx.ExecContext(ctx,
		`UPDATE transactions SET status = $1 WHERE id = $2`,
		model.StatusCompleted, txnID,
	)
	if err != nil {
		metrics.ErrorCounter.WithLabelValues("status_update_failed").Inc()
		utils.Error("Failed to update transaction status", err, map[string]interface{}{
			"transaction_id": txnID,
		})
		return nil, fmt.Errorf("failed to update transaction status: %w", err)
	}

	// Create outbox event
	event := model.PaymentEvent{
		EventType:     model.TopicPaymentCompleted,
		TransactionID: txnID,
		SenderID:      req.SenderID,
		ReceiverID:    req.ReceiverID,
		Amount:        req.Amount,
		Currency:      currency,
		Note:          req.Note,
		SenderBalance: newSenderBalance,
		OccurredAt:    time.Now().UTC().Format(time.RFC3339),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		metrics.ErrorCounter.WithLabelValues("event_encoding_failed").Inc()
		utils.Error("Failed to encode outbox event", err, map[string]interface{}{
			"transaction_id": txnID,
		})
		return nil, fmt.Errorf("failed to encode outbox event: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO outbox_events (topic, event_key, payload, status)
		 VALUES ($1, $2, $3::jsonb, $4)`,
		model.TopicPaymentCompleted,
		txnID,
		string(payload),
		model.OutboxStatusPending,
	)
	if err != nil {
		metrics.ErrorCounter.WithLabelValues("outbox_insert_failed").Inc()
		utils.Error("Failed to enqueue outbox event", err, map[string]interface{}{
			"transaction_id": txnID,
		})
		return nil, fmt.Errorf("failed to enqueue outbox event: %w", err)
	}

	if err = tx.Commit(); err != nil {
		metrics.ErrorCounter.WithLabelValues("transaction_commit_failed").Inc()
		utils.Error("Failed to commit transaction", err, map[string]interface{}{
			"transaction_id": txnID,
		})
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	var newBalance float64
	db.DB.QueryRowContext(ctx,
		`SELECT balance FROM users WHERE id = $1`, req.SenderID,
	).Scan(&newBalance)

	// Record metrics
	metrics.PaymentCounter.WithLabelValues("success", currency).Inc()
	metrics.PaymentAmount.WithLabelValues(currency).Observe(req.Amount)

	// Structured logging for successful payment
	utils.Info("Payment completed successfully", map[string]interface{}{
		"transaction_id": txnID,
		"sender_id":      req.SenderID,
		"receiver_id":    req.ReceiverID,
		"amount":         req.Amount,
		"currency":       currency,
		"sender_balance": newBalance,
	})

	return &pb.SendPaymentResponse{
		TransactionID: txnID,
		Status:        model.StatusCompleted,
		SenderBalance: newBalance,
		Message:       fmt.Sprintf("Successfully sent %.2f %s", req.Amount, currency),
		CreatedAt:     time.Now().Format(time.RFC3339),
	}, nil
}

func (s *Server) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.GetTransactionResponse, error) {
	if req.TransactionID == "" {
		return nil, fmt.Errorf("transaction_id is required")
	}

	var txn model.Transaction
	err := db.DB.QueryRowContext(ctx,
		`SELECT id, sender_id, receiver_id, amount, currency, status, note, created_at
		 FROM transactions WHERE id = $1`,
		req.TransactionID,
	).Scan(&txn.ID, &txn.SenderID, &txn.ReceiverID, &txn.Amount, &txn.Currency, &txn.Status, &txn.Note, &txn.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &pb.GetTransactionResponse{
		TransactionID: txn.ID,
		SenderID:      txn.SenderID,
		ReceiverID:    txn.ReceiverID,
		Amount:        txn.Amount,
		Currency:      txn.Currency,
		Status:        txn.Status,
		Note:          txn.Note,
		CreatedAt:     txn.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Server) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	var balance float64
	err := db.DB.QueryRowContext(ctx,
		`SELECT balance FROM users WHERE id = $1`,
		req.UserID,
	).Scan(&balance)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &pb.GetBalanceResponse{
		UserID:   req.UserID,
		Balance:  balance,
		Currency: "NPR",
	}, nil
}
