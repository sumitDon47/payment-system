package handler

import (
	"fmt"
	"log"

	"github.com/sumitDon47/payment-system/notification-service/email"
	"github.com/sumitDon47/payment-system/notification-service/models"
)

var emailClient *email.SendGridClient

func init() {
	// Initialize SendGrid client (reads SENDGRID_API_KEY from env)
	emailClient = email.NewSendGridClient("")
}

// HandlePaymentCompleted fires when a payment succeeds.
// Sends confirmation email to sender and receipt email to receiver.
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
//  Notification implementations using SendGrid
// ─────────────────────────────────────────────────────────────────────────────

func notifySender(event models.PaymentEvent) error {
	// Generate HTML email body
	htmlBody := email.PaymentCompletedHTML(
		event.Amount,
		event.Currency,
		event.ReceiverID, // In a real system, would fetch receiver name from DB
		event.TransactionID,
		event.SenderBalance,
	)

	// Note: In production, you'd fetch the actual sender name and email from the user service
	// For now, we use the user ID as a placeholder
	subject := fmt.Sprintf("Payment Confirmation: %.2f %s sent", event.Amount, event.Currency)

	// Send via SendGrid
	// TODO: Get actual sender email from user service
	if err := emailClient.SendEmail(
		fmt.Sprintf("sender-%s@example.com", event.SenderID), // placeholder
		"",
		subject,
		htmlBody,
	); err != nil {
		log.Printf("  ❌ [EMAIL] Failed to send to sender %s: %v", event.SenderID, err)
		// Don't fail the whole operation if email fails
		// In production, this would go to a retry queue
		return nil
	}

	log.Printf("  ✅ [EMAIL] Sent to sender %s", event.SenderID)
	return nil
}

func notifyReceiver(event models.PaymentEvent) error {
	// Generate HTML email body
	htmlBody := email.PaymentReceivedHTML(
		event.SenderID, // In a real system, would fetch sender name from DB
		event.Amount,
		event.Currency,
		event.TransactionID,
	)

	subject := fmt.Sprintf("Payment Received: +%.2f %s", event.Amount, event.Currency)

	// Send via SendGrid
	// TODO: Get actual receiver email from user service
	if err := emailClient.SendEmail(
		fmt.Sprintf("receiver-%s@example.com", event.ReceiverID), // placeholder
		"",
		subject,
		htmlBody,
	); err != nil {
		log.Printf("  ❌ [EMAIL] Failed to send to receiver %s: %v", event.ReceiverID, err)
		// Don't fail the whole operation if email fails
		return nil
	}

	log.Printf("  ✅ [EMAIL] Sent to receiver %s", event.ReceiverID)
	return nil
}

func notifyFailure(event models.PaymentEvent) error {
	// Generate HTML email body
	htmlBody := email.PaymentFailedHTML(
		event.Amount,
		event.Currency,
		"Insufficient funds or system error", // In production, capture actual reason
		event.TransactionID,
	)

	subject := fmt.Sprintf("Payment Failed: %.2f %s", event.Amount, event.Currency)

	// Send via SendGrid
	// TODO: Get actual sender email from user service
	if err := emailClient.SendEmail(
		fmt.Sprintf("sender-%s@example.com", event.SenderID), // placeholder
		"",
		subject,
		htmlBody,
	); err != nil {
		log.Printf("  ❌ [EMAIL] Failed to send failure notification to sender %s: %v", event.SenderID, err)
		// Don't fail the whole operation if email fails
		return nil
	}

	log.Printf("  ✅ [EMAIL] Sent failure notification to sender %s", event.SenderID)
	return nil
}
