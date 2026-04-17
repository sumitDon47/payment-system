# Payment System API Integration Guide

## Overview

The Payment System is a microservices architecture consisting of three services providing payment processing, user management, and notifications.

```
┌─────────────────┐         ┌──────────────────┐         ┌──────────────────┐
│  User Service   │         │ Payment Service  │         │Notification Svc  │
│   (HTTP/REST)   │         │   (gRPC)         │         │   (Kafka)        │
│   Port 8080     │         │   Port 9090      │         │   Internal       │
└────────┬────────┘         └────────┬─────────┘         └────────┬─────────┘
         │                           │                             │
         │  Register/Login           │  SendPayment RPC           │
         │  Profile/Wallet           │  GetTransaction RPC        │  Consumes
         │  Rate: 5/100 req/min      │  GetBalance RPC            │  Payment
         │                           │  Rate: 100 req/sec         │  Events
         │                           │  (SendPayment)             │
         │                           │  1000 req/min              │
         │                           │  (GetTransaction)          │
         │                           │                             │
         │◄──────────────────────────┤                            │
         │  Invalidate Cache         │                            │
         │  (after transaction)      │                            │
         │                           │◄───────────────────────────┤
         │                           │  Email Notifications       │
         │                           │  (via Kafka)               │
         │                           │                             │
```

## Quick Start

### 1. Register a User

**HTTP POST** to User Service:

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "email": "alice@example.com",
    "password": "SecurePass123!"
  }'
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "name": "Alice",
  "balance": 0.00,
  "created_at": "2026-04-17T10:30:00Z"
}
```

### 2. Login to Get JWT Token

**HTTP POST** to User Service:

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "SecurePass123!"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 3. Check Wallet Balance

**HTTP GET** to User Service (authenticated):

```bash
curl http://localhost:8080/wallet \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "balance": 1000.00,
  "currency": "NPR"
}
```

### 4. Send Payment (gRPC)

**gRPC RPC** to Payment Service:

```bash
grpcurl -d '{
  "sender_id": "550e8400-e29b-41d4-a716-446655440000",
  "receiver_id": "660e8400-e29b-41d4-a716-446655440001",
  "amount": 100.50,
  "currency": "NPR",
  "note": "Payment for services"
}' \
  -plaintext localhost:9090 \
  payment.PaymentService/SendPayment
```

**Response:**
```json
{
  "transaction_id": "txn_550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "sender_balance": 899.50,
  "message": "Payment successful",
  "created_at": "2026-04-17T10:35:00Z"
}
```

---

## Service Details

### User Service (HTTP REST)

**Base URL**: `http://localhost:8080`  
**Port**: `8080`  
**Authentication**: JWT Bearer Token

#### Endpoints

| Method | Path | Auth | Rate Limit | Purpose |
|--------|------|------|-----------|---------|
| GET | `/health` | No | - | Health check |
| POST | `/register` | No | 5/min | Register new user |
| POST | `/login` | No | 5/min | Get JWT token |
| GET | `/profile` | Yes | 100/min | Get user profile |
| GET | `/wallet` | Yes | 100/min | Get wallet balance |
| POST | `/internal/cache/invalidate` | No | - | Invalidate cache (internal) |

#### Authentication

Include JWT token in `Authorization` header:

```
Authorization: Bearer <token>
```

**Token Details**:
- Algorithm: HS256
- Duration: 1 hour (3600 seconds)
- Environment Variable: `JWT_SECRET` (set in `.env`)

#### Rate Limiting

- **Auth Endpoints** (`/register`, `/login`): 5 requests/minute per IP
- **API Endpoints** (`/profile`, `/wallet`): 100 requests/minute per IP
- **Response on Limit**: HTTP 429 with `Retry-After: 60` header

#### Dependencies

- **PostgreSQL**: User data, authentication
- **Redis** (optional): Wallet balance cache (5-minute TTL)
- **Environment Variables**:
  - `PORT`: Server port (default: 8080)
  - `DATABASE_URL`: PostgreSQL connection string
  - `REDIS_URL`: Redis connection string (optional)
  - `JWT_SECRET`: Secret key for signing tokens

---

### Payment Service (gRPC)

**Server**: `localhost:9090`  
**Port**: `9090`  
**Protocol**: gRPC over HTTP/2

#### RPC Methods

| Method | Input | Output | Rate Limit |
|--------|-------|--------|-----------|
| `SendPayment` | SendPaymentRequest | SendPaymentResponse | 100 req/sec/user |
| `GetTransaction` | GetTransactionRequest | GetTransactionResponse | 1000 req/min/user |
| `GetBalance` | GetBalanceRequest | GetBalanceResponse | 1000 req/min/user |

#### Rate Limiting

- **SendPayment**: 100 requests/second per user (burst: 5)
- **GetTransaction**: 1000 requests/minute per user (burst: 10)
- **GetBalance**: 1000 requests/minute per user (burst: 10)
- **Response on Limit**: gRPC `RESOURCE_EXHAUSTED` error

#### Guaranteed Properties

- **ACID Transactions**: SERIALIZABLE isolation with row-level locking
- **Double-Spend Prevention**: Only one concurrent payment succeeds per sender
- **Event Sourcing**: Outbox pattern for reliable Kafka event publication
- **Idempotent**: Transaction IDs can be used to detect duplicates

#### Dependencies

- **PostgreSQL**: Transaction log, user balances, outbox
- **Kafka**: Reliable event publication
- **Environment Variables**:
  - `GRPC_PORT`: Server port (default: 9090)
  - `DATABASE_URL`: PostgreSQL connection string
  - `KAFKA_BROKER`: Kafka broker address (e.g., localhost:29092)
  - `KAFKA_TOPIC`: Payment event topic (default: payment-events)
  - `KAFKA_DLQ_TOPIC`: Dead letter queue topic (default: payment-dlq)

---

### Notification Service (Kafka Consumer)

**Type**: Internal service  
**Transport**: Kafka messages  
**Topics Consumed**: `payment-events`

#### Functionality

- Consumes payment completion events from Kafka
- Sends email notifications via SendGrid
- Graceful error handling (failed emails don't crash service)

#### Email Templates

| Event | Template | Recipient |
|-------|----------|-----------|
| Payment Completed | PaymentCompletedHTML (green) | Sender |
| Payment Received | PaymentReceivedHTML (blue) | Receiver |
| Payment Failed | PaymentFailedHTML (red) | Sender |

#### Dependencies

- **Kafka**: Message broker for events
- **SendGrid**: Email delivery
- **Environment Variables**:
  - `KAFKA_BROKER`: Kafka broker address
  - `KAFKA_TOPIC`: Payment events topic
  - `SENDGRID_API_KEY`: SendGrid API key

---

## Common Workflows

### Complete Payment Flow

1. **User registers** via POST `/register` on User Service
2. **User logs in** via POST `/login` on User Service (gets JWT)
3. **Check balance** via GET `/wallet` on User Service
4. **Send payment** via gRPC `SendPayment` on Payment Service
5. **Payment Service** creates transaction, updates balances
6. **Payment Service** publishes event to Kafka (outbox pattern)
7. **Notification Service** consumes event from Kafka
8. **Notification Service** sends email via SendGrid
9. **Payment Service** invalidates cache via POST `/internal/cache/invalidate`
10. **User Service** returns fresh balance on next `/wallet` request

### Checking Transaction History

```bash
# Get transaction details via gRPC
grpcurl -d '{"transaction_id": "txn_550e8400-e29b-41d4-a716-446655440000"}' \
  -plaintext localhost:9090 \
  payment.PaymentService/GetTransaction
```

### Error Handling

#### Insufficient Funds

```
gRPC Error: code = FailedPrecondition
desc = insufficient funds: have 50.00, need 100.00
```

**Response**: No transaction created, balances unchanged

#### Rate Limit Exceeded

**User Service** (HTTP):
```
HTTP 429 Too Many Requests
Retry-After: 60
Body: rate limit exceeded
```

**Payment Service** (gRPC):
```
rpc error: code = ResourceExhausted
desc = rate limit exceeded: too many requests
```

#### Invalid User

```
gRPC Error: code = NotFound
desc = receiver not found
```

---

## Integration Examples

### Go Client (Payment Service)

```go
package main

import (
	"context"
	"log"
	"time"

	pb "github.com/sumitDon47/payment-system/payment-service/proto"
	"google.golang.org/grpc"
)

func main() {
	// Connect to Payment Service
	conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewPaymentServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send payment
	resp, err := client.SendPayment(ctx, &pb.SendPaymentRequest{
		SenderID:   "alice-001",
		ReceiverID: "bob-001",
		Amount:     100.50,
		Currency:   "NPR",
		Note:       "Payment for lunch",
	})

	if err != nil {
		log.Fatalf("SendPayment failed: %v", err)
	}

	log.Printf("Transaction ID: %s", resp.TransactionID)
	log.Printf("Status: %s", resp.Status)
	log.Printf("New Balance: %.2f", resp.SenderBalance)
}
```

### Python Client (User Service)

```python
import requests
import json

BASE_URL = "http://localhost:8080"

# Register
register_resp = requests.post(f"{BASE_URL}/register", json={
    "name": "Alice",
    "email": "alice@example.com",
    "password": "SecurePass123!"
})
user_id = register_resp.json()["id"]

# Login
login_resp = requests.post(f"{BASE_URL}/login", json={
    "email": "alice@example.com",
    "password": "SecurePass123!"
})
token = login_resp.json()["token"]

# Get wallet
wallet_resp = requests.get(
    f"{BASE_URL}/wallet",
    headers={"Authorization": f"Bearer {token}"}
)
balance = wallet_resp.json()["balance"]
print(f"Balance: {balance}")
```

### Node.js Client (User Service)

```javascript
const axios = require('axios');

const BASE_URL = 'http://localhost:8080';

async function main() {
  try {
    // Register
    const registerRes = await axios.post(`${BASE_URL}/register`, {
      name: 'Alice',
      email: 'alice@example.com',
      password: 'SecurePass123!'
    });
    const userId = registerRes.data.id;

    // Login
    const loginRes = await axios.post(`${BASE_URL}/login`, {
      email: 'alice@example.com',
      password: 'SecurePass123!'
    });
    const token = loginRes.data.token;

    // Get wallet
    const walletRes = await axios.get(`${BASE_URL}/wallet`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    console.log(`Balance: ${walletRes.data.balance}`);
  } catch (error) {
    console.error('Error:', error.message);
  }
}

main();
```

---

## Testing & Tools

### API Testing Tools

#### Postman (User Service REST)

1. Import the OpenAPI spec (`openapi.yaml`)
2. Set base URL to `http://localhost:8080`
3. Create environment variables:
   - `token`: JWT token from login
   - `user_id`: User ID from register

#### grpcurl (Payment Service gRPC)

```bash
# List services
grpcurl -plaintext localhost:9090 list

# Describe service
grpcurl -plaintext localhost:9090 describe payment.PaymentService

# Test method
grpcurl -d '{"user_id": "alice-001"}' \
  -plaintext localhost:9090 \
  payment.PaymentService/GetBalance
```

#### Evans (gRPC CLI)

```bash
# Install: https://github.com/ktr0731/evans
evans -r -p 9090

# Interactive REPL
payment.PaymentService> call SendPayment
? sender_id: alice-001
? receiver_id: bob-001
? amount: 100.5
```

### Load Testing

#### Apache Bench (User Service)

```bash
ab -n 1000 -c 10 http://localhost:8080/health
```

#### ghz (Payment Service gRPC)

```bash
ghz --insecure \
  --proto ./proto/payment.proto \
  --call payment.PaymentService/SendPayment \
  -d '{"sender_id":"alice","receiver_id":"bob","amount":10}' \
  -n 1000 -c 10 \
  localhost:9090
```

---

## Documentation Files

| File | Purpose |
|------|---------|
| [openapi.yaml](./openapi.yaml) | OpenAPI 3.0 spec for User Service REST API |
| [GRPC_API.md](./GRPC_API.md) | gRPC API methods, validation, examples |
| [RATE_LIMITING.md](./RATE_LIMITING.md) | Rate limit configuration and strategy |
| [README.md](./README.md) | Project overview and setup |
| [EMAIL_SETUP.md](./EMAIL_SETUP.md) | Email notification configuration |

---

## Troubleshooting

### Cannot Connect to User Service

```
Connection refused on http://localhost:8080
```

**Solution**:
```bash
# Start User Service
cd user-service
go run main.go
```

### Cannot Connect to Payment Service

```
connection refused: localhost:9090
```

**Solution**:
```bash
# Start Payment Service
cd payment-service
go run main.go
```

### JWT Token Invalid

```
Error: invalid token
```

**Solution**:
1. Get new token via POST `/login`
2. Include token in `Authorization: Bearer <token>` header
3. Tokens expire in 1 hour

### Rate Limit Exceeded

**Solution**: Implement exponential backoff:

```go
import "math"
import "time"

for attempt := 0; attempt < 5; attempt++ {
    resp, err := client.SendPayment(ctx, req)
    if err == nil {
        break
    }
    if strings.Contains(err.Error(), "ResourceExhausted") {
        waitTime := time.Duration(math.Pow(2, float64(attempt))) * time.Second
        time.Sleep(waitTime)
    }
}
```

---

## Performance Characteristics

| Metric | Target | Notes |
|--------|--------|-------|
| Register | <100ms | Password hashing adds latency |
| Login | <100ms | JWT signing + DB query |
| SendPayment | <500ms | ACID transaction, event publishing |
| GetTransaction | <50ms | Simple DB query |
| GetBalance | <20ms | Redis cached (5min TTL) |

---

## Security Checklist

- ✅ Passwords hashed with bcrypt
- ✅ JWT tokens signed with HS256
- ✅ ACID transactions prevent double-spending
- ✅ Rate limiting prevents abuse
- ⚠️ TLS not enabled (plaintext connections)
- ⚠️ API keys stored in environment variables
- ⚠️ No input validation on string length (consider XSS)

**Production Recommendations**:
1. Enable TLS/SSL for all connections
2. Use environment variable encryption (e.g., HashiCorp Vault)
3. Add input validation and sanitization
4. Implement request logging and monitoring
5. Use distributed rate limiting (Redis) for multi-node setup

---

## Support & Monitoring

### Logging

Each service logs to stdout:

```
User Service:
⚠️  rate limit exceeded for IP 203.0.113.45 on endpoint /register

Payment Service:
⚠️  gRPC rate limit exceeded for user-123 on method /payment.v1.PaymentService/SendPayment
```

### Health Checks

```bash
# User Service
curl http://localhost:8080/health

# Output:
# {"status":"ok","service":"user-service","redis":"ok"}
```

### Metrics

Future enhancements:
- Prometheus metrics export
- Transaction success/failure rates
- API response time histograms
- Rate limit hit counter
