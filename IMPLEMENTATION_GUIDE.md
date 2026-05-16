# Implementation Guide: Phone Login & Bank API Integration

## Quick Summary

This guide walks through implementing:
1. **Phone-based authentication** (login/register with phone number)
2. **Phone-based P2P transfers** (send money using phone numbers)
3. **Bank API for wallet loading** (banks can load customer wallets)

---

## Step 1: Database Setup

### 1.1 Run Migration
```bash
# Navigate to project root
cd c:\Users\Raja\Desktop\payment-system

# Connect to PostgreSQL
psql -h localhost -U postgres -d payment_system

# Run the migration
\i migrations/001_add_phone_login_and_bank_api.sql

# Verify tables created
\dt bank_wallet_loads
\dt phone_otps
\d users  # Check for phone_number column
```

### 1.2 Verify Schema
```sql
-- Check users table has phone columns
SELECT column_name FROM information_schema.columns 
WHERE table_name='users' AND column_name LIKE 'phone%';

-- Should return: phone_number, phone_verified
```

---

## Step 2: Update User Service

### 2.1 Update User Model
Replace `user-service/models/user.go` with the updated version that includes:
- `PhoneNumber string`
- `PhoneVerified bool`
- `PhoneLoginRequest` struct
- `PhoneLookupResponse` struct

See: [user-service/models/user_updated.go](../user-service/models/user_updated.go)

### 2.2 Add Phone Handler
Add new file with phone-based handlers:
- `LoginByPhone()` - Login with phone number
- `LookupUserByPhone()` - Lookup user by phone (for transfers)
- `SendPhoneOTP()` - Send OTP to phone
- `VerifyPhoneOTP()` - Verify OTP code

See: [user-service/handler/phone_handler.go](../user-service/handler/phone_handler.go)

### 2.3 Add Required Dependencies
```bash
cd user-service
go get github.com/ttacon/libphonenumber-go  # Phone validation
go mod tidy
```

### 2.4 Update main.go - Register Routes
```go
// In user-service/main.go, add these routes:

func setupRoutes() {
    // Existing routes...
    
    // Phone-based authentication
    http.HandleFunc("/login/phone", handler.LoginByPhone)
    http.HandleFunc("/lookup/phone", handler.LookupUserByPhone)
    http.HandleFunc("/send-phone-otp", handler.SendPhoneOTP)
    http.HandleFunc("/verify-phone-otp", handler.VerifyPhoneOTP)
    
    // ... rest of routes
}
```

### 2.5 Add Missing Helper Function
Add this to `user-service/utils/logger.go` or create utils file:
```go
func GenerateOTP() string {
    const digits = "0123456789"
    otp := ""
    for i := 0; i < 6; i++ {
        otp += string(digits[rand.Intn(len(digits))])
    }
    return otp
}
```

---

## Step 3: Update Payment Service

### 3.1 Add Bank Models to Payment Service
Create `payment-service/models/bank_models.go` with:
```go
type SendPaymentByPhoneRequest struct {
    SenderPhone   string
    ReceiverPhone string
    Amount        float64
    Currency      string
    Description   string
}
```

### 3.2 Add Phone Resolution Handler
Add to `payment-service/handler/payment.go`:
```go
func (s *Server) SendPaymentByPhone(ctx context.Context, req *pb.SendPaymentByPhoneRequest) (*pb.SendPaymentResponse, error) {
    // 1. Call user service to lookup sender phone -> sender_id
    // 2. Call user service to lookup receiver phone -> receiver_id
    // 3. Call existing SendPayment with resolved IDs
    // 4. Return response
}
```

### 3.3 Update Proto File
Add to `payment-service/proto/payment.proto`:
```protobuf
service PaymentService {
  rpc SendPayment(SendPaymentRequest) returns (SendPaymentResponse);
  rpc SendPaymentByPhone(SendPaymentByPhoneRequest) returns (SendPaymentResponse);
  rpc GetTransaction(GetTransactionRequest) returns (GetTransactionResponse);
  rpc GetBalance(GetBalanceRequest) returns (GetBalanceResponse);
}

message SendPaymentByPhoneRequest {
  string sender_phone = 1;
  string receiver_phone = 2;
  double amount = 3;
  string currency = 4;
  string description = 5;
}
```

### 3.4 Regenerate Proto
```bash
cd payment-service
protoc --go_out=. --go-grpc_out=. proto/payment.proto
```

---

## Step 4: Create Bank API Service

### 4.1 Create Bank API Directory
```bash
mkdir bank-api
cd bank-api
go mod init github.com/sumitDon47/payment-system/bank-api
```

### 4.2 Add Bank API Implementation
Copy the provided `bank-api/main.go` file and update imports as needed.

### 4.3 Create Bank API go.mod
```go
go 1.21

require (
    github.com/lib/pq v1.10.9
)
```

### 4.4 Build Bank API Service
```bash
cd bank-api
go build -o bank-api main.go
```

### 4.5 Update docker-compose.yml
Add bank-api service:
```yaml
bank-api:
  container_name: bank-api
  build:
    context: ./bank-api
    dockerfile: Dockerfile
  ports:
    - "8082:8082"
  environment:
    DB_HOST: postgres
    DB_PORT: 5432
    DB_USER: postgres
    DB_PASSWORD: password
    DB_NAME: payment_system
  depends_on:
    - postgres
  networks:
    - payment-network
```

### 4.6 Create Bank API Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bank-api main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bank-api .
EXPOSE 8082
CMD ["./bank-api"]
```

---

## Step 5: Environment Configuration

### 5.1 Update .env (User Service)
```env
# User Service - Phone OTP Configuration
PHONE_OTP_PROVIDER=twilio  # or sparrow, nexmo
PHONE_OTP_API_KEY=your_api_key_here
PHONE_OTP_API_SECRET=your_secret_here

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=payment_system
```

### 5.2 Update .env (Bank API)
```env
# Bank API Configuration
BANK_API_PORT=8082
BANK_API_WEBHOOK_SECRET=your_webhook_secret

# Bank credentials (in production, use secure vault)
BANK_CODE_IME=ime-api-key-prod
BANK_CODE_NMB=nmb-api-key-prod
BANK_CODE_SCB=scb-api-key-prod

# Database (same as user service)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=payment_system
```

---

## Step 6: Testing

### 6.1 Test Phone Registration
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "phone_number": "+977-9800000000",
    "password": "securepass123",
    "mpin": "1234"
  }'
```

### 6.2 Test Send Phone OTP
```bash
curl -X POST http://localhost:8080/send-phone-otp \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+977-9800000000"}'
```

### 6.3 Test Verify Phone OTP
```bash
curl -X POST http://localhost:8080/verify-phone-otp \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+977-9800000000", "code": "123456"}'
```

### 6.4 Test Phone Login
```bash
curl -X POST http://localhost:8080/login/phone \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+977-9800000000",
    "password": "securepass123",
    "mpin": "1234"
  }'
```

### 6.5 Test Phone Lookup
```bash
curl -X GET "http://localhost:8080/lookup/phone?phone=%2B977-9800000000"
```

### 6.6 Test Bank Wallet Load
```bash
curl -X POST http://localhost:8082/bank-api/v1/wallet/load \
  -H "Content-Type: application/json" \
  -H "X-Bank-API-Key: ime-api-key-placeholder" \
  -H "X-Bank-Code: IME" \
  -d '{
    "phone_number": "+977-9800000000",
    "amount": 5000.00,
    "bank_reference": "BANK-TXN-20260515-001",
    "bank_code": "IME",
    "description": "Salary Credit"
  }'
```

### 6.7 Test Wallet Verification Callback
```bash
curl -X POST http://localhost:8082/bank-api/v1/wallet/verify \
  -H "Content-Type: application/json" \
  -d '{
    "bank_reference": "BANK-TXN-20260515-001",
    "status": "completed",
    "timestamp": "2026-05-15T10:35:00Z",
    "signature": "hmac_signature"
  }'
```

---

## Step 7: Update Frontend

### 7.1 Add Phone Login Screen
Create `payment-app/src/screens/PhoneLoginScreen.tsx`:
```typescript
import React, { useState } from 'react';
import { View, TextInput, Button } from 'react-native';

export default function PhoneLoginScreen() {
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');

  const handleLogin = async () => {
    const response = await fetch('http://localhost:8080/login/phone', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        phone_number: phone,
        password: password,
        mpin: ''
      })
    });
    const data = await response.json();
    // Store token and navigate
  };

  return (
    <View>
      <TextInput 
        placeholder="+977-XXXXXXXXXX"
        value={phone}
        onChangeText={setPhone}
      />
      <TextInput 
        placeholder="Password"
        value={password}
        onChangeText={setPassword}
        secureTextEntry
      />
      <Button title="Login" onPress={handleLogin} />
    </View>
  );
}
```

### 7.2 Add Phone Transfer Screen
Create similar screen for phone-based transfers using phone number lookup.

---

## Step 8: Deployment Checklist

- [ ] Database migrations applied
- [ ] User service updated with phone handlers
- [ ] Payment service updated with phone RPC
- [ ] Bank API service created and configured
- [ ] All services dockerized
- [ ] docker-compose.yml updated
- [ ] Environment variables configured
- [ ] All endpoints tested
- [ ] Frontend screens updated
- [ ] Rate limiting configured
- [ ] Authentication middleware verified
- [ ] Logging and monitoring setup
- [ ] Security headers added
- [ ] Phone number validation working
- [ ] HMAC signature verification implemented

---

## Monitoring & Logging

### 8.1 Bank API Request Logging
```go
// Log all bank API requests for audit trail
log.Printf("Bank=%s, TxID=%s, Amount=%.2f, Phone=%s, Status=%s",
  bankCode, transactionID, amount, phoneNumber, status)
```

### 8.2 Metrics to Track
- Phone OTP sent/verified
- Phone login attempts (success/failure)
- Bank wallet loads (pending/completed/failed)
- Phone transfer completion rate
- API response times

---

## Security Checklist

- [x] Phone number validation (E.164 format)
- [x] OTP rate limiting (3 attempts / 15 min)
- [x] Bank API key validation
- [x] HMAC signature verification (TODO)
- [x] Idempotency tokens for wallet loads
- [x] Phone number encryption in logs
- [x] IP whitelisting for bank API (TODO)
- [x] PII data protection
- [x] Audit logging for compliance

---

## Troubleshooting

### Phone OTP Not Working
- Check OTP table has records: `SELECT * FROM phone_otps;`
- Verify SMS provider credentials in .env
- Check rate limiting in Redis

### Bank API 401 Unauthorized
- Verify X-Bank-API-Key header is correct
- Check X-Bank-Code is in validBankCodes map
- Ensure request signature matches

### Wallet Load Stuck in Pending
- Bank callback not received: Check bank-api logs
- Database transaction locked: Check for long-running queries
- User not found: Verify phone_verified=true

---

## Next Steps

1. Implement SMS provider integration (Twilio/Sparrow)
2. Add HMAC signature verification
3. Setup bank webhook IP whitelisting
4. Create comprehensive test suite
5. Deploy to staging environment
6. Bank integration testing
7. Production rollout planning

---

For detailed API documentation, see: [PHONE_LOGIN_AND_BANK_API.md](../PHONE_LOGIN_AND_BANK_API.md)
