package handler

import (
	"fmt"
	"log"

	"github.com/sumitDon47/payment-system/notification-service/db"
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
	// Fetch real sender and receiver info from DB
	senderName, senderEmail, err := db.GetUserInfo(event.SenderID)
	if err != nil {
		log.Printf("  ⚠️ [EMAIL] Could not fetch sender info for %s: %v", event.SenderID, err)
		senderName = "Valued Customer"
		senderEmail = fmt.Sprintf("sender-%s@example.com", event.SenderID)
	}

	receiverName, _, err := db.GetUserInfo(event.ReceiverID)
	if err != nil {
		log.Printf("  ⚠️ [EMAIL] Could not fetch receiver info for %s: %v", event.ReceiverID, err)
		receiverName = "Unknown Receiver"
	}

	// Generate HTML email body
	htmlBody := email.PaymentCompletedHTML(
		event.Amount,
		event.Currency,
		receiverName,
		event.TransactionID,
		event.SenderBalance,
	)

	subject := fmt.Sprintf("Payment Confirmation: %.2f %s sent", event.Amount, event.Currency)

	// Send via SendGrid
	if err := emailClient.SendEmail(
		senderEmail,
		senderName,
		subject,
		htmlBody,
	); err != nil {
		log.Printf("  ❌ [EMAIL] Failed to send to sender %s (%s): %v", event.SenderID, senderEmail, err)
		return nil
	}

	log.Printf("  ✅ [EMAIL] Sent to sender %s (%s)", event.SenderID, senderEmail)
	return nil
}

func notifyReceiver(event models.PaymentEvent) error {
	// Fetch real receiver and sender info from DB
	receiverName, receiverEmail, err := db.GetUserInfo(event.ReceiverID)
	if err != nil {
		log.Printf("  ⚠️ [EMAIL] Could not fetch receiver info for %s: %v", event.ReceiverID, err)
		receiverName = "Valued Customer"
		receiverEmail = fmt.Sprintf("receiver-%s@example.com", event.ReceiverID)
	}

	senderName, _, err := db.GetUserInfo(event.SenderID)
	if err != nil {
		log.Printf("  ⚠️ [EMAIL] Could not fetch sender info for %s: %v", event.SenderID, err)
		senderName = "Unknown Sender"
	}

	// Generate HTML email body
	htmlBody := email.PaymentReceivedHTML(
		senderName,
		event.Amount,
		event.Currency,
		event.TransactionID,
	)

	subject := fmt.Sprintf("Payment Received: +%.2f %s", event.Amount, event.Currency)

	// Send via SendGrid
	if err := emailClient.SendEmail(
		receiverEmail,
		receiverName,
		subject,
		htmlBody,
	); err != nil {
		log.Printf("  ❌ [EMAIL] Failed to send to receiver %s (%s): %v", event.ReceiverID, receiverEmail, err)
		return nil
	}

	log.Printf("  ✅ [EMAIL] Sent to receiver %s (%s)", event.ReceiverID, receiverEmail)
	return nil
}

func notifyFailure(event models.PaymentEvent) error {
	senderName, senderEmail, err := db.GetUserInfo(event.SenderID)
	if err != nil {
		log.Printf("  ⚠️ [EMAIL] Could not fetch sender info for %s: %v", event.SenderID, err)
		senderName = "Valued Customer"
		senderEmail = fmt.Sprintf("sender-%s@example.com", event.SenderID)
	}

	// Generate HTML email body
	htmlBody := email.PaymentFailedHTML(
		event.Amount,
		event.Currency,
		"Insufficient funds or system error", // In production, capture actual reason
		event.TransactionID,
	)

	subject := fmt.Sprintf("Payment Failed: %.2f %s", event.Amount, event.Currency)

	// Send via SendGrid
	if err := emailClient.SendEmail(
		senderEmail,
		senderName,
		subject,
		htmlBody,
	); err != nil {
		log.Printf("  ❌ [EMAIL] Failed to send failure notification to sender %s (%s): %v", event.SenderID, senderEmail, err)
		// Don't fail the whole operation if email fails
		return nil
	}

	log.Printf("  ✅ [EMAIL] Sent failure notification to sender %s (%s)", event.SenderID, senderEmail)
	return nil
}
