# Bank Transfer API Quick Reference

## Bank Transfer OUT (Withdrawal) - Quick Reference

### Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/bank-api/v1/wallet/transfer` | Initiate transfer from wallet to bank |
| POST | `/bank-api/v1/wallet/transfer/verify` | Verify & complete transfer (bank callback) |
| GET | `/bank-api/v1/wallet/transfer/status` | Check transfer status |
| POST | `/bank-api/v1/wallet/transfer/failure` | Record transfer failure (bank callback) |

### 1. Initiate Transfer

**Request:**
```bash
curl -X POST http://localhost:8082/bank-api/v1/wallet/transfer \
  -H "X-Bank-API-Key: ime-api-key-placeholder" \
  -H "X-Bank-Code: IME" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+977-9800000000",
    "amount": 5000.00,
    "bank_account": "1234567890",
    "account_holder": "John Doe",
    "bank_code": "IME",
    "bank_reference": "TXN-20260517-001"
  }'
```

**Success Response (200):**
```json
{
  "status": "pending",
  "transaction_id": "abc123def456",
  "phone_number": "+977-9800000000",
  "amount": 5000.00,
  "bank_account": "1234567890",
  "bank_reference": "TXN-20260517-001",
  "wallet_balance": 45000.00,
  "timestamp": "2026-05-17T10:30:00Z"
}
```

**Error Response (402 - Insufficient Balance):**
```json
{
  "error": "Insufficient balance",
  "code": "BALANCE_001",
  "details": "Available balance: 1000.00, Requested: 5000.00"
}
```

### 2. Verify Transfer (Bank Callback)

**Request:**
```bash
curl -X POST http://localhost:8082/bank-api/v1/wallet/transfer/verify \
  -H "Content-Type: application/json" \
  -d '{
    "bank_reference": "TXN-20260517-001",
    "status": "completed",
    "timestamp": "2026-05-17T10:35:00Z",
    "signature": "hmac_sig_here"
  }'
```

**Success Response (200):**
```json
{
  "status": "verified",
  "bank_reference": "TXN-20260517-001",
  "message": "Wallet transfer completed successfully"
}
```

### 3. Check Status

**Request:**
```bash
curl -X GET "http://localhost:8082/bank-api/v1/wallet/transfer/status?transaction_id=abc123def456"
```

**Response (200):**
```json
{
  "transaction_id": "abc123def456",
  "status": "completed",
  "amount": 5000.00,
  "phone_number": "+977-9800000000",
  "bank_account": "1234567890",
  "bank_reference": "TXN-20260517-001",
  "failure_reason": null,
  "created_at": "2026-05-17T10:30:00Z",
  "completed_at": "2026-05-17T10:35:00Z"
}
```

### 4. Record Failure (Bank Callback)

**Request:**
```bash
curl -X POST http://localhost:8082/bank-api/v1/wallet/transfer/failure \
  -H "Content-Type: application/json" \
  -d '{
    "bank_reference": "TXN-20260517-001",
    "status": "failed",
    "reason": "Account closed",
    "timestamp": "2026-05-17T10:40:00Z"
  }'
```

**Success Response (200):**
```json
{
  "status": "recorded",
  "bank_reference": "TXN-20260517-001",
  "message": "Transfer failure recorded and balance refunded"
}
```

## Comparing: Transfer IN vs Transfer OUT

| Feature | Wallet Load (IN) | Wallet Transfer (OUT) |
|---------|------------------|----------------------|
| **Endpoint** | `/wallet/load` | `/wallet/transfer` |
| **Direction** | Bank → Wallet | Wallet → Bank |
| **Balance Change** | ➕ Added | ➖ Deducted |
| **Failure Action** | Do nothing | Refund amount |
| **Status** | pending, completed, failed | pending, processing, completed, failed, cancelled |
| **Verification** | Callback from bank | Callback from bank |
| **Table** | `bank_wallet_loads` | `bank_wallet_transfers` |

## HTTP Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| 200 | OK | Transfer initiated/verified/status retrieved |
| 400 | Bad Request | Missing required fields |
| 402 | Payment Required | Insufficient balance |
| 404 | Not Found | User or transaction not found |
| 409 | Conflict | Duplicate bank_reference |
| 500 | Internal Server Error | Database error |

## Error Codes

| Code | HTTP | Description |
|------|------|-------------|
| AUTH_001 | 401 | Missing authentication headers |
| AUTH_002 | 401 | Invalid bank code |
| AUTH_003 | 401 | Invalid API key |
| REQ_001 | 400 | Invalid request body |
| REQ_002 | 400 | Missing required fields |
| REQ_003 | 400 | Missing query parameters |
| USER_001 | 404 | User not found |
| USER_002 | 404 | User not verified |
| BALANCE_001 | 402 | Insufficient balance |
| TXN_001 | 404 | Transaction not found |
| DB_001 | 500 | Database connection error |
| DB_002 | 500 | Failed to create transfer |
| DB_003 | 500 | Failed to reserve balance |
| DB_006 | 500 | Failed to update transaction |
| DB_007 | 500 | Failed to refund balance |

## Required Headers

```
X-Bank-API-Key: [Bank API Key]
X-Bank-Code: [Bank Code: IME/NMB/SCB]
Content-Type: application/json
X-Signature: [HMAC Signature - Optional]
```

## Valid Bank Codes

```
IME  - Imam Bank Limited
NMB  - Nepal Merchant Bank
SCB  - Standard Chartered Bank Nepal
```

## Transaction Flow Summary

1. **User initiates transfer** → POST `/wallet/transfer`
   - System checks balance
   - Deducts amount from wallet
   - Returns transaction_id with "pending" status

2. **Bank processes** → [Internal bank processing]

3. **Bank sends callback** → POST `/wallet/transfer/verify` (success) OR `/wallet/transfer/failure` (failure)
   - Success: Transaction marked "completed"
   - Failure: Amount refunded to wallet

4. **User checks status** → GET `/wallet/transfer/status?transaction_id=xxx`
   - Shows current transfer status
