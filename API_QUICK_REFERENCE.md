# API Quick Reference

## HTTP REST API (User Service)

```bash
# Health Check
GET /health

# Register New User
POST /register
Content-Type: application/json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePass123!"
}

# Login (Get JWT Token)
POST /login
Content-Type: application/json
{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
# Returns: { "token": "...", "expires_in": 3600, "user_id": "..." }

# Get User Profile (Authenticated)
GET /profile
Authorization: Bearer <token>

# Get Wallet Balance (Authenticated)
GET /wallet
Authorization: Bearer <token>

# Invalidate User Cache (Internal)
POST /internal/cache/invalidate?user_id=<user_id>
```

## gRPC API (Payment Service)

```bash
# Send Payment
grpcurl -d '{
  "sender_id": "...",
  "receiver_id": "...",
  "amount": 100.50,
  "currency": "NPR",
  "note": "Optional memo"
}' -plaintext localhost:9090 payment.PaymentService/SendPayment

# Get Transaction Details
grpcurl -d '{"transaction_id": "..."}' -plaintext localhost:9090 payment.PaymentService/GetTransaction

# Get User Balance
grpcurl -d '{"user_id": "..."}' -plaintext localhost:9090 payment.PaymentService/GetBalance
```

## Rate Limits

| Service | Endpoint | Limit | Burst |
|---------|----------|-------|-------|
| User Service | /register, /login | 5 req/min | 1 |
| User Service | /profile, /wallet | 100 req/min | 5 |
| Payment Service | SendPayment | 100 req/sec | 5 |
| Payment Service | GetTransaction, GetBalance | 1000 req/min | 10 |

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 404 | Not Found |
| 409 | Conflict |
| 429 | Rate Limited |
| 500 | Server Error |

## gRPC Status Codes

| Code | HTTP | Meaning |
|------|------|---------|
| OK | 200 | Success |
| INVALID_ARGUMENT | 400 | Missing/invalid field |
| NOT_FOUND | 404 | Resource not found |
| ALREADY_EXISTS | 409 | Duplicate |
| PERMISSION_DENIED | 403 | Unauthorized |
| RESOURCE_EXHAUSTED | 429 | Rate limit hit |
| FAILED_PRECONDITION | 400 | Invalid state (e.g., insufficient funds) |
| INTERNAL | 500 | Server error |

## Common Errors

### User Service

**Invalid Email Format**
```json
{"error": "invalid email"}
```

**Password Too Short**
```json
{"error": "password must be at least 8 characters"}
```

**Email Already Exists**
```json
{"error": "email already registered"}
```

**Invalid Credentials**
```json
{"error": "invalid email or password"}
```

**Missing JWT Token**
```json
{"error": "missing authorization header"}
```

**Rate Limited (Auth)**
```
HTTP 429 Too Many Requests
Retry-After: 60
rate limit exceeded
```

### Payment Service

**Sender Not Found**
```
rpc error: code = NotFound desc = sender not found
```

**Insufficient Funds**
```
rpc error: code = FailedPrecondition desc = insufficient funds: have 50.00, need 100.00
```

**Invalid Amount**
```
rpc error: code = InvalidArgument desc = amount must be greater than zero
```

**Rate Limited**
```
rpc error: code = ResourceExhausted desc = rate limit exceeded: too many requests
```

## Example Workflows

### Complete Payment Flow

```bash
# 1. Register user
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "email": "alice@example.com",
    "password": "Pass123456"
  }'
# Save: user_id, response includes it

# 2. Login to get token
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "Pass123456"
  }'
# Save: token

# 3. Check balance
curl http://localhost:8080/wallet \
  -H "Authorization: Bearer <token>"
# Shows: balance, currency

# 4. Send payment (using gRPC)
grpcurl -d '{
  "sender_id": "<alice_id>",
  "receiver_id": "<bob_id>",
  "amount": 100.00,
  "currency": "NPR"
}' -plaintext localhost:9090 payment.PaymentService/SendPayment
# Shows: transaction_id, status, new balance

# 5. Check new balance
curl http://localhost:8080/wallet \
  -H "Authorization: Bearer <token>"
# Balance updated
```

## Authentication

**Get JWT Token**:
```bash
curl -X POST http://localhost:8080/login \
  -d '{"email": "user@example.com", "password": "..."}' \
  -H "Content-Type: application/json"
```

**Use JWT Token**:
```bash
curl http://localhost:8080/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

**Token Details**:
- Valid for: 1 hour
- Algorithm: HS256
- Payload: Contains user_id, exp, iat

## Pagination (Future Enhancement)

Currently not implemented. Each request returns full resource.

## Sorting & Filtering (Future Enhancement)

Currently not implemented. All queries return default order.

## Tools & Testing

### Postman
1. Import `openapi.yaml`
2. Set base URL: `http://localhost:8080`
3. Set environment variable: `token` (from login response)

### grpcurl
```bash
# List all services
grpcurl -plaintext localhost:9090 list

# Describe service
grpcurl -plaintext localhost:9090 describe payment.PaymentService.SendPayment

# Test with proto file
grpcurl -plaintext -proto proto/payment.proto \
  -d '{"sender_id":"alice","receiver_id":"bob","amount":100}' \
  localhost:9090 payment.PaymentService/SendPayment
```

### curl
```bash
# With JWT token
curl -H "Authorization: Bearer <token>" http://localhost:8080/profile

# POST with JSON
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"...","email":"...","password":"..."}'

# Verbose output
curl -v http://localhost:8080/health

# Save response to file
curl http://localhost:8080/profile > response.json
```

## Environment Setup

```bash
# User Service
PORT=8080
DATABASE_URL=postgresql://user:pass@localhost:5432/payment_system
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key-change-this

# Payment Service
GRPC_PORT=9090
DATABASE_URL=postgresql://user:pass@localhost:5432/payment_system
KAFKA_BROKER=localhost:29092
KAFKA_TOPIC=payment-events
KAFKA_DLQ_TOPIC=payment-dlq

# Notification Service
KAFKA_BROKER=localhost:29092
KAFKA_TOPIC=payment-events
SENDGRID_API_KEY=SG.xxx...
```

## Deployment

### Docker Compose

```bash
docker-compose up -d
# Starts: PostgreSQL, Redis, Kafka, all 3 services
```

### Verify Services

```bash
# User Service
curl http://localhost:8080/health

# Payment Service reflection
grpcurl -plaintext localhost:9090 list

# Kafka topics
docker exec kafka kafka-topics --list --bootstrap-server localhost:9092
```

## Monitoring

### Check Logs

```bash
# User Service
docker logs payment-system-user-service-1

# Payment Service
docker logs payment-system-payment-service-1

# Notification Service
docker logs payment-system-notification-service-1

# Kafka
docker logs payment-system-kafka-1
```

### Health Status

```bash
# All services should return status: ok
curl http://localhost:8080/health
```

## Rate Limit Handling

### Exponential Backoff (Go)

```go
for attempt := 0; attempt < 5; attempt++ {
    resp, err := client.SendPayment(ctx, req)
    if err == nil {
        return resp, nil
    }
    if strings.Contains(err.Error(), "ResourceExhausted") {
        backoff := time.Duration(math.Pow(2, float64(attempt)))
        time.Sleep(backoff * time.Second)
        continue
    }
    return nil, err
}
```

### Check Retry-After Header

```bash
curl -i http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"...","email":"...","password":"..."}'

# Look for header: Retry-After: 60
```

## Security Notes

- Always use HTTPS in production
- Store JWT_SECRET in secure vault (not hardcoded)
- Don't expose user IDs in client-facing URLs
- Implement request validation on client side
- Log all payment transactions for audit
- Use TLS for gRPC connections in production
- Rotate API keys regularly

## Common Issues

### Port Already in Use

```bash
# Find process using port
lsof -i :8080
# Kill process
kill -9 <PID>
```

### Database Connection Failed

```
Error: dial tcp localhost:5432: connection refused
```

Solution: Start PostgreSQL
```bash
docker-compose up -d postgres
```

### Cannot Reach gRPC Service

```
Error: connection refused
```

Solution: Start Payment Service
```bash
cd payment-service && go run main.go
```

### JWT Token Expired

Solution: Get new token from `/login` endpoint

### Transaction Failed with Insufficient Funds

Receiver successfully created. Check balance with GET `/wallet`.

---

## Documentation Links

- [Full API Guide](./API.md) - Complete documentation
- [gRPC API Details](./GRPC_API.md) - gRPC methods and examples
- [OpenAPI Spec](./openapi.yaml) - REST API specification
- [Rate Limiting](./RATE_LIMITING.md) - Rate limit details
- [Email Setup](./EMAIL_SETUP.md) - Email configuration
