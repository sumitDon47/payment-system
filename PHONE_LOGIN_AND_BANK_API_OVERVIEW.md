# Phone Login & Bank API - Complete Overview

## What We're Building

You now have everything to implement three major features:

### 1. **Phone-Based Login & Registration**
- Users can register with phone number instead of email
- Login using phone number + password + MPIN
- OTP verification for phone authenticity

### 2. **Phone-to-Phone Money Transfer (P2P)**
- Send money to another user by their phone number
- Automatic phone-to-user-ID lookup
- Same security as existing transfers

### 3. **Bank API for Wallet Top-up**
- Banks can load money into user wallets
- Direct integration at port 8082
- Idempotent transactions (banks can retry safely)
- Webhook callbacks for real-time status updates

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Mobile App (React Native)                 │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │  Phone Login   │  │  Phone Trans │  │  Bank Top-up    │ │
│  └────────────────┘  └──────────────┘  └─────────────────┘ │
└────────────────────┬────────────────────────────────────────┘
                     │
         ┌───────────┴────────────┬──────────────────┐
         │                        │                  │
    ┌────▼──────┐         ┌──────▼────┐      ┌──────▼──────┐
    │ User Srv  │         │Payment Srv │     │ Bank API    │
    │ :8080     │         │  :9090     │     │ :8082       │
    │ (REST)    │         │  (gRPC)    │     │ (REST)      │
    └────┬──────┘         └──────┬─────┘     └──────┬──────┘
         │                       │                   │
         │  Phone Handlers:      │  Phone Payment:   │  Bank Wallet:
         │  • LoginByPhone       │  • SendByPhone    │  • /wallet/load
         │  • LookupByPhone      │                   │  • /wallet/verify
         │  • SendPhoneOTP       │  Outbox Pattern:  │  • /wallet/status
         │  • VerifyPhoneOTP     │  • Serializable TX│  • /wallet/failure
         │                       │  • Kafka publish  │
         │                       │  • Dead-letter Q  │
         └───────────────────────┴───────────────────┘
                     │
         ┌───────────┴──────────────────┐
         │                              │
    ┌────▼──────┐              ┌───────▼───────┐
    │ PostgreSQL│              │ Kafka Broker  │
    │ • users   │              │ • payment.    │
    │ • trans   │              │   completed   │
    │ • bank_   │              │ • payment.    │
    │   wallet_ │              │   failed      │
    │   loads   │              └───────┬───────┘
    │ • phone_  │                      │
    │   otps    │              ┌───────▼──────────┐
    └───────────┘              │Notification Srv │
                               │ • Email/SMS send │
                               └──────────────────┘
```

---

## Files Created/Modified

### **Documentation Files**
- `PHONE_LOGIN_AND_BANK_API.md` - Complete API specification
- `IMPLEMENTATION_GUIDE.md` - Step-by-step implementation
- `README.md` - This file

### **User Service Files**
- `user-service/models/user_updated.go` - Updated user model with phone fields
- `user-service/models/bank_models.go` - Bank-related data models
- `user-service/handler/phone_handler.go` - Phone authentication handlers

### **Payment Service Files**
- `payment-service/models/bank_models.go` - Bank transaction models

### **Bank API Service** (NEW)
- `bank-api/main.go` - Complete bank API implementation
- `bank-api/Dockerfile` - Container configuration

### **Database**
- `migrations/001_add_phone_login_and_bank_api.sql` - All schema updates

---

## Database Schema Changes

### New Tables
```
phone_otps
├── phone_number (indexed)
├── code
├── expires_at
└── attempts

bank_wallet_loads
├── user_id (FK to users)
├── phone_number (indexed)
├── amount
├── bank_reference (unique, indexed)
├── status (pending/completed/failed/reversed)
└── timestamps

phone_transfer_logs
├── sender_phone
├── receiver_phone
├── transaction_id (FK to transactions)
└── timestamps

bank_api_audit_logs
├── bank_code
├── request_type
├── bank_reference
└── timestamps
```

### Modified Tables
```
users
├── + phone_number (UNIQUE)
├── + phone_verified (BOOLEAN)
└── indexed for lookups
```

---

## API Endpoints

### **User Service (REST) - Port 8080**

#### Authentication
```
POST   /register              - Register with phone number
POST   /login/phone           - Login with phone number
POST   /send-phone-otp        - Send OTP to phone
POST   /verify-phone-otp      - Verify OTP code
```

#### User Lookup
```
GET    /lookup/phone?phone=   - Lookup user by phone (for transfers)
```

### **Bank API (REST) - Port 8082**

#### Wallet Operations
```
POST   /bank-api/v1/wallet/load         - Load wallet from bank
POST   /bank-api/v1/wallet/verify       - Verify & complete load
GET    /bank-api/v1/wallet/status       - Check load status
POST   /bank-api/v1/wallet/failure      - Record failure
```

### **Payment Service (gRPC) - Port 9090**

#### Transfer Operations
```
SendPaymentByPhone(SenderPhone, ReceiverPhone, Amount) -> TransactionID
```

---

## Implementation Sequence

### **Phase 1: Database & Models** (1-2 hours)
1. Run SQL migrations
2. Update user model with phone fields
3. Create bank models
4. Verify schema

### **Phase 2: User Service** (2-3 hours)
1. Implement phone handlers (4 functions)
2. Add phone lookup endpoint
3. Register routes in main.go
4. Add phone OTP SMS integration
5. Test all phone endpoints

### **Phase 3: Payment Service** (1-2 hours)
1. Add phone-based RPC to proto
2. Implement phone-to-ID resolution
3. Integrate with existing payment flow
4. Test phone transfers

### **Phase 4: Bank API Service** (2-3 hours)
1. Create bank-api service directory
2. Implement wallet load endpoints
3. Add bank authentication middleware
4. Add to docker-compose
5. Test all bank endpoints

### **Phase 5: Frontend** (2-3 hours)
1. Create phone login screen
2. Create phone transfer screen
3. Update navigation
4. Integrate API calls
5. Test UI flows

### **Phase 6: Testing & Security** (2-3 hours)
1. Unit tests for phone handlers
2. Integration tests for transfers
3. Bank API security testing
4. Rate limiting tests
5. Idempotency tests

---

## Data Flow Examples

### **Scenario 1: Phone-Based Login**
```
User (App)
  ↓ POST /login/phone (+977-9800000000, password, mpin)
User Service
  ↓ Query users WHERE phone_number = '+977-9800000000'
Database (PostgreSQL)
  ↓ Return user record
User Service
  ↓ Verify password with bcrypt
  ↓ Generate JWT token
  ↓ Return token + user profile
User (App)
  ↓ Store token in AsyncStorage
  ↓ Navigate to wallet screen
```

### **Scenario 2: Phone-Based Money Transfer**
```
User A (App)
  ↓ Enter recipient phone: +977-9801111111, Amount: 1000
  ↓ gRPC SendPaymentByPhone(+977-9800000000, +977-9801111111, 1000)
Payment Service
  ↓ Call User Service: LookupByPhone(+977-9800000000) -> UserA_ID
  ↓ Call User Service: LookupByPhone(+977-9801111111) -> UserB_ID
  ↓ Execute SendPayment(UserA_ID, UserB_ID, 1000)
Payment Service
  ↓ Begin transaction (SERIALIZABLE)
  ↓ Check balance, lock rows
  ↓ Debit UserA, Credit UserB
  ↓ Create outbox record
  ↓ Commit transaction
  ↓ Publish to Kafka
  ↓ Return TransactionID + Status
User A (App)
  ↓ Display success message
User B
  ↓ Notification Service consumes Kafka event
  ↓ Sends notification to User B
  ↓ User B receives SMS/Email notification
```

### **Scenario 3: Bank Wallet Top-up**
```
Bank System
  ↓ POST /bank-api/v1/wallet/load
     {phone: +977-9800000000, amount: 5000, bank_ref: BANK-001}
Bank API Service
  ↓ Validate API key & bank code
  ↓ Lookup user by phone_number
  ↓ Create bank_wallet_loads record (status=pending)
  ↓ Return TransactionID to bank
Bank System
  [... internal processing ...]
  ↓ POST /bank-api/v1/wallet/verify (callback)
     {bank_reference: BANK-001, status: completed}
Bank API Service
  ↓ Begin transaction
  ↓ Update user balance (balance += 5000)
  ↓ Update bank_wallet_loads (status=completed)
  ↓ Commit
  ↓ Return acknowledgement to bank
User (App)
  ↓ User Service notified of balance update
  ↓ Wallet screen shows +5000
User B (Mobile)
  ↓ Receives notification: "₹5000 added to wallet"
```

---

## Security Implementation

### **Phone Number Validation**
```go
// E.164 Format: +{CountryCode}{Number}
+977-9800000000 ✓ Valid
9800000000      ✗ Invalid
977-9800000000  ✗ Invalid
```

### **Rate Limiting**
```
Phone OTP:     3 attempts per 15 minutes
Phone Login:   5 failures per 30 minutes
Bank API:      10 requests per hour per bank
```

### **Bank API Authentication**
```
Headers Required:
- X-Bank-API-Key: {api_key}
- X-Bank-Code: {IME|NMB|SCB}
- X-Signature: HMAC-SHA256(request_body)

Idempotency:
- bank_reference acts as idempotency key
- Duplicate requests return same response
```

---

## Testing Checklist

### **Unit Tests**
- [ ] Phone validation (valid/invalid formats)
- [ ] OTP generation and verification
- [ ] Phone login flow
- [ ] Phone lookup
- [ ] Bank reference validation

### **Integration Tests**
- [ ] Register → Login with phone
- [ ] Send OTP → Verify → Login
- [ ] Phone-based transfer
- [ ] Bank wallet load
- [ ] Webhook verification

### **Security Tests**
- [ ] Invalid API keys rejected
- [ ] Rate limiting enforced
- [ ] Signature verification works
- [ ] Duplicate requests handled
- [ ] Invalid phone formats rejected

---

## Monitoring & Observability

### **Metrics to Track**
```
User Service:
- phone_login_attempts_total
- phone_otp_sent_total
- phone_otp_verified_total
- phone_lookup_requests_total

Payment Service:
- phone_transfers_total
- phone_transfers_failed
- phone_transfer_duration_seconds

Bank API:
- wallet_load_requests_total
- wallet_load_completed_total
- wallet_load_failed_total
- bank_api_response_time_seconds
- bank_api_errors_total
```

### **Logging**
```
All events logged to structured logs with:
- timestamp
- service name
- request_id/correlation_id
- phone_number (encrypted in production)
- bank_reference (if applicable)
- status
- error (if applicable)
```

---

## Production Deployment

### **Pre-deployment Checklist**
- [ ] All migrations tested on staging
- [ ] All services built and tested
- [ ] Environment variables configured
- [ ] SSL certificates configured
- [ ] Bank credentials stored in secret manager
- [ ] Database backups verified
- [ ] Monitoring and alerting setup
- [ ] Runbook created for bank API issues
- [ ] Load testing completed
- [ ] Security audit passed

### **Deployment Steps**
```bash
# 1. Backup database
pg_dump payment_system > backup_$(date +%Y%m%d).sql

# 2. Apply migrations
psql -f migrations/001_add_phone_login_and_bank_api.sql

# 3. Update services
docker-compose pull
docker-compose up -d

# 4. Verify services
curl http://localhost:8080/health
curl http://localhost:9090/health
curl http://localhost:8082/health

# 5. Test critical flows
# Run integration tests
# Verify phone login works
# Verify bank API works

# 6. Monitor
tail -f logs/user-service.log
tail -f logs/payment-service.log
tail -f logs/bank-api.log
```

---

## Support & Contact

For questions during implementation:
1. Check [IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md)
2. Check [PHONE_LOGIN_AND_BANK_API.md](./PHONE_LOGIN_AND_BANK_API.md)
3. Review example test cases
4. Check troubleshooting section

---

## Version History

- **v1.0** - Initial implementation guide
  - Phone-based login/registration
  - Phone-based money transfer
  - Bank wallet top-up API
  - All models and handlers
  - Database migrations
  - Documentation

---

**Last Updated:** May 15, 2026
**Status:** Ready for Implementation
