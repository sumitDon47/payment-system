# Bank Transfer OUT (Withdrawal) Feature Guide

## Overview

The Bank Transfer OUT feature enables users to withdraw money from their wallet to their linked bank accounts. This complements the existing Bank Transfer IN (Wallet Load) functionality.

## Architecture

### Data Models

#### BankWalletTransfer (Database Model)
- **id**: UUID primary key
- **user_id**: Reference to user
- **phone_number**: User's phone number
- **amount**: Transfer amount
- **bank_account**: Destination bank account number
- **account_holder**: Account holder name
- **bank_code**: Bank identifier (IME, NMB, SCB)
- **bank_reference**: Unique reference from bank (idempotency key)
- **description**: Transfer description
- **status**: pending, processing, completed, failed, cancelled
- **failure_reason**: Reason if transfer failed
- **created_at, updated_at, completed_at**: Timestamps

#### UserBankAccount (Database Model)
- **id**: UUID primary key
- **user_id**: Reference to user
- **account_number**: Bank account number
- **account_holder**: Account holder name
- **bank_code**: Bank identifier
- **is_verified**: Account verification status
- **is_default**: Default account flag
- **created_at, updated_at**: Timestamps

### API Endpoints

#### 1. Initiate Bank Transfer
**Endpoint:** `POST /bank-api/v1/wallet/transfer`

**Authentication:** Bank API Key (Headers: `X-Bank-API-Key`, `X-Bank-Code`)

**Request:**
```json
{
  "phone_number": "+977-9800000000",
  "amount": 5000.00,
  "bank_account": "1234567890",
  "account_holder": "John Doe",
  "bank_code": "IME",
  "bank_reference": "TRANSFER-20260517-001",
  "description": "Withdrawal to bank account"
}
```

**Validation:**
- All fields required (phone_number, amount, bank_account, account_holder, bank_reference)
- amount > 0
- User must have sufficient balance
- Phone number must be verified
- bank_reference must be unique (for idempotency)

**Response (200 OK):**
```json
{
  "status": "pending",
  "transaction_id": "abc123def456",
  "phone_number": "+977-9800000000",
  "amount": 5000.00,
  "bank_account": "1234567890",
  "bank_reference": "TRANSFER-20260517-001",
  "wallet_balance": 45000.00,
  "timestamp": "2026-05-17T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Missing required fields
- `404 Not Found`: User not found with verified phone
- `402 Payment Required`: Insufficient balance
- `409 Conflict`: Duplicate bank_reference

#### 2. Verify Transfer Completion
**Endpoint:** `POST /bank-api/v1/wallet/transfer/verify`

**Called by Bank after processing transfer**

**Request:**
```json
{
  "bank_reference": "TRANSFER-20260517-001",
  "status": "completed",
  "timestamp": "2026-05-17T10:35:00Z",
  "signature": "hmac_signature"
}
```

**Response (200 OK):**
```json
{
  "status": "verified",
  "bank_reference": "TRANSFER-20260517-001",
  "message": "Wallet transfer completed successfully"
}
```

#### 3. Check Transfer Status
**Endpoint:** `GET /bank-api/v1/wallet/transfer/status?transaction_id=abc123def456`

**Response (200 OK):**
```json
{
  "transaction_id": "abc123def456",
  "status": "completed",
  "amount": 5000.00,
  "phone_number": "+977-9800000000",
  "bank_account": "1234567890",
  "bank_reference": "TRANSFER-20260517-001",
  "failure_reason": null,
  "created_at": "2026-05-17T10:30:00Z",
  "completed_at": "2026-05-17T10:35:00Z"
}
```

#### 4. Record Transfer Failure
**Endpoint:** `POST /bank-api/v1/wallet/transfer/failure`

**Called by Bank if transfer fails**

**Request:**
```json
{
  "bank_reference": "TRANSFER-20260517-001",
  "status": "failed",
  "reason": "Account closed",
  "timestamp": "2026-05-17T10:40:00Z"
}
```

**Response (200 OK):**
```json
{
  "status": "recorded",
  "bank_reference": "TRANSFER-20260517-001",
  "message": "Transfer failure recorded and balance refunded"
}
```

## Flow Diagram

```
User/Wallet App
  ↓ POST /bank-api/v1/wallet/transfer
     {phone, amount, bank_account, account_holder, bank_ref}
Bank API Service
  ↓ Validate user & phone_number
  ↓ Check sufficient balance
  ↓ Create bank_wallet_transfers record (status=pending)
  ↓ Deduct amount from user balance (reserve funds)
  ↓ Return TransactionID to user
  ↓ Return to user with pending status
User (Wallet Updated)
  ↓ Shows pending transfer

[User -> Bank Processing]

Bank System
  ↓ Processes transfer internally
  ↓ Either succeeds or fails
  ↓ Calls callback:

If Success:
  ↓ POST /bank-api/v1/wallet/transfer/verify
     {bank_reference: TRANSFER-001, status: completed}
Bank API Service
  ↓ Update bank_wallet_transfers (status=completed)
  ↓ Return acknowledgement to bank
User (Wallet Updated)
  ↓ Shows completed transfer

If Failure:
  ↓ POST /bank-api/v1/wallet/transfer/failure
     {bank_reference: TRANSFER-001, reason: "Account closed"}
Bank API Service
  ↓ Update bank_wallet_transfers (status=failed)
  ↓ Refund amount back to user balance
  ↓ Return acknowledgement to bank
User (Wallet Updated)
  ↓ Shows failed transfer, balance restored
```

## Key Features

### 1. **Balance Management**
- Funds are deducted immediately upon transfer initiation (pessimistic locking)
- If transfer fails, funds are automatically refunded

### 2. **Idempotency**
- Duplicate requests with same `bank_reference` return existing transaction
- Prevents double-transfers

### 3. **Transaction Safety**
- Database transactions ensure consistency
- If balance update fails, transfer record is rolled back

### 4. **Failure Handling**
- Automatic refund on transfer failure
- Stores failure reason for user reference

### 5. **Bank Authentication**
- Validates X-Bank-API-Key and X-Bank-Code headers
- Supports multiple banks (IME, NMB, SCB)

## Database Indexes

```
- idx_bank_transfer_user: Lookup transfers by user
- idx_bank_transfer_phone: Lookup transfers by phone
- idx_bank_transfer_reference: Lookup by bank_reference (unique)
- idx_bank_transfer_status: Filter by status
- idx_bank_transfer_created: Time-based queries
- idx_bank_transfer_account: Lookup by bank account
```

## Migration Steps

1. Run migration SQL to create tables:
   ```bash
   psql -U $DB_USER -d $DB_NAME -f migrations/002_add_bank_transfer_out.sql
   ```

2. Tables created:
   - `bank_wallet_transfers`: Stores transfer records
   - `user_bank_accounts`: Stores user's bank account details

## Statuses Explained

| Status | Description | User Balance | Next Action |
|--------|-------------|--------------|-------------|
| pending | Transfer initiated, awaiting bank processing | Reduced | Wait for bank callback |
| processing | Bank is processing the transfer | Reduced | Wait for bank callback |
| completed | Transfer successfully completed | Reduced | Display success |
| failed | Transfer failed, funds refunded | Restored | Show error reason |
| cancelled | Transfer cancelled by user/system | Restored | Show cancellation |

## Testing

### Test Initiate Transfer
```bash
curl -X POST http://localhost:8082/bank-api/v1/wallet/transfer \
  -H "Content-Type: application/json" \
  -H "X-Bank-API-Key: ime-api-key-placeholder" \
  -H "X-Bank-Code: IME" \
  -d '{
    "phone_number": "+977-9800000000",
    "amount": 5000.00,
    "bank_account": "1234567890",
    "account_holder": "John Doe",
    "bank_code": "IME",
    "bank_reference": "TRANSFER-20260517-001",
    "description": "Test withdrawal"
  }'
```

### Test Verify Transfer
```bash
curl -X POST http://localhost:8082/bank-api/v1/wallet/transfer/verify \
  -H "Content-Type: application/json" \
  -d '{
    "bank_reference": "TRANSFER-20260517-001",
    "status": "completed",
    "timestamp": "2026-05-17T10:35:00Z",
    "signature": "hmac_signature"
  }'
```

### Test Check Status
```bash
curl -X GET "http://localhost:8082/bank-api/v1/wallet/transfer/status?transaction_id=abc123def456"
```

### Test Handle Failure
```bash
curl -X POST http://localhost:8082/bank-api/v1/wallet/transfer/failure \
  -H "Content-Type: application/json" \
  -d '{
    "bank_reference": "TRANSFER-20260517-001",
    "status": "failed",
    "reason": "Insufficient account balance",
    "timestamp": "2026-05-17T10:40:00Z"
  }'
```

## Frontend Integration (Optional)

To add bank transfer UI to the React Native app:

1. Add "Withdraw" button in WalletScreen
2. Show bank account selection
3. Allow entering transfer amount
4. Show transfer history with status badges

## Security Considerations

1. **HMAC Signature Verification** (TODO in production)
   - Implement HMAC-SHA256 validation for bank callbacks
   
2. **Rate Limiting**
   - Implement rate limiting on transfer endpoint
   - Prevent abuse with per-user transfer limits

3. **Account Verification**
   - Implement 2FA for bank transfer initiation
   - Send OTP to phone before processing

4. **Audit Logging**
   - Log all transfer operations for compliance
   - Track failure reasons and refunds

## Next Steps

1. ✅ Database migration created
2. ✅ Models added
3. ✅ API endpoints implemented
4. ⏳ Frontend UI for bank transfers (optional)
5. ⏳ Integration tests
6. ⏳ HMAC signature verification
7. ⏳ Rate limiting middleware
8. ⏳ 2FA for transfer confirmation
9. ⏳ Webhook notifications to mobile app
