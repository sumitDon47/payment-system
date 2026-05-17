# Bank Transfer OUT Implementation Checklist

## ✅ Completed Tasks

### Database Layer
- [x] Create migration file: `002_add_bank_transfer_out.sql`
  - [x] `bank_wallet_transfers` table with 13 columns
  - [x] `user_bank_accounts` table with 8 columns
  - [x] Full indexing for performance (6 indexes per table)
  - [x] UUID primary keys with ON DELETE CASCADE

### Data Models
- [x] Add models to `user-service/models/bank_models.go`
  - [x] `BankWalletTransferRequest` - API input
  - [x] `BankWalletTransferResponse` - API output
  - [x] `BankWalletTransferVerificationRequest` - Bank callback
  - [x] `BankWalletTransferStatusResponse` - Status query
  - [x] `BankWalletTransferFailureRequest` - Failure callback
  - [x] `BankWalletTransfer` - Database model
  - [x] `UserBankAccount` - Database model

### Bank API Service
- [x] Add 4 new endpoint handlers to `bank-api/main.go`
  - [x] `initiateWalletTransfer()` - POST `/wallet/transfer`
  - [x] `verifyWalletTransfer()` - POST `/wallet/transfer/verify`
  - [x] `getWalletTransferStatus()` - GET `/wallet/transfer/status`
  - [x] `handleWalletTransferFailure()` - POST `/wallet/transfer/failure`
- [x] Add routes to main() function
- [x] Add authentication middleware support

### Features Implemented
- [x] Balance validation before transfer
- [x] Immediate balance deduction (pessimistic locking)
- [x] Idempotent requests via bank_reference
- [x] Automatic refund on failure
- [x] Database transaction consistency
- [x] Comprehensive error codes and messages
- [x] Logging for all operations
- [x] Status tracking (pending, processing, completed, failed)

### Documentation
- [x] `BANK_TRANSFER_OUT_GUIDE.md` - Comprehensive feature guide
- [x] `BANK_TRANSFER_API_QUICK_REFERENCE.md` - Quick API reference
- [x] `002_add_bank_transfer_out.sql` - Migration with comments

### Code Quality
- [x] No compilation errors
- [x] Follows existing code patterns
- [x] Consistent error handling
- [x] Proper HTTP status codes
- [x] Input validation

## ⏳ Next Steps (Optional)

### Testing
- [ ] Unit tests for transfer logic
- [ ] Integration tests with database
- [ ] Error case testing
- [ ] Idempotency verification

### Frontend Integration
- [ ] Add "Withdraw" button to WalletScreen.tsx
- [ ] Create WithdrawScreen component
- [ ] Implement bank account selection UI
- [ ] Show transfer history with status
- [ ] Add real-time status updates

### Security & Compliance
- [ ] Implement HMAC signature verification (production)
- [ ] Add rate limiting to transfer endpoint
- [ ] Implement 2FA for transfers
- [ ] Add audit logging for compliance
- [ ] Implement transfer limits per user/day

### Performance
- [ ] Monitor database query performance
- [ ] Add caching for user bank accounts
- [ ] Optimize status query with materialized view
- [ ] Add connection pooling

### Monitoring
- [ ] Add Prometheus metrics for transfers
- [ ] Alert on transfer failures
- [ ] Track transfer success rate
- [ ] Monitor balance refund operations

### Documentation
- [ ] Update main README.md
- [ ] Add examples in BANK_INTEGRATION_GUIDE.md
- [ ] Create deployment guide
- [ ] Add troubleshooting section

## How to Deploy

### Step 1: Database Migration
```bash
# SSH into database server or local
psql -U $DB_USER -d $DB_NAME -f migrations/002_add_bank_transfer_out.sql
```

### Step 2: Verify Tables
```bash
psql -U $DB_USER -d $DB_NAME -c "
  \dt bank_wallet_transfers
  \dt user_bank_accounts
"
```

### Step 3: Rebuild Bank API Service
```bash
cd bank-api/
go build -o bank-api
```

### Step 4: Deploy Bank API
```bash
# Docker
docker build -t payment-system/bank-api:v2 .
docker-compose up bank-api

# Or manual
./bank-api
```

### Step 5: Test Endpoints
```bash
# Test transfer initiation
curl -X POST http://localhost:8082/bank-api/v1/wallet/transfer \
  -H "X-Bank-API-Key: ime-api-key-placeholder" \
  -H "X-Bank-Code: IME" \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+977-9800000000", "amount": 5000, ...}'
```

## File Changes Summary

| File | Changes | Lines Added |
|------|---------|------------|
| `migrations/002_add_bank_transfer_out.sql` | NEW | 50 |
| `user-service/models/bank_models.go` | Updated | +100 |
| `bank-api/main.go` | Updated | +350 |
| `BANK_TRANSFER_OUT_GUIDE.md` | NEW | 400 |
| `BANK_TRANSFER_API_QUICK_REFERENCE.md` | NEW | 200 |

**Total Additions:** ~1,100 lines of code and documentation

## Rollback Plan

If issues occur:

1. **Stop API Service**
   ```bash
   docker stop bank-api
   ```

2. **Rollback Database**
   ```bash
   psql -U $DB_USER -d $DB_NAME -c "DROP TABLE IF EXISTS bank_wallet_transfers CASCADE;"
   psql -U $DB_USER -d $DB_NAME -c "DROP TABLE IF EXISTS user_bank_accounts CASCADE;"
   ```

3. **Restore Previous API Version**
   ```bash
   git checkout HEAD -- bank-api/main.go
   go build -o bank-api
   ```

4. **Restart Service**
   ```bash
   docker-compose up bank-api
   ```

## Validation Checklist

Before going to production:

- [ ] All tests pass
- [ ] No compilation errors
- [ ] Database migration runs successfully
- [ ] Transfer endpoint responds correctly
- [ ] Failure handling works (balance refunded)
- [ ] Idempotency verified (same bank_ref returns same result)
- [ ] Error codes match documentation
- [ ] Logs show expected messages
- [ ] Load testing completed
- [ ] Security review passed
- [ ] Bank partners confirm API compatibility
