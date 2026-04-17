# Payment System

A production-style fintech backend built with Go, demonstrating microservices architecture, gRPC communication, and ACID-compliant payment processing.

```
Client (HTTP/REST)          Client (gRPC)
        │                         │
        ▼                         ▼
┌─────────────────┐    ┌──────────────────────┐
│   User Service  │    │   Payment Service    │
│   :8080 (HTTP)  │    │   :9090 (gRPC)       │
│                 │    │                      │
│  • Register     │    │  • SendPayment       │
│  • Login (JWT)  │    │  • GetTransaction    │
│  • Profile      │    │  • GetBalance        │
│  • Wallet       │    │                      │
└────────┬────────┘    └──────────┬───────────┘
         │                        │
         └──────────┬─────────────┘
                    ▼
          ┌──────────────────┐
          │   PostgreSQL     │
          │   :5432          │
          │                  │
          │  users table     │
          │  transactions    │
          └──────────────────┘
```

---

## Why I built this

I wanted to understand how real fintech backends work — not just write CRUD endpoints, but understand the engineering decisions that prevent money from being lost, duplicated, or corrupted.

This project taught me: why ACID transactions exist, what a race condition looks like in a payment system, how gRPC differs from REST for service-to-service communication, and how to structure a Go codebase that multiple engineers could work on.

---

## Tech stack

| Technology | Purpose | Why this choice |
|---|---|---|
| Go 1.21 | Primary language | Goroutines make concurrent payment handling efficient |
| PostgreSQL 15 | Database | ACID guarantees — essential for financial data |
| gRPC + protobuf | Service communication | Type-safe contracts, faster than JSON for internal calls |
| JWT (HS256) | Authentication | Stateless auth — no session store needed |
| Docker + Compose | Infrastructure | Reproducible environment across machines |

---

## Architecture decisions

### Why two separate services?

**User Service** handles identity — registration, login, JWT issuance. It speaks HTTP because clients (browsers, Postman, mobile apps) need HTTP.

**Payment Service** handles money movement. It speaks gRPC because it is designed for internal service-to-service calls where type safety and performance matter more than human readability.

This mirrors the architecture used by companies like Stripe and Grab: HTTP on the edge facing clients, gRPC internally between services.

### Why PostgreSQL over MongoDB?

Financial data requires ACID guarantees. Every payment involves multiple writes that must all succeed or all fail atomically:

```
BEGIN TRANSACTION
  INSERT transaction record   (audit trail)
  UPDATE sender balance -500  (debit)
  UPDATE receiver balance +500 (credit)
  UPDATE transaction status   (mark complete)
COMMIT
```

If the server crashes between any of these steps, PostgreSQL rolls back everything automatically. MongoDB's default write model does not provide this guarantee without extra configuration.

### How double-spending is prevented

The SendPayment function uses `SELECT ... FOR UPDATE` to lock the sender's row before reading their balance:

```sql
SELECT balance FROM users WHERE id = $1 FOR UPDATE
```

Without this lock, two simultaneous payments from the same account could both read the same balance, both pass the balance check, and both deduct — resulting in a negative balance. The row lock forces them to queue.

---

## Project structure

```
payment-system/
│
├── user-service/               # HTTP service — identity and auth
│   ├── main.go                 # Server setup, route registration
│   ├── handler/user.go         # Register, login, profile, wallet handlers
│   ├── middleware/auth.go      # JWT validation middleware
│   ├── utils/jwt.go            # Token generation and validation
│   ├── models/user.go          # User struct and request/response types
│   ├── db/db.go                # PostgreSQL connection and migrations
│   └── Dockerfile
│
├── payment-service/            # gRPC service — payment processing
│   ├── main.go                 # gRPC server setup, service registration
│   ├── interceptor.go          # Logging middleware for all RPC calls
│   ├── handler/payment.go      # SendPayment (9-step ACID transaction)
│   ├── proto/payment.proto     # gRPC service contract
│   ├── proto/payment.go        # Generated interfaces and types
│   ├── models/transaction.go   # Transaction struct and status constants
│   ├── db/db.go                # PostgreSQL connection and migrations
│   ├── cmd/test_client/main.go # gRPC test client
│   └── Dockerfile
│
└── docker-compose.yml          # Orchestrates all three containers
```

---

## Getting started

### Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ (only needed if running without Docker)

### Run with Docker (recommended)

```bash
# Clone the repository
git clone https://github.com/yourusername/payment-system.git
cd payment-system

# Start all services
docker-compose up --build
```

That single command starts PostgreSQL, runs database migrations automatically, and boots both services. No manual setup required.

**Verify everything is running:**

```bash
curl http://localhost:8080/health
# {"status":"ok","service":"user-service"}
```

### Run without Docker

```bash
# Start PostgreSQL separately, then:

cd user-service
cp .env.example .env
# Edit .env with your database credentials
go mod tidy
go run main.go

# In another terminal:
cd payment-service
cp .env.example .env
go mod tidy
go run main.go
```

---

## Documentation

Comprehensive documentation for integrating, deploying, and maintaining the Payment System:

### API Documentation

| Document | Purpose |
|----------|---------|
| [API Quick Reference](./API_QUICK_REFERENCE.md) | **Start here** — Common endpoints, errors, examples, quick curl commands |
| [Complete API Guide](./API.md) | Full integration guide covering both services, workflows, client examples |
| [gRPC API Documentation](./GRPC_API.md) | Detailed gRPC methods, validation rules, error codes, testing with grpcurl |
| [OpenAPI Specification](./openapi.yaml) | REST API specification — import into Postman or Swagger UI |

### Deployment & Operations

| Document | Purpose |
|----------|---------|
| [CI/CD Pipeline](./CI_CD.md) | **Recommended reading** — GitHub Actions workflows, automated testing, Docker builds |
| [Deployment Guide](./DEPLOYMENT.md) | Step-by-step deployment procedures (automatic, manual, rollback, recovery) |
| [Rate Limiting](./RATE_LIMITING.md) | Rate limit configuration (5 req/min auth, 100 req/min API, 100 req/sec gRPC) |

### Configuration

| Document | Purpose |
|----------|---------|
| [Email Setup](./EMAIL_SETUP.md) | SendGrid integration for payment notifications |
| [README.md](./README.md) | This file — architecture overview and getting started |

### Example Workflows

**Complete payment flow (curl + grpcurl):**

```bash
# 1. Register
REGISTER=$(curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "email": "alice@example.com",
    "password": "SecurePass123!"
  }')
ALICE_ID=$(echo $REGISTER | jq -r '.id')

# 2. Login (get JWT)
LOGIN=$(curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"SecurePass123!"}')
TOKEN=$(echo $LOGIN | jq -r '.token')

# 3. Check wallet
curl http://localhost:8080/wallet \
  -H "Authorization: Bearer $TOKEN"

# 4. Send payment (gRPC)
grpcurl -d "{
  \"sender_id\": \"$ALICE_ID\",
  \"receiver_id\": \"bob-001\",
  \"amount\": 100.50,
  \"currency\": \"NPR\"
}" -plaintext localhost:9090 payment.PaymentService/SendPayment
```

---

## API reference

### User Service — HTTP (port 8080)

**Register**
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"secret123"}'
```
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Alice",
    "email": "alice@example.com",
    "balance": 0
  }
}
```

**Login**
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123"}'
```

**Get profile** *(requires JWT)*
```bash
curl http://localhost:8080/profile \
  -H "Authorization: Bearer <token>"
```

**Get wallet balance** *(requires JWT)*
```bash
curl http://localhost:8080/wallet \
  -H "Authorization: Bearer <token>"
```

---

### Payment Service — gRPC (port 9090)

Install grpcurl to test without writing code:
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

**Send a payment**
```bash
grpcurl -plaintext \
  -d '{
    "sender_id":   "uuid-of-sender",
    "receiver_id": "uuid-of-receiver",
    "amount":      500.00,
    "currency":    "NPR",
    "note":        "for coffee"
  }' \
  localhost:9090 payment.PaymentService/SendPayment
```
```json
{
  "transaction_id": "txn-uuid",
  "status": "completed",
  "sender_balance": 4500.00,
  "message": "Successfully sent 500.00 NPR",
  "created_at": "2026-04-11T14:22:01Z"
}
```

**Get transaction**
```bash
grpcurl -plaintext \
  -d '{"transaction_id":"txn-uuid"}' \
  localhost:9090 payment.PaymentService/GetTransaction
```

**Get balance**
```bash
grpcurl -plaintext \
  -d '{"user_id":"your-uuid"}' \
  localhost:9090 payment.PaymentService/GetBalance
```

---

## Key engineering concepts demonstrated

**ACID transactions** — SendPayment wraps all database writes in a serializable transaction. Either all writes commit together or none do. Prevents partial payment states.

**Row-level locking** — `SELECT FOR UPDATE` prevents race conditions when two payments from the same account arrive simultaneously.

**Fail-fast startup** — Both services call `db.Ping()` at startup. If the database is unreachable, the service exits immediately with a clear error rather than starting and failing silently on every request.

**gRPC interceptor pattern** — A single logging interceptor wraps all RPC methods, recording method name, duration, and success/failure. Applied once in `main.go`, covers all current and future methods automatically.

**Deferred rollback** — A `defer` at the start of every transaction automatically rolls back if any step sets an error, without requiring explicit rollback calls at every possible failure point.

**Separation of concerns** — Each package has one job: `models` holds data shapes, `db` handles connection, `handler` contains business logic, `proto` defines the service contract. No package crosses into another's responsibility.

---

## What I learned building this

**Before this project** I could write basic REST APIs. I understood HTTP. I could connect to a database.

**After this project** I understand why financial systems use ACID databases, what a race condition looks like at the SQL level and how to prevent it, how gRPC interfaces enforce contracts at compile time, why `defer` with rollback is safer than explicit rollback calls, how Docker Compose orchestrates dependent services with health checks, and how to structure a Go codebase that separates concerns cleanly.

The hardest concept was understanding that `tx.Commit()` is the only moment money actually moves — all the preceding UPDATE statements are written to a private temporary workspace inside the transaction that no other connection can see until commit.

---

## Roadmap

- [ ] Phase 3 — Notification Service with Kafka event streaming
- [ ] Redis for wallet balance caching and rate limiting
- [ ] Unit and integration tests
- [ ] Kubernetes deployment manifests
- [ ] Prometheus metrics and Grafana dashboards
- [ ] Graceful shutdown handling

---

## Running the test client

```bash
cd payment-service/cmd/test_client
go run main.go
# Tests GetBalance, SendPayment, and GetTransaction against your running service
```

---

## Outbox admin helper (dead-event replay)

Use this helper when outbox events are moved to `dead` after max retries.

```bash
cd payment-service

# List latest dead events
go run ./cmd/outbox_admin list-dead --limit 10

# Replay one dead event by outbox ID
go run ./cmd/outbox_admin replay --id <outbox-event-id>

# Replay all dead events
go run ./cmd/outbox_admin replay --all
```

Replay sets event status back to `pending` and resets retry metadata so the dispatcher can publish again.

---

## License

MIT