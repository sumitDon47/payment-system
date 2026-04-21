package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// SendGridClient handles email delivery via SendGrid API
type SendGridClient struct {
	apiKey string
	client *http.Client
}

// NewSendGridClient initializes a SendGrid email client
func NewSendGridClient(apiKey string) *SendGridClient {
	if apiKey == "" {
		apiKey = os.Getenv("SENDGRID_API_KEY")
	}
	return &SendGridClient{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// EmailRequest represents a SendGrid email request
type EmailRequest struct {
	Personalizations []Personalization `json:"personalizations"`
	From             Email             `json:"from"`
	Subject          string            `json:"subject"`
	Content          []Content         `json:"content"`
}

// Personalization represents recipient(s) for an email
type Personalization struct {
	To []Email `json:"to"`
}

// Email represents an email address with optional name
type Email struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Content represents email content (text/html)
type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// SendEmail sends an email via SendGrid API
// Returns error if API key is missing or SendGrid request fails
func (s *SendGridClient) SendEmail(toEmail, toName, subject, htmlBody string) error {
	if s.apiKey == "" {
		return fmt.Errorf("SENDGRID_API_KEY environment variable not set")
	}

	// Build email request
	req := EmailRequest{
		Personalizations: []Personalization{
			{
				To: []Email{
					{
						Email: toEmail,
						Name:  toName,
					},
				},
			},
		},
		From: Email{
			Email: "sumitsapkota47@gmail.com",
			Name:  "Payment System",
		},
		Subject: subject,
		Content: []Content{
			{
				Type:  "text/html",
				Value: htmlBody,
			},
		},
	}

	// Marshal to JSON
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	// Make HTTP POST request to SendGrid API
	httpReq, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send email request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SendGrid API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// ============================================================================
//  HTML Email Templates
// ============================================================================

// PaymentCompletedHTML generates HTML email for successful payment
func PaymentCompletedHTML(amount float64, currency string, receiverName string, transactionID string, newBalance float64) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; background-color: #f5f5f5; margin: 0; padding: 0; }
    .container { max-width: 600px; margin: 20px auto; background-color: white; border-radius: 8px; padding: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
    .header { background-color: #4CAF50; color: white; padding: 20px; border-radius: 8px 8px 0 0; text-align: center; }
    .header h1 { margin: 0; font-size: 24px; }
    .content { padding: 20px; }
    .amount { font-size: 32px; font-weight: bold; color: #4CAF50; text-align: center; margin: 20px 0; }
    .details { background-color: #f9f9f9; padding: 15px; border-radius: 4px; margin: 15px 0; }
    .detail-row { margin: 10px 0; display: flex; justify-content: space-between; }
    .label { font-weight: bold; color: #333; }
    .value { color: #666; }
    .footer { text-align: center; color: #999; font-size: 12px; margin-top: 20px; border-top: 1px solid #eee; padding-top: 20px; }
    .ref-id { background-color: #f0f0f0; padding: 5px 10px; border-radius: 3px; font-family: monospace; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>✅ Payment Sent Successfully</h1>
    </div>
    <div class="content">
      <p>Hi,</p>
      <p>Your payment to <strong>%s</strong> has been completed successfully.</p>
      
      <div class="amount">%.2f %s</div>
      
      <div class="details">
        <div class="detail-row">
          <span class="label">Recipient:</span>
          <span class="value">%s</span>
        </div>
        <div class="detail-row">
          <span class="label">Amount:</span>
          <span class="value">%.2f %s</span>
        </div>
        <div class="detail-row">
          <span class="label">Your New Balance:</span>
          <span class="value">%.2f %s</span>
        </div>
        <div class="detail-row">
          <span class="label">Transaction ID:</span>
          <span class="value"><span class="ref-id">%s</span></span>
        </div>
      </div>
      
      <p>If you didn't make this payment or have any concerns, please contact our support team immediately.</p>
      
      <p>Best regards,<br>Payment System Team</p>
    </div>
    <div class="footer">
      <p>This is an automated email. Please do not reply directly.</p>
    </div>
  </div>
</body>
</html>
`, receiverName, amount, currency, receiverName, amount, currency, newBalance, currency, transactionID)
}

// PaymentReceivedHTML generates HTML email for payment received
func PaymentReceivedHTML(senderName string, amount float64, currency string, transactionID string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; background-color: #f5f5f5; margin: 0; padding: 0; }
    .container { max-width: 600px; margin: 20px auto; background-color: white; border-radius: 8px; padding: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
    .header { background-color: #2196F3; color: white; padding: 20px; border-radius: 8px 8px 0 0; text-align: center; }
    .header h1 { margin: 0; font-size: 24px; }
    .content { padding: 20px; }
    .amount { font-size: 32px; font-weight: bold; color: #2196F3; text-align: center; margin: 20px 0; }
    .details { background-color: #f9f9f9; padding: 15px; border-radius: 4px; margin: 15px 0; }
    .detail-row { margin: 10px 0; display: flex; justify-content: space-between; }
    .label { font-weight: bold; color: #333; }
    .value { color: #666; }
    .footer { text-align: center; color: #999; font-size: 12px; margin-top: 20px; border-top: 1px solid #eee; padding-top: 20px; }
    .ref-id { background-color: #f0f0f0; padding: 5px 10px; border-radius: 3px; font-family: monospace; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>💰 Payment Received!</h1>
    </div>
    <div class="content">
      <p>Hi,</p>
      <p>You have received a payment from <strong>%s</strong>.</p>
      
      <div class="amount">+%.2f %s</div>
      
      <div class="details">
        <div class="detail-row">
          <span class="label">From:</span>
          <span class="value">%s</span>
        </div>
        <div class="detail-row">
          <span class="label">Amount:</span>
          <span class="value">%.2f %s</span>
        </div>
        <div class="detail-row">
          <span class="label">Transaction ID:</span>
          <span class="value"><span class="ref-id">%s</span></span>
        </div>
      </div>
      
      <p>The funds have been added to your account and are immediately available.</p>
      
      <p>Best regards,<br>Payment System Team</p>
    </div>
    <div class="footer">
      <p>This is an automated email. Please do not reply directly.</p>
    </div>
  </div>
</body>
</html>
`, senderName, amount, currency, senderName, amount, currency, transactionID)
}

// PaymentFailedHTML generates HTML email for failed payment
func PaymentFailedHTML(amount float64, currency string, reason string, transactionID string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; background-color: #f5f5f5; margin: 0; padding: 0; }
    .container { max-width: 600px; margin: 20px auto; background-color: white; border-radius: 8px; padding: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
    .header { background-color: #f44336; color: white; padding: 20px; border-radius: 8px 8px 0 0; text-align: center; }
    .header h1 { margin: 0; font-size: 24px; }
    .content { padding: 20px; }
    .amount { font-size: 32px; font-weight: bold; color: #f44336; text-align: center; margin: 20px 0; }
    .details { background-color: #f9f9f9; padding: 15px; border-radius: 4px; margin: 15px 0; }
    .detail-row { margin: 10px 0; display: flex; justify-content: space-between; }
    .label { font-weight: bold; color: #333; }
    .value { color: #666; }
    .reason-box { background-color: #ffebee; border-left: 4px solid #f44336; padding: 10px 15px; margin: 15px 0; }
    .footer { text-align: center; color: #999; font-size: 12px; margin-top: 20px; border-top: 1px solid #eee; padding-top: 20px; }
    .ref-id { background-color: #f0f0f0; padding: 5px 10px; border-radius: 3px; font-family: monospace; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>❌ Payment Failed</h1>
    </div>
    <div class="content">
      <p>Hi,</p>
      <p>Unfortunately, your payment could not be completed.</p>
      
      <div class="amount">%.2f %s</div>
      
      <div class="details">
        <div class="detail-row">
          <span class="label">Amount:</span>
          <span class="value">%.2f %s</span>
        </div>
        <div class="detail-row">
          <span class="label">Transaction ID:</span>
          <span class="value"><span class="ref-id">%s</span></span>
        </div>
      </div>
      
      <div class="reason-box">
        <strong>Reason:</strong> %s
      </div>
      
      <p>No funds have been deducted from your account. You can try the payment again.</p>
      
      <p>If the problem persists, please contact our support team.</p>
      
      <p>Best regards,<br>Payment System Team</p>
    </div>
    <div class="footer">
      <p>This is an automated email. Please do not reply directly.</p>
    </div>
  </div>
</body>
</html>
`, amount, currency, amount, currency, transactionID, reason)
}
