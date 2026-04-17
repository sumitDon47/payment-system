# Payment Service gRPC API Documentation

## Overview

The Payment Service is a gRPC-based microservice responsible for payment processing, transaction management, and balance tracking. It provides three core RPC methods for payment operations.

**Service**: `payment.PaymentService`  
**Port**: `9090` (default)  
**Protocol**: gRPC over HTTP/2  
**Proto File**: `payment-service/proto/payment.proto`

## Service Definition

```protobuf
service PaymentService {
  rpc SendPayment    (SendPaymentRequest)    returns (SendPaymentResponse);
  rpc GetTransaction (GetTransactionRequest) returns (GetTransactionResponse);
  rpc GetBalance     (GetBalanceRequest)     returns (GetBalanceResponse);
}
```

## Methods

### 1. SendPayment

Send money from one user to another. This is the core payment processing operation.

#### Request

```protobuf
message SendPaymentRequest {
  string sender_id   = 1;      // UUID of sender (required)
  string receiver_id = 2;      // UUID of receiver (required)
  double amount      = 3;      // Amount in decimal (required)
  string currency    = 4;      // Currency code (optional, defaults to "NPR")
  string note        = 5;      // Transaction memo (optional)
}
```

#### Response

```protobuf
message SendPaymentResponse {
  string transaction_id = 1;   // Unique transaction ID
  string status         = 2;   // "completed" | "failed" | "pending"
  double sender_balance = 3;   // Sender's new balance after transfer
  string message        = 4;   // Human-readable status message
  string created_at     = 5;   // ISO 8601 timestamp
}
```

#### Rate Limiting

- **Limit**: 100 requests/second per user
- **Burst**: 5 requests allowed immediately
- **Error Code**: `RESOURCE_EXHAUSTED`

#### Validation Rules

| Rule | Condition | Error Message |
|------|-----------|---|
| Required Fields | sender_id, receiver_id missing | "sender_id and receiver_id are required" |
| Self Payment | sender_id == receiver_id | "cannot send payment to yourself" |
| Positive Amount | amount <= 0 | "amount must be greater than zero" |
| Amount Limit | amount > 1,000,000 | "amount exceeds single-transaction limit" |
| Receiver Exists | receiver_id not in DB | "receiver not found" |
| Sender Exists | sender_id not in DB | "sender not found" |
| Sufficient Funds | sender_balance < amount | "insufficient funds: have X.XX, need Y.YY" |

#### Default Values

| Field | Default | Notes |
|-------|---------|-------|
| currency | "NPR" | Applied if empty |
| note | "" | Optional memo |

#### Example Request (Go)

```go
client := pb.NewPaymentServiceClient(conn)
req := &pb.SendPaymentRequest{
    SenderID:   "550e8400-e29b-41d4-a716-446655440000",
    ReceiverID: "660e8400-e29b-41d4-a716-446655440001",
    Amount:     100.50,
    Currency:   "NPR",
    Note:       "Payment for services",
}

resp, err := client.SendPayment(ctx, req)
if err != nil {
    log.Fatalf("SendPayment failed: %v", err)
}
fmt.Printf("Transaction ID: %s, Status: %s\n", resp.TransactionID, resp.Status)
```

#### Example Request (grpcurl)

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

#### Transaction Guarantees

- **ACID Compliance**: Uses PostgreSQL SERIALIZABLE isolation level
- **Atomicity**: All-or-nothing transaction execution
- **Consistency**: Double-spending prevention via row-level locking
- **Isolation**: SERIALIZABLE isolation prevents race conditions
- **Durability**: Committed to PostgreSQL before response

---

### 2. GetTransaction

Retrieve details of a previously executed transaction.

#### Request

```protobuf
message GetTransactionRequest {
  string transaction_id = 1;   // UUID of the transaction (required)
}
```

#### Response

```protobuf
message GetTransactionResponse {
  string transaction_id = 1;   // Transaction ID
  string sender_id      = 2;   // Sender's user ID
  string receiver_id    = 3;   // Receiver's user ID
  double amount         = 4;   // Amount transferred
  string currency       = 5;   // Currency used
  string status         = 6;   // "completed" | "failed" | "pending"
  string note           = 7;   // Transaction memo
  string created_at     = 8;   // ISO 8601 timestamp
}
```

#### Rate Limiting

- **Limit**: 1000 requests/minute per user (16.67 req/sec)
- **Burst**: 10 requests allowed immediately
- **Error Code**: `RESOURCE_EXHAUSTED`

#### Error Cases

| Condition | Error Code | Message |
|-----------|-----------|---------|
| Transaction not found | `NOT_FOUND` | "transaction not found" |
| Missing transaction_id | `INVALID_ARGUMENT` | "transaction_id is required" |

#### Example Request (Go)

```go
req := &pb.GetTransactionRequest{
    TransactionID: "txn_123456789",
}

resp, err := client.GetTransaction(ctx, req)
if err != nil {
    log.Fatalf("GetTransaction failed: %v", err)
}
fmt.Printf("Amount: %f %s, Status: %s\n", resp.Amount, resp.Currency, resp.Status)
```

#### Example Request (grpcurl)

```bash
grpcurl -d '{"transaction_id": "txn_123456789"}' \
  -plaintext localhost:9090 \
  payment.PaymentService/GetTransaction
```

---

### 3. GetBalance

Retrieve the current balance for a user.

#### Request

```protobuf
message GetBalanceRequest {
  string user_id = 1;   // UUID of the user (required)
}
```

#### Response

```protobuf
message GetBalanceResponse {
  string user_id = 1;   // User ID
  double balance = 2;   // Current balance
  string currency = 3;  // Currency (default: "NPR")
}
```

#### Rate Limiting

- **Limit**: 1000 requests/minute per user (16.67 req/sec)
- **Burst**: 10 requests allowed immediately
- **Error Code**: `RESOURCE_EXHAUSTED`

#### Error Cases

| Condition | Error Code | Message |
|-----------|-----------|---------|
| User not found | `NOT_FOUND` | "user not found" |
| Missing user_id | `INVALID_ARGUMENT` | "user_id is required" |

#### Example Request (Go)

```go
req := &pb.GetBalanceRequest{
    UserID: "550e8400-e29b-41d4-a716-446655440000",
}

resp, err := client.GetBalance(ctx, req)
if err != nil {
    log.Fatalf("GetBalance failed: %v", err)
}
fmt.Printf("Balance: %f %s\n", resp.Balance, resp.Currency)
```

#### Example Request (grpcurl)

```bash
grpcurl -d '{"user_id": "550e8400-e29b-41d4-a716-446655440000"}' \
  -plaintext localhost:9090 \
  payment.PaymentService/GetBalance
```

---

## Error Handling

### gRPC Status Codes

The Payment Service uses standard gRPC status codes for error conditions:

| Code | HTTP | Meaning | Example |
|------|------|---------|---------|
| OK | 200 | Success | Transaction completed |
| CANCELLED | 499 | Operation cancelled | Graceful shutdown |
| INVALID_ARGUMENT | 400 | Invalid request parameters | Missing sender_id |
| NOT_FOUND | 404 | Resource not found | User or transaction not found |
| ALREADY_EXISTS | 409 | Resource already exists | Duplicate transaction |
| PERMISSION_DENIED | 403 | Permission denied | Unauthorized access |
| RESOURCE_EXHAUSTED | 429 | Rate limit exceeded | 100 requests/sec hit |
| FAILED_PRECONDITION | 400 | Invalid state | Insufficient funds |
| INTERNAL | 500 | Internal server error | Database error |
| UNAVAILABLE | 503 | Service unavailable | Database down |

### Error Response Format

```
rpc error: code = <CODE> desc = <MESSAGE>
```

Example:
```
rpc error: code = ResourceExhausted desc = rate limit exceeded: too many requests
rpc error: code = FailedPrecondition desc = insufficient funds: have 50.00, need 100.00
```

---

## Transaction States

Transactions can be in one of three states:

```
┌──────────┐
│ PENDING  │  Initial state when first created
└────┬─────┘
     │
     ├─→ ┌──────────┐
     │   │COMPLETED │  Successful transfer (both balance updates done)
     │   └──────────┘
     │
     └─→ ┌────────┐
         │ FAILED │   Error during processing (rolled back)
         └────────┘
```

---

## Connection & Authentication

### Unary Interceptors

The gRPC server uses unary interceptors for cross-cutting concerns:

1. **Logging Interceptor**: Logs all RPC calls with timestamp, method name, and duration
2. **Rate Limiting Interceptor**: Enforces per-user rate limits using token bucket algorithm

### Context Propagation

User ID extraction for rate limiting (in priority order):

1. **Context Metadata** `user_id` key (set by auth middleware)
2. **Method Name** (fallback for non-authenticated requests)

Example with authenticated context:

```go
ctx = context.WithValue(baseCtx, "user_id", "user-123")
// Rate limiter will use "user-123" for limiting
```

### gRPC Reflection

gRPC reflection is enabled on the server, allowing tools like `grpcurl` to discover and test the API without proto files:

```bash
grpcurl -plaintext localhost:9090 list
# Output:
# grpc.reflection.v1.ServerReflection
# payment.PaymentService
```

---

## Testing with grpcurl

### Install grpcurl

```bash
# macOS
brew install grpcurl

# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Windows (from Go)
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### List Services

```bash
grpcurl -plaintext localhost:9090 list payment.PaymentService
```

### Test SendPayment

```bash
grpcurl -d '{
  "sender_id": "alice-001",
  "receiver_id": "bob-001",
  "amount": 100,
  "currency": "NPR"
}' \
  -plaintext localhost:9090 \
  payment.PaymentService/SendPayment
```

### Test GetBalance

```bash
grpcurl -d '{"user_id": "alice-001"}' \
  -plaintext localhost:9090 \
  payment.PaymentService/GetBalance
```

---

## Load Testing

### Example: Generate 100 Payments per Second

```bash
# Using grpcurl with a loop (NOT recommended for production)
for i in {1..100}; do
  grpcurl -d '{
    "sender_id": "alice-001",
    "receiver_id": "bob-001",
    "amount": 1
  }' -plaintext localhost:9090 payment.PaymentService/SendPayment &
done
wait
```

### Recommended Load Testing Tools

- **ghz** (gRPC load testing): https://ghz.sh/
- **grpc-load-test**: Custom Go client with concurrent goroutines
- **k6**: Kubernetes load testing with gRPC support

---

## Best Practices

### 1. Always Validate User IDs

Ensure both sender and receiver are legitimate UUIDs:

```go
senderID := "550e8400-e29b-41d4-a716-446655440000"
receiverID := "660e8400-e29b-41d4-a716-446655440001"
// Validate format before calling SendPayment
```

### 2. Handle Rate Limits Gracefully

Implement exponential backoff when hitting rate limits:

```go
var resp *pb.SendPaymentResponse
var err error
for attempt := 0; attempt < 5; attempt++ {
    resp, err = client.SendPayment(ctx, req)
    if err == nil {
        break
    }
    if strings.Contains(err.Error(), "ResourceExhausted") {
        waitTime := time.Duration(math.Pow(2, float64(attempt))) * time.Second
        time.Sleep(waitTime)
    }
}
```

### 3. Timeout Context

Always use a timeout context to prevent hanging requests:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.SendPayment(ctx, req)
```

### 4. Monitor Rate Limits

Track rate limit errors and adjust request rates accordingly:

```go
if strings.Contains(err.Error(), "rate limit exceeded") {
    log.Println("Hit rate limit, backing off...")
    // Implement backoff strategy
}
```

### 5. Verify Transaction Status

Check transaction status in response, don't assume success:

```go
resp, err := client.SendPayment(ctx, req)
if err != nil {
    // Network/RPC error
    return err
}
if resp.Status != "completed" {
    // Business logic error (insufficient funds, etc.)
    return fmt.Errorf("transaction failed: %s", resp.Message)
}
```

---

## Security Considerations

### GRPC Security

Currently, the gRPC service runs on port 9090 with:
- ✅ Unary interceptors for logging and rate limiting
- ⚠️ No TLS encryption (plaintext)
- ⚠️ No authentication interceptor (use JWT from context)

### Production Recommendations

1. **Enable TLS**: Use `grpc.Creds()` with certificate credentials
2. **Add Auth Interceptor**: Validate JWT tokens from context
3. **Network Isolation**: Run Payment Service on internal network only
4. **Rate Limiting**: Already implemented per-user limits
5. **Audit Logging**: Log all payment transactions with user info

---

## Troubleshooting

### Connection Refused

```
rpc error: code = Unavailable desc = connection refused
```

**Solution**: Verify Payment Service is running:

```bash
ps aux | grep payment-service
# Or check if port 9090 is listening:
netstat -tuln | grep 9090
```

### Rate Limit Exceeded

```
rpc error: code = ResourceExhausted desc = rate limit exceeded: too many requests
```

**Solution**: Implement exponential backoff or reduce request frequency.

### Transaction Timeout

```
context deadline exceeded
```

**Solution**: Increase timeout context duration or check database performance.

---

## Related Documentation

- [Main API Guide](./API.md) - Complete system API overview
- [OpenAPI Spec](./openapi.yaml) - User Service REST API specification
- [Rate Limiting](./RATE_LIMITING.md) - Rate limit configuration and strategy
