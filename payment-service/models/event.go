package model

// PaymentEvent is written into the outbox table inside the same DB transaction
// as the balance updates, then published asynchronously by an outbox worker.
type PaymentEvent struct {
	EventType     string  `json:"event_type"`
	TransactionID string  `json:"transaction_id"`
	SenderID      string  `json:"sender_id"`
	ReceiverID    string  `json:"receiver_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Note          string  `json:"note"`
	SenderBalance float64 `json:"sender_balance"`
	OccurredAt    string  `json:"occurred_at"`
}

const (
	TopicPaymentCompleted = "payment.completed"
	OutboxStatusPending   = "pending"
)
