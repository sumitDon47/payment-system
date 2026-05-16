# Bank Integration Quick Start Guide

## For Banks: How to Integrate with Our Payment System

This guide explains how banks can load customer wallets through our Bank API.

---

## Overview

**What:** Banks can credit customer wallets directly
**How:** Simple REST API with secure authentication
**Status:** Pending, Completed, Failed tracking
**Retries:** Safe via idempotent bank_reference

---

## Getting Started

### Step 1: Request API Credentials

Contact: integration@paymentsystem.com

You'll receive:
```
Bank Code:      IME (your unique identifier)
API Key:        IME-1234567890abcdef
Webhook Secret: webhook-secret-12345
API Endpoint:   https://paymentsystem.com/bank-api/v1
```

### Step 2: Environment Setup

```bash
# For testing
BASE_URL=http://localhost:8082

# For production
BASE_URL=https://api.paymentsystem.com
```

### Step 3: Install Libraries

Choose your language:

**JavaScript/Node.js**
```bash
npm install axios crypto
```

**Python**
```bash
pip install requests
```

**Go**
```bash
go get github.com/go-resty/resty/v2
```

---

## API Reference

### 1. Load Wallet

**Endpoint:** `POST /bank-api/v1/wallet/load`

**Headers:**
```
X-Bank-API-Key: IME-1234567890abcdef
X-Bank-Code: IME
Content-Type: application/json
```

**Request:**
```json
{
  "phone_number": "+977-9800000000",
  "amount": 5000.00,
  "bank_reference": "BANK-TXN-20260515-001",
  "bank_code": "IME",
  "description": "Salary Credit"
}
```

**Response (Success 200):**
```json
{
  "status": "success",
  "transaction_id": "wallet-load-uuid",
  "phone_number": "+977-9800000000",
  "amount": 5000.00,
  "wallet_balance": 15000.00,
  "bank_reference": "BANK-TXN-20260515-001",
  "timestamp": "2026-05-15T10:30:00Z"
}
```

**Response (Duplicate 200):**
```json
{
  "status": "already_processed",
  "transaction_id": "wallet-load-uuid",
  "phone_number": "+977-9800000000",
  "amount": 5000.00,
  "wallet_balance": 15000.00,
  "bank_reference": "BANK-TXN-20260515-001",
  "timestamp": "2026-05-15T10:30:00Z"
}
```

**Response (User Not Found 404):**
```json
{
  "error": "User not found",
  "code": "USER_001",
  "details": "No verified user found with this phone number"
}
```

**Response (Auth Error 401):**
```json
{
  "error": "Invalid API key",
  "code": "AUTH_003"
}
```

---

### 2. Verify Wallet Load (Callback)

After processing the load request, call this to confirm completion.

**Endpoint:** `POST /bank-api/v1/wallet/verify`

**Request:**
```json
{
  "bank_reference": "BANK-TXN-20260515-001",
  "status": "completed",
  "timestamp": "2026-05-15T10:35:00Z",
  "signature": "hmac_sha256_signature"
}
```

**Response (Success 200):**
```json
{
  "status": "verified",
  "bank_reference": "BANK-TXN-20260515-001",
  "message": "Wallet load completed successfully"
}
```

---

### 3. Check Wallet Load Status

Check the status of a previously requested wallet load.

**Endpoint:** `GET /bank-api/v1/wallet/status?transaction_id=wallet-load-uuid`

**Response:**
```json
{
  "transaction_id": "wallet-load-uuid",
  "status": "completed",
  "amount": 5000.00,
  "phone_number": "+977-9800000000",
  "bank_reference": "BANK-TXN-20260515-001",
  "created_at": "2026-05-15T10:30:00Z",
  "completed_at": "2026-05-15T10:35:00Z"
}
```

---

### 4. Failure Notification (Webhook)

If the wallet load fails, call this to notify us.

**Endpoint:** `POST /bank-api/v1/wallet/failure`

**Request:**
```json
{
  "bank_reference": "BANK-TXN-20260515-001",
  "status": "failed",
  "reason": "User not found / Account closed / Insufficient funds",
  "timestamp": "2026-05-15T10:40:00Z"
}
```

**Response:**
```json
{
  "status": "recorded",
  "bank_reference": "BANK-TXN-20260515-001",
  "message": "Failure notification recorded"
}
```

---

## Code Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');
const crypto = require('crypto');

const API_KEY = 'IME-1234567890abcdef';
const BANK_CODE = 'IME';
const BASE_URL = 'http://localhost:8082';

async function loadWallet(phone, amount, bankRef) {
  const payload = {
    phone_number: phone,
    amount: amount,
    bank_reference: bankRef,
    bank_code: BANK_CODE,
    description: 'Salary Credit'
  };

  try {
    const response = await axios.post(
      `${BASE_URL}/bank-api/v1/wallet/load`,
      payload,
      {
        headers: {
          'X-Bank-API-Key': API_KEY,
          'X-Bank-Code': BANK_CODE,
          'Content-Type': 'application/json'
        }
      }
    );

    console.log('Wallet load initiated:', response.data);
    return response.data.transaction_id;

  } catch (error) {
    console.error('Error:', error.response.data);
    throw error;
  }
}

async function verifyWalletLoad(bankRef) {
  const payload = {
    bank_reference: bankRef,
    status: 'completed',
    timestamp: new Date().toISOString(),
    signature: 'hmac_signature_here'
  };

  try {
    const response = await axios.post(
      `${BASE_URL}/bank-api/v1/wallet/verify`,
      payload,
      {
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );

    console.log('Verification response:', response.data);
    return response.data.status;

  } catch (error) {
    console.error('Error:', error.response.data);
    throw error;
  }
}

// Usage
(async () => {
  try {
    const txId = await loadWallet('+977-9800000000', 5000, 'BANK-TXN-001');
    console.log('Transaction ID:', txId);

    // Wait a bit, then verify
    setTimeout(async () => {
      await verifyWalletLoad('BANK-TXN-001');
    }, 2000);

  } catch (error) {
    console.error('Failed:', error);
  }
})();
```

### Python

```python
import requests
import json
from datetime import datetime

API_KEY = 'IME-1234567890abcdef'
BANK_CODE = 'IME'
BASE_URL = 'http://localhost:8082'

def load_wallet(phone, amount, bank_ref):
    """Load money into customer wallet"""
    
    payload = {
        'phone_number': phone,
        'amount': amount,
        'bank_reference': bank_ref,
        'bank_code': BANK_CODE,
        'description': 'Salary Credit'
    }

    headers = {
        'X-Bank-API-Key': API_KEY,
        'X-Bank-Code': BANK_CODE,
        'Content-Type': 'application/json'
    }

    response = requests.post(
        f'{BASE_URL}/bank-api/v1/wallet/load',
        json=payload,
        headers=headers
    )

    if response.status_code == 200:
        data = response.json()
        print(f"Wallet load initiated: {data}")
        return data['transaction_id']
    else:
        print(f"Error: {response.status_code}")
        print(response.json())
        return None

def verify_wallet_load(bank_ref):
    """Verify wallet load completion"""
    
    payload = {
        'bank_reference': bank_ref,
        'status': 'completed',
        'timestamp': datetime.utcnow().isoformat() + 'Z',
        'signature': 'hmac_signature_here'
    }

    response = requests.post(
        f'{BASE_URL}/bank-api/v1/wallet/verify',
        json=payload,
        headers={'Content-Type': 'application/json'}
    )

    if response.status_code == 200:
        data = response.json()
        print(f"Verification successful: {data}")
        return True
    else:
        print(f"Error: {response.status_code}")
        print(response.json())
        return False

# Usage
if __name__ == '__main__':
    tx_id = load_wallet('+977-9800000000', 5000, 'BANK-TXN-001')
    if tx_id:
        print(f'Transaction ID: {tx_id}')
        
        # Wait for processing, then verify
        import time
        time.sleep(2)
        verify_wallet_load('BANK-TXN-001')
```

### Go

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	APIKey   = "IME-1234567890abcdef"
	BankCode = "IME"
	BaseURL  = "http://localhost:8082"
)

type WalletLoadRequest struct {
	PhoneNumber   string  `json:"phone_number"`
	Amount        float64 `json:"amount"`
	BankReference string  `json:"bank_reference"`
	BankCode      string  `json:"bank_code"`
	Description   string  `json:"description"`
}

type WalletLoadResponse struct {
	Status        string    `json:"status"`
	TransactionID string    `json:"transaction_id"`
	PhoneNumber   string    `json:"phone_number"`
	Amount        float64   `json:"amount"`
	WalletBalance float64   `json:"wallet_balance"`
	BankReference string    `json:"bank_reference"`
	Timestamp     time.Time `json:"timestamp"`
}

func loadWallet(phone string, amount float64, bankRef string) (string, error) {
	payload := WalletLoadRequest{
		PhoneNumber:   phone,
		Amount:        amount,
		BankReference: bankRef,
		BankCode:      BankCode,
		Description:   "Salary Credit",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", BaseURL+"/bank-api/v1/wallet/load", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("X-Bank-API-Key", APIKey)
	req.Header.Set("X-Bank-Code", BankCode)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var data WalletLoadResponse
	if err := json.Unmarshal(respBody, &data); err != nil {
		return "", err
	}

	fmt.Printf("Wallet load initiated: %+v\n", data)
	return data.TransactionID, nil
}

func main() {
	txID, err := loadWallet("+977-9800000000", 5000, "BANK-TXN-001")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Transaction ID: %s\n", txID)
}
```

---

## Error Codes

| Code | HTTP | Meaning | Action |
|------|------|---------|--------|
| AUTH_001 | 401 | Missing auth headers | Add X-Bank-API-Key & X-Bank-Code |
| AUTH_002 | 401 | Invalid bank code | Check your bank code |
| AUTH_003 | 401 | Invalid API key | Verify your API key |
| USER_001 | 404 | User not found | Phone number not registered |
| TXN_001 | 404 | Transaction not found | Check transaction_id |
| REQ_001 | 400 | Invalid request body | Check JSON format |
| REQ_002 | 400 | Missing required fields | Phone, amount, bank_ref required |
| REQ_003 | 400 | Missing parameter | Check query parameters |
| DB_001 | 500 | Database error | Retry after 60 seconds |
| DB_002 | 500 | Create failed | Retry with new bank_reference |

---

## Rate Limits

```
Per Bank:
- 10 requests per hour (per bank code)
- 1000 requests per day

Per Phone Number:
- 100 wallet loads per day per phone
- 5 concurrent pending requests
```

If limit exceeded: HTTP 429 Too Many Requests

---

## Best Practices

### 1. Use Unique Bank References
```javascript
// ✓ Good - unique for each transaction
bank_reference: `BANK-${bankCode}-${Date.now()}-${Math.random()}`

// ✗ Bad - not unique
bank_reference: 'BANK-LOAD-001'
```

### 2. Handle Retries Safely
```javascript
// Bank reference ensures idempotency
// Safe to retry with same bank_reference
// Will return same transaction_id
async function loadWithRetry(phone, amount, bankRef) {
  for (let i = 0; i < 3; i++) {
    try {
      return await loadWallet(phone, amount, bankRef);
    } catch (error) {
      if (i === 2) throw error;
      await sleep(2000); // Wait 2 seconds before retry
    }
  }
}
```

### 3. Verify After Loading
```javascript
// Always verify after load
const txId = await loadWallet('+977-9800000000', 5000, bankRef);
setTimeout(() => {
  verifyWalletLoad(bankRef);  // Confirm completion
}, 1000);
```

### 4. Handle Errors Gracefully
```javascript
try {
  await loadWallet(phone, amount, bankRef);
} catch (error) {
  if (error.code === 'USER_001') {
    // User not found - phone not registered
    notifyCustomer('Please register with payment system first');
  } else if (error.code === 'AUTH_003') {
    // Invalid API key
    logAlert('API key invalid, update required');
  } else {
    // Generic error - retry later
    queue.push({ phone, amount, bankRef });
  }
}
```

---

## Testing

### Sandbox Environment

```
Base URL: https://sandbox.paymentsystem.com
API Key:  SANDBOX-IME-12345
```

Test Credentials:
```
Phone: +977-9800000000
Email: test@example.com
Password: test123
```

### Test Cases

**Test 1: Successful Load**
```
Request: Load 1000 to +977-9800000000
Expected: status = success, transaction_id returned
Action: Verify balance increased by 1000
```

**Test 2: Duplicate Request (Idempotency)**
```
Request 1: Load with bank_ref = BANK-TEST-001
Request 2: Load with same bank_ref = BANK-TEST-001
Expected: Same transaction_id returned both times
```

**Test 3: Invalid Phone**
```
Request: Load to +977-1111111111 (not registered)
Expected: status = 404, code = USER_001
```

---

## Support

### Contact Information
- **Email:** integration@paymentsystem.com
- **Phone:** +977-1-4100000
- **Slack:** #bank-integrations
- **Documentation:** https://docs.paymentsystem.com

### Troubleshooting

**Issue: 401 Unauthorized**
- Check X-Bank-API-Key header
- Check X-Bank-Code header
- Verify API key hasn't expired

**Issue: 404 User not found**
- Verify phone number format (+977-XXXXXXXXXX)
- Confirm user registered in app
- Check phone_verified = true

**Issue: Stuck in Pending**
- Call GET /wallet/status to check status
- Verify callback was sent (POST /wallet/verify)
- Contact support if still pending after 5 minutes

**Issue: Rate Limited (429)**
- Wait 1 hour before retrying
- Contact us to increase limit
- Consider batch processing

---

## Security Guidelines

1. **Keep API Key Secure**
   - Store in environment variables
   - Never commit to repository
   - Rotate yearly

2. **Use HTTPS in Production**
   - All requests must use HTTPS
   - Verify SSL certificates

3. **Validate Responses**
   - Check HTTP status code
   - Verify transaction_id format
   - Validate signature (if provided)

4. **Log Transactions**
   - Log all requests and responses
   - Keep audit trail for 7 years
   - Encrypt sensitive data in logs

---

## FAQ

**Q: How long does wallet load take?**
A: Typically 1-2 seconds. Check status with GET /wallet/status

**Q: Can I retry a failed load?**
A: Yes, use a new bank_reference. Old reference will still be "failed"

**Q: What if user deletes account?**
A: Pending loads will fail. Completed loads cannot be reversed (contact support)

**Q: Is there a test environment?**
A: Yes, use sandbox URL above with different credentials

**Q: How many requests can I make?**
A: 10/hour per bank, 1000/day. Contact us for higher limits

---

**Version:** 1.0
**Last Updated:** May 15, 2026
**Status:** Production Ready
