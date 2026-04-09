package model

import "time"

type Transaction struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"` // pending | completed | failed
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

// TransactionStatus constants — never use raw strings
const (
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)
