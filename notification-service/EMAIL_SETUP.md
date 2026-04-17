# Email Notification Integration Guide

## Overview

The notification service now integrates with **SendGrid** to send real emails when payments are completed or fail.

## Architecture

```
Payment Service (gRPC)
        ↓
    Payment completed/failed
        ↓
  Outbox table (reliable async)
        ↓
Outbox Dispatcher
        ↓
   Kafka topic
        ↓
Notification Service (Consumer)
        ↓
  Email Handler
        ↓
  SendGrid API
        ↓
    User Email
```

## Features

### ✅ Completed Payments
- **Sender receives:** Confirmation email with amount, recipient, and new balance
- **Receiver receives:** Receipt email with amount and sender information
- **Color scheme:** Green (#4CAF50) for success

### ❌ Failed Payments
- **Sender receives:** Failure notification with reason and transaction ID
- **Color scheme:** Red (#f44336) for failure
- **Note:** No funds deducted - safe to retry

### 📧 Email Templates
- Professional HTML formatting
- Responsive design (mobile-friendly)
- Transaction ID for reference
- Clear call-to-action

## Setup Instructions

### 1. Get SendGrid API Key

1. Create a free account at [SendGrid](https://sendgrid.com)
2. Go to Settings → [API Keys](https://app.sendgrid.com/settings/api_keys)
3. Create a new API key with "Mail Send" permissions
4. Copy the key (it starts with `SG.`)

### 2. Configure Environment Variables

```bash
# Copy the example file
cp .env.example .env

# Edit .env and add your SendGrid API key
SENDGRID_API_KEY=SG.your-api-key-here
KAFKA_BROKER=kafka:9092
KAFKA_TOPIC=payment.completed
KAFKA_GROUP_ID=notification-service
```

**IMPORTANT:** Never commit `.env` to version control! It contains secrets.

### 3. Update User Emails in Database

Currently, the notification service uses placeholder emails (`sender-{userID}@example.com`).

To use real emails, you need to:

1. Update the `users` table to include email addresses
2. Fetch user email from the user service before sending notifications

**Example update to `notification.go`:**

```go
// Fetch user email from user service (you'd need a user service gRPC client)
userEmail := getUserEmailByID(event.SenderID)

if err := emailClient.SendEmail(
    userEmail,
    userName,
    subject,
    htmlBody,
); err != nil {
    log.Printf("Failed to send email to %s: %v", userEmail, err)
}
```

### 4. Update Sender Email Address

In `email/sendgrid.go`, line 52, update the sender address:

```go
From: Email{
    Email: "noreply@paymentsystem.com",  // Change this to your domain
    Name:  "Payment System",
},
```

For SendGrid to allow sending from this address:
1. Go to [Sender Authentication](https://app.sendgrid.com/settings/sender_auth)
2. Verify your domain or add the email as a sender

## Testing

### Unit Tests

```bash
cd notification-service
go test ./email -v
```

Expected output:
```
✓ TestPaymentCompletedHTML
✓ TestPaymentReceivedHTML
✓ TestPaymentFailedHTML
✓ TestNewSendGridClient
✓ TestSendEmail_MissingAPIKey
✓ TestEmailRequest_Structure
```

### Integration Test (with Kafka)

1. Ensure Docker Compose is running: `docker-compose up`
2. Send a test payment via gRPC to trigger a Kafka event
3. Check notification service logs:
   ```bash
   docker logs payment-system-notification-service-1
   ```
4. Check SendGrid dashboard for sent emails:
   - [Activity Feed](https://app.sendgrid.com/email_activity)

### Manual Test

Use the payment-service test client:

```bash
cd payment-service
go run cmd/test_client/main.go
```

This will:
1. Create test users
2. Send a payment
3. Trigger the outbox event
4. Send email via SendGrid

## Error Handling

### Missing API Key
If `SENDGRID_API_KEY` is not set:
- Email sending fails gracefully (logged but doesn't crash)
- Payment transaction still succeeds (no user-facing impact)
- In production, you'd retry via a background job

### SendGrid API Errors
Common errors:
- **401 Unauthorized:** API key is invalid or expired
- **400 Bad Request:** Email format is invalid or missing fields
- **429 Too Many Requests:** Rate limit exceeded (wait before retrying)

Logs will show: `SendGrid API error (status XXX): ...`

## Troubleshooting

### Emails not sending

1. **Check SENDGRID_API_KEY is set:**
   ```bash
   echo $SENDGRID_API_KEY
   ```

2. **Check logs:**
   ```bash
   docker logs payment-system-notification-service-1 | grep -i email
   ```

3. **Verify sender email is authenticated:**
   - Go to [Sender Authentication](https://app.sendgrid.com/settings/sender_auth)
   - Ensure your domain or email is verified

4. **Test SendGrid API directly:**
   ```bash
   curl -X POST https://api.sendgrid.com/v3/mail/send \
     -H "Authorization: Bearer $SENDGRID_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{
       "personalizations": [{"to": [{"email": "test@example.com"}]}],
       "from": {"email": "noreply@yourdomain.com"},
       "subject": "Test",
       "content": [{"type": "text/html", "value": "<h1>Test</h1>"}]
     }'
   ```

### Emails in spam folder

1. Add SPF record for your domain (prevents spoofing)
2. Add DKIM record for your domain (verifies authenticity)
3. Use a from address matching your domain

SendGrid provides guides in Settings → [Sender Authentication](https://app.sendgrid.com/settings/sender_auth)

## Production Checklist

- [ ] SENDGRID_API_KEY configured (not hardcoded)
- [ ] Sender email address verified with SendGrid
- [ ] SPF and DKIM records added
- [ ] User emails stored in database (not placeholders)
- [ ] Fetch real user emails from user service
- [ ] Test email delivery end-to-end
- [ ] Monitor SendGrid activity dashboard
- [ ] Set up alerts for delivery failures
- [ ] Add retry queue for failed emails
- [ ] Add unsubscribe link to emails (SendGrid requirement)

## Alternative Email Providers

If you prefer other services:

- **AWS SES:** Lower cost, good for high volume
- **Mailgun:** Developer-friendly API, good docs
- **Postmark:** Transactional email focused
- **Brevo (Sendinblue):** Good for Nepal region

Just replace the `email/sendgrid.go` with the provider's SDK.

## Cost

SendGrid offers:
- **Free tier:** 100 emails/day
- **Paid:** $9.95/month for 100K emails/month

For Nepal-based application with likely <10K emails/month, free tier is sufficient.

## Next Steps

1. ✅ Set up SendGrid account and API key
2. ✅ Add real user emails to your database
3. ✅ Update notification handler to fetch real emails
4. ✅ Test email delivery
5. ✅ Deploy to production
