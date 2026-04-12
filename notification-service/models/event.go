package models

// PaymentEvent is the message published to Kafka after every payment.
// This is the contract between payment-service (producer)
// and notification-service (consumer).
// If you change this struct, both services must be updated together.
type PaymentEvent struct {
	EventType     string  `json:"event_type"`      // "payment.completed" | "payment.failed"
	TransactionID string  `json:"transaction_id"`
	SenderID      string  `json:"sender_id"`
	ReceiverID    string  `json:"receiver_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Note          string  `json:"note"`
	SenderBalance float64 `json:"sender_balance"`  // new balance after transfer
	OccurredAt    string  `json:"occurred_at"`     // RFC3339 timestamp
}

// Kafka topic names — constants so typos are caught at compile time
const (
	TopicPaymentCompleted = "payment.completed"
	TopicPaymentFailed    = "payment.failed"
)
