package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/sumitDon47/payment-system/payment-service/db"
	model "github.com/sumitDon47/payment-system/payment-service/models"
	pb "github.com/sumitDon47/payment-system/payment-service/proto"
)

// Server implements the gRPC PaymentServiceServer interface
type Server struct{}

func (s *Server) SendPayment(ctx context.Context, req *pb.SendPaymentRequest) (*pb.SendPaymentResponse, error) {

	if req.SenderID == "" || req.ReceiverID == "" {
		return nil, fmt.Errorf("sender_id and receiver_id are required")
	}
	if req.SenderID == req.ReceiverID {
		return nil, fmt.Errorf("cannot send payment to yourself")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	if req.Amount > 1_000_000 {
		return nil, fmt.Errorf("amount exceeds single-transaction limit")
	}

	currency := req.Currency
	if currency == "" {
		currency = "NPR"
	}

	var receiverExists bool
	err := db.DB.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`,
		req.ReceiverID,
	).Scan(&receiverExists)
	if err != nil || !receiverExists {
		return nil, fmt.Errorf("receiver not found")
	}

	tx, err := db.DB.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			log.Printf("Transaction rolled back: %v", err)
		}
	}()

	var senderBalance float64
	err = tx.QueryRowContext(ctx,
		`SELECT balance FROM users WHERE id = $1 FOR UPDATE`,
		req.SenderID,
	).Scan(&senderBalance)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("sender not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sender: %w", err)
	}
	if senderBalance < req.Amount {
		return nil, fmt.Errorf("insufficient funds: have %.2f, need %.2f", senderBalance, req.Amount)
	}

	var txnID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO transactions (sender_id, receiver_id, amount, currency, status, note)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		req.SenderID, req.ReceiverID, req.Amount, currency, model.StatusPending, req.Note,
	).Scan(&txnID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE users SET balance = balance - $1, updated_at = NOW() WHERE id = $2`,
		req.Amount, req.SenderID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to debit sender: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE users SET balance = balance + $1, updated_at = NOW() WHERE id = $2`,
		req.Amount, req.ReceiverID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to credit receiver: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE transactions SET status = $1 WHERE id = $2`,
		model.StatusCompleted, txnID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction status: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	var newBalance float64
	db.DB.QueryRowContext(ctx,
		`SELECT balance FROM users WHERE id = $1`, req.SenderID,
	).Scan(&newBalance)

	log.Printf("Payment %s: %s -> %s | %.2f %s", txnID, req.SenderID, req.ReceiverID, req.Amount, currency)

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
