package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type SendGridClient struct {
	apiKey     string
	fromEmail  string
	fromName   string
	httpClient *http.Client
}

// NewSendGridClient initializes a SendGrid email client
func NewSendGridClient() *SendGridClient {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	fromEmail := os.Getenv("SENDER_EMAIL")
	fromName := os.Getenv("SENDER_NAME")

	if fromEmail == "" {
		fromEmail = "noreply@paymentapp.com"
	}
	if fromName == "" {
		fromName = "PaymentApp"
	}

	if apiKey == "" {
		log.Println("⚠️  SENDGRID_API_KEY not configured — email sending disabled")
		return nil
	}

	return &SendGridClient{
		apiKey:     apiKey,
		fromEmail:  fromEmail,
		fromName:   fromName,
		httpClient: &http.Client{},
	}
}

// EmailRequest represents a SendGrid email request
type EmailRequest struct {
	Personalizations []Personalization `json:"personalizations"`
	From             EmailAddress      `json:"from"`
	Subject          string            `json:"subject"`
	Content          []Content         `json:"content"`
}

// Personalization represents recipient(s) for an email
type Personalization struct {
	To []EmailAddress `json:"to"`
}

// EmailAddress represents an email with optional name
type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Content represents email content (text/html)
type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// SendPasswordResetEmail sends a password reset email to the user
func (client *SendGridClient) SendPasswordResetEmail(toEmail, userName, resetToken string) error {
	if client == nil {
		log.Printf("⚠️  Email service disabled — skipping password reset email to %s", toEmail)
		return nil
	}

	// Build reset link
	// In production, this should be your frontend URL
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:8081"
	}
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, resetToken)

	// HTML email template
	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2563eb; color: white; padding: 20px; border-radius: 8px; text-align: center; }
        .content { padding: 20px; background-color: #f9fafb; border-radius: 8px; margin-top: 20px; }
        .button { display: inline-block; background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin-top: 20px; }
        .footer { text-align: center; color: #666; font-size: 12px; margin-top: 30px; }
        .warning { color: #dc2626; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        
        <div class="content">
            <p>Hi <strong>%s</strong>,</p>
            
            <p>We received a request to reset your PaymentApp password. If you didn't make this request, you can safely ignore this email.</p>
            
            <p>To reset your password, click the button below:</p>
            
            <a href="%s" class="button">Reset Password</a>
            
            <p><strong>Or copy this link:</strong></p>
            <p style="background-color: #e5e7eb; padding: 10px; border-radius: 4px; word-break: break-all; font-size: 12px;">
                %s
            </p>
            
            <p><span class="warning">⚠️ This link will expire in 15 minutes.</span></p>
            
            <p>If you didn't request a password reset, please ignore this email or contact our support team.</p>
        </div>
        
        <div class="footer">
            <p>&copy; 2026 PaymentApp. All rights reserved.</p>
            <p>This is an automated message — please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>
	`, userName, resetLink, resetLink)

	// Plain text fallback
	textContent := fmt.Sprintf(
		`Password Reset Request

Hi %s,

We received a request to reset your PaymentApp password.

Click the link below to reset your password (valid for 15 minutes):
%s

If you didn't request this, please ignore this email.

PaymentApp
		`, userName, resetLink)

	emailReq := EmailRequest{
		Personalizations: []Personalization{
			{
				To: []EmailAddress{
					{
						Email: toEmail,
						Name:  userName,
					},
				},
			},
		},
		From: EmailAddress{
			Email: client.fromEmail,
			Name:  client.fromName,
		},
		Subject: "Reset Your PaymentApp Password",
		Content: []Content{
			{
				Type:  "text/plain",
				Value: textContent,
			},
			{
				Type:  "text/html",
				Value: htmlContent,
			},
		},
	}

	return client.sendRequest(emailReq)
}

// sendRequest sends the email via SendGrid API
func (client *SendGridClient) sendRequest(emailReq EmailRequest) error {
	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	// SendGrid returns 202 on success
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sendgrid error (status %d): %s", resp.StatusCode, string(body))
	}

	log.Printf("📧 Password reset email sent to %s", emailReq.Personalizations[0].To[0].Email)
	return nil
}
