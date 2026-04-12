package handler

import (
	"fmt"
	"log"

	"github.com/yourname/payment-system/notification-service/models"
)

// HandlePaymentCompleted fires when a payment succeeds.
// In production this would call an email API, SMS gateway, or push service.
// Right now it logs clearly so you can see it working.
func HandlePaymentCompleted(event models.PaymentEvent) error {
	log.Printf("📧 Sending notifications for completed payment %s", event.TransactionID)

	// Notify sender
	if err := notifySender(event); err != nil {
		return fmt.Errorf("failed to notify sender: %w", err)
	}

	// Notify receiver
	if err := notifyReceiver(event); err != nil {
		return fmt.Errorf("failed to notify receiver: %w", err)
	}

	log.Printf("✅ Notifications sent for transaction %s", event.TransactionID)
	return nil
}

// HandlePaymentFailed fires when a payment fails after being attempted.
func HandlePaymentFailed(event models.PaymentEvent) error {
	log.Printf("⚠️  Sending failure notification for transaction %s", event.TransactionID)

	if err := notifyFailure(event); err != nil {
		return fmt.Errorf("failed to send failure notification: %w", err)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  Private notification functions
//  Replace these with real email/SMS/push implementations
// ─────────────────────────────────────────────────────────────────────────────

func notifySender(event models.PaymentEvent) error {
	// TODO: replace with real email provider (SendGrid, AWS SES, etc.)
	log.Printf(
		"  → [EMAIL] To sender %s: You sent %.2f %s | New balance: %.2f %s | Ref: %s",
		event.SenderID,
		event.Amount,
		event.Currency,
		event.SenderBalance,
		event.Currency,
		event.TransactionID,
	)
	return nil
}

func notifyReceiver(event models.PaymentEvent) error {
	// TODO: replace with real SMS gateway (Twilio, Sparrow SMS for Nepal, etc.)
	log.Printf(
		"  → [SMS]   To receiver %s: You received %.2f %s | Ref: %s",
		event.ReceiverID,
		event.Amount,
		event.Currency,
		event.TransactionID,
	)
	return nil
}

func notifyFailure(event models.PaymentEvent) error {
	log.Printf(
		"  → [EMAIL] To sender %s: Payment of %.2f %s failed | Ref: %s",
		event.SenderID,
		event.Amount,
		event.Currency,
		event.TransactionID,
	)
	return nil
}
