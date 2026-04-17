package email

import (
	"strings"
	"testing"
)

func TestPaymentCompletedHTML(t *testing.T) {
	html := PaymentCompletedHTML(500.00, "NPR", "Bob", "txn-123", 1500.00)

	// Check for key content
	checks := map[string]string{
		"Subject line":   "Payment Sent Successfully",
		"Amount":         "500.00 NPR",
		"Recipient":      "Bob",
		"New Balance":    "1500.00 NPR",
		"Transaction ID": "txn-123",
		"HTML structure": "<!DOCTYPE html>",
		"Email template": "<div class=\"container\">",
	}

	for name, expected := range checks {
		if !strings.Contains(html, expected) {
			t.Errorf("%s check failed: expected to find %q in HTML", name, expected)
		}
	}

	if !strings.Contains(html, "4CAF50") { // Green color for success
		t.Error("Expected success color (green) in HTML")
	}
}

func TestPaymentReceivedHTML(t *testing.T) {
	html := PaymentReceivedHTML("Alice", 500.00, "NPR", "txn-123")

	checks := map[string]string{
		"Subject line":    "Payment Received",
		"Amount":          "500.00 NPR",
		"Sender name":     "Alice",
		"Transaction ID":  "txn-123",
		"Positive amount": "+500.00 NPR",
		"HTML structure":  "<!DOCTYPE html>",
	}

	for name, expected := range checks {
		if !strings.Contains(html, expected) {
			t.Errorf("%s check failed: expected to find %q in HTML", name, expected)
		}
	}

	if !strings.Contains(html, "2196F3") { // Blue color for received
		t.Error("Expected received color (blue) in HTML")
	}
}

func TestPaymentFailedHTML(t *testing.T) {
	html := PaymentFailedHTML(500.00, "NPR", "Insufficient funds", "txn-123")

	checks := map[string]string{
		"Subject line":   "Payment Failed",
		"Amount":         "500.00 NPR",
		"Reason":         "Insufficient funds",
		"Transaction ID": "txn-123",
		"HTML structure": "<!DOCTYPE html>",
		"Error box":      "reason-box",
	}

	for name, expected := range checks {
		if !strings.Contains(html, expected) {
			t.Errorf("%s check failed: expected to find %q in HTML", name, expected)
		}
	}

	if !strings.Contains(html, "f44336") { // Red color for failure
		t.Error("Expected failure color (red) in HTML")
	}
}

func TestNewSendGridClient(t *testing.T) {
	client := NewSendGridClient("test-key-123")
	if client == nil {
		t.Fatal("NewSendGridClient returned nil")
	}
	if client.apiKey != "test-key-123" {
		t.Errorf("Expected API key 'test-key-123', got %q", client.apiKey)
	}
}

func TestSendEmail_MissingAPIKey(t *testing.T) {
	// Create client with empty API key
	client := &SendGridClient{apiKey: "", client: nil}

	err := client.SendEmail("test@example.com", "Test", "Subject", "<h1>Test</h1>")
	if err == nil {
		t.Fatal("Expected error when SENDGRID_API_KEY is missing, got nil")
	}
	if !strings.Contains(err.Error(), "SENDGRID_API_KEY") {
		t.Errorf("Expected error mentioning SENDGRID_API_KEY, got: %v", err)
	}
}

func TestEmailRequest_Structure(t *testing.T) {
	req := EmailRequest{
		Personalizations: []Personalization{
			{
				To: []Email{
					{
						Email: "user@example.com",
						Name:  "John Doe",
					},
				},
			},
		},
		From: Email{
			Email: "noreply@paymentsystem.com",
			Name:  "Payment System",
		},
		Subject: "Test Email",
		Content: []Content{
			{
				Type:  "text/html",
				Value: "<h1>Hello</h1>",
			},
		},
	}

	// Verify structure
	if len(req.Personalizations) != 1 {
		t.Error("Expected 1 personalization")
	}
	if req.Personalizations[0].To[0].Email != "user@example.com" {
		t.Error("Recipient email incorrect")
	}
	if req.From.Email != "noreply@paymentsystem.com" {
		t.Error("Sender email incorrect")
	}
	if req.Subject != "Test Email" {
		t.Error("Subject incorrect")
	}
}
