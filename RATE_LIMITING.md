# Rate Limiting Configuration & Implementation

This document describes the rate limiting strategy implemented across the payment system to protect APIs from abuse and ensure fair usage.

## Overview

The system implements **two-tier rate limiting**:
- **User Service (HTTP)**: Per-IP rate limiting for HTTP endpoints
- **Payment Service (gRPC)**: Per-user rate limiting for gRPC methods

This dual approach provides:
- **Protection at the edge** (HTTP layer) for user-facing APIs
- **Protection at the service** (gRPC layer) for internal payment operations

## User Service - Per-IP HTTP Rate Limiting

### Architecture

The user service implements HTTP request rate limiting using a middleware that extracts the client IP address and tracks request quotas per IP.

**Key File**: `user-service/middleware/ratelimit.go`

### Rate Limits

| Endpoint | Rate Limit | Burst | Per |
|----------|-----------|-------|-----|
| `/register` | 5 req/min | 1 | IP |
| `/login` | 5 req/min | 1 | IP |
| `/profile` | 100 req/min | 5 | IP |
| `/wallet` | 100 req/min | 5 | IP |

### Configuration

#### Authentication Endpoints (Auth Limiter)
```go
// 5 requests per minute with burst of 1
limiter := rate.NewLimiter(rate.Limit(5.0/60), 1)
```

**Purpose**: Protect against brute-force password attacks and account enumeration

#### API Endpoints (API Limiter)
```go
// 100 requests per minute with burst of 5
limiter := rate.NewLimiter(rate.Limit(100.0/60), 5)
```

**Purpose**: Allow normal application usage while preventing traffic spikes

### IP Extraction

The middleware extracts client IP in this priority order:
1. **X-Forwarded-For header** (first IP in comma-separated list) - for proxy scenarios
2. **RemoteAddr** (socket address) - for direct connections

Example proxy headers:
```
X-Forwarded-For: 203.0.113.45, 198.51.100.178
→ Uses: 203.0.113.45 (client IP, rightmost addresses are proxies)
```

### Usage Integration

In `user-service/main.go`, routes are wrapped with middleware:
```go
router.HandleFunc("/register", middleware.LimitAuth(handler.Register)).Methods("POST")
router.HandleFunc("/login", middleware.LimitAuth(handler.Login)).Methods("POST")
router.HandleFunc("/profile", middleware.LimitApi(handler.GetProfile)).Methods("GET")
router.HandleFunc("/wallet", middleware.LimitApi(handler.GetWallet)).Methods("GET")
```

### Response Behavior

When rate limit is exceeded:
- **HTTP Status**: `429 Too Many Requests`
- **Headers**: `Retry-After: 60` (wait 60 seconds)
- **Body**: `"rate limit exceeded"`

Example response:
```
HTTP/1.1 429 Too Many Requests
Retry-After: 60
Content-Type: text/plain

rate limit exceeded
```

## Payment Service - Per-User gRPC Rate Limiting

### Architecture

The payment service implements gRPC request rate limiting using a unary server interceptor that extracts the user ID from request context and tracks quotas per user.

**Key File**: `payment-service/middleware/ratelimit.go`

### Rate Limits

| Method | Rate Limit | Burst | Per |
|--------|-----------|-------|-----|
| `SendPayment` | 100 req/sec | 5 | User |
| `GetTransaction` | 1000 req/min | 10 | User |

### Configuration

#### SendPayment RPC
```go
// 100 requests per second with burst of 5
// Allows ~360,000 payments per hour per user (sufficient for high-throughput applications)
limiter := rate.NewLimiter(100, 5)
```

**Purpose**: Prevent payment flooding while allowing legitimate batch operations

#### GetTransaction RPC
```go
// 1000 requests per minute = 16.67 req/sec with burst of 10
// Allows account holders to query transaction history
limiter := rate.NewLimiter(rate.Limit(1000.0/60), 10)
```

**Purpose**: Balance analytics queries with system stability

### User ID Extraction

The interceptor extracts user ID in this priority:
1. **Context metadata** (`user_id` key) - set by auth middleware
2. **Method name** - fallback for non-authenticated requests

Example with authenticated request:
```go
// Auth middleware sets user_id in context
ctx := context.WithValue(baseCtx, "user_id", "user-123")
// Interceptor uses "user-123" for rate limiting
```

### Response Behavior

When rate limit is exceeded:
- **gRPC Status Code**: `RESOURCE_EXHAUSTED`
- **Error Message**: `"rate limit exceeded: too many requests"`

Example response:
```
rpc error: code = ResourceExhausted desc = rate limit exceeded: too many requests
```

## Implementation Details

### Token Bucket Algorithm

Both services use Go's `golang.org/x/time/rate` package which implements the **token bucket algorithm**:

1. **Tokens accumulate** at the specified rate (e.g., 100 tokens/sec for SendPayment)
2. **Burst allows** immediate consumption of up to `burst` tokens
3. **Requests consume** 1 token; allowed if tokens available
4. **Blocked requests** wait for tokens to accumulate

Example: SendPayment with limit=100/sec, burst=5
```
Time 0ms:    Tokens = 5 (burst)     → Allow 5 rapid requests
Time 1ms:    Tokens = 5             → All consumed, block request 6
Time 10ms:   Tokens = 5 + 1         → Allow 1 more request
Time 50ms:   Tokens = 5 + 5         → Allow 5 more requests
Time 1000ms: Tokens = 5 + 100       → Full second has passed
```

### Concurrency Safety

Both implementations use `sync.RWMutex` for thread-safe access to the limiter map:
```go
type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex  // Protects limiters map
}
```

This allows safe concurrent access from multiple goroutines.

### Per-User vs Per-IP Design

| Aspect | HTTP (Per-IP) | gRPC (Per-User) |
|--------|---|---|
| **Why different** | HTTP clients share IPs (proxies, NAT, mobile networks) | gRPC clients have authenticated user context |
| **Attack prevention** | Blocks malicious IPs (network-level attacks) | Blocks abusive users (application-level abuse) |
| **Performance impact** | Low (fewer keys in map) | Medium (one limiter per active user) |

## Customization

To adjust rate limits, edit the configuration in the middleware files:

### User Service (`user-service/middleware/ratelimit.go`)

Change auth limiter limits:
```go
func NewAuthLimiter() *rate.Limiter {
    // Modify these values:
    return rate.NewLimiter(rate.Limit(5.0/60), 1)  // 5 req/min with burst 1
}
```

Change API limiter limits:
```go
func NewApiLimiter() *rate.Limiter {
    // Modify these values:
    return rate.NewLimiter(rate.Limit(100.0/60), 5)  // 100 req/min with burst 5
}
```

### Payment Service (`payment-service/middleware/ratelimit.go`)

Change SendPayment limits:
```go
limiter = rate.NewLimiter(100, 5)  // 100 req/sec with burst 5
```

Change GetTransaction limits:
```go
limiter = rate.NewLimiter(rate.Limit(1000.0/60), 10)  // 1000 req/min with burst 10
```

**Note**: No environment variables are used; limits are compile-time constants for security.

## Testing

### User Service Tests

Run the rate limiting test suite:
```bash
cd user-service
go test ./middleware -v
```

Tests verify:
- ✅ IP extraction from direct connections and proxies
- ✅ Per-IP limiter isolation
- ✅ Auth endpoint 5 req/min enforcement
- ✅ API endpoint 100 req/min enforcement
- ✅ HTTP 429 responses with Retry-After headers
- ✅ Concurrent request thread-safety

### Payment Service Tests

Run the rate limiting test suite:
```bash
cd payment-service
go test ./middleware -v
```

Tests verify:
- ✅ Per-user limiter isolation
- ✅ SendPayment 100 req/sec limit
- ✅ GetTransaction 1000 req/min limit
- ✅ gRPC ResourceExhausted error responses
- ✅ Concurrent request thread-safety

## Monitoring & Observability

### Logging

Rate limit events are logged at INFO level:
```
User Service:
⚠️  rate limit exceeded for IP 203.0.113.45 on endpoint /register

Payment Service:
⚠️  gRPC rate limit exceeded for user-123 on method /payment.v1.PaymentService/SendPayment
```

### Metrics

Future enhancements could add:
- **Counter**: Total rate-limited requests per endpoint/user
- **Gauge**: Current limiter queue depth per IP/user
- **Histogram**: Request wait times during rate limit

### Client Experience

When rate-limited:
- **User Service**: Client receives `Retry-After: 60` header and can exponentially backoff
- **Payment Service**: Client receives gRPC error `codes.ResourceExhausted` and can retry with exponential backoff

## Security Considerations

### Protection Against

1. **Brute-force attacks**: Auth endpoint limits prevent password guessing
2. **Account enumeration**: Per-IP limits prevent rapid user discovery
3. **Denial-of-service**: Rate limits prevent resource exhaustion
4. **Payment fraud**: SendPayment limits prevent rapid unauthorized transfers

### Limitations

1. **Shared IP addresses**: NAT/proxy environments may hit limits with legitimate traffic
2. **Distributed attacks**: Rate limiting per-IP only protects against single-source attacks
3. **Burst absorption**: Bursts allow temporary spikes that could bypass some protections

### Recommendations

- Monitor rate limit logs for patterns indicating attacks
- Adjust burst sizes if legitimate clients hit limits
- Consider additional protections:
  - CAPTCHA for repeated 429 responses
  - IP reputation services
  - Geographic restriction policies
  - Account behavior analytics

## Future Enhancements

Potential improvements to consider:
1. **Sliding window algorithm**: Replace token bucket for stricter enforcement
2. **Distributed rate limiting**: Use Redis for multi-node systems
3. **Dynamic limits**: Adjust rates based on system load
4. **Allowlist/denylists**: Override limits for trusted/malicious IPs/users
5. **Metrics export**: Prometheus integration for monitoring

## References

- [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate)
- [Token Bucket Algorithm](https://en.wikipedia.org/wiki/Token_bucket)
- [HTTP Status 429](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/429)
- [gRPC Status Codes](https://grpc.io/docs/guides/status-codes/)
