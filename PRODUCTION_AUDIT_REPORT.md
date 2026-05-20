# 🏛️ PRODUCTION-GRADE FINTECH ARCHITECTURE AUDIT REPORT
**Date**: May 20, 2026  
**Scope**: Digital Wallet Payment System  
**Assessment**: High-priority vulnerabilities requiring immediate remediation  
**Target Standard**: eSewa / Khalti / PayPal / Stripe-grade Security

---

## EXECUTIVE SUMMARY

Your system has **foundational microservices architecture** with proper database transactions and outbox patterns, but **critical security gaps** and **scalability limitations** prevent production deployment. 

### Current State
✅ **Good**:
- SERIALIZABLE transactions prevent double-spending
- Outbox pattern with Kafka for event reliability
- Structured logging and metrics
- Basic bcrypt password hashing
- JWT authentication framework

❌ **Critical Issues** (Must Fix):
- No refresh token rotation (JWT expires in 24h with no renewal)
- Zero protection against brute-force attacks
- MPIN stored as bcrypt but not rate-limited
- No idempotency key mechanism for duplicate transfer prevention
- Hardcoded secrets in environment variables
- No request encryption or HMAC signing
- Zero fraud detection system
- No WAF or DDoS protection
- Missing audit logging for compliance
- No device fingerprinting or geo-location tracking

⚠️ **Scalability Issues**:
- Redis cache limited to single-node (no cluster)
- Outbox batch processing could exceed memory with large event volume
- No connection pooling optimization
- Kafka consumer processing sequential (should be parallel)
- No circuit breaker for payment service failures
- Single PostgreSQL instance (no read replicas)

---

## 1. AUTHENTICATION & SECURITY AUDIT

### 1.1 Current Implementation Review

**JWT Configuration:**
```go
// Current: Single 24-hour token, no refresh
token, err := utils.GenerateToken(user.ID, user.Email)
// Issue: No expiry control, no refresh mechanism
```

**Password Hashing:**
```go
bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
// GOOD: bcrypt is secure (cost factor 10+)
// ISSUE: No password policy enforcement (length, complexity)
// ISSUE: No password history tracking (user could reuse compromised passwords)
```

**MPIN Implementation:**
```go
// Stored in users table as mpin_hash
// ISSUE: Same hash used for every attempt - no rate limiting
// ISSUE: No attempt tracking per IP/device
// ISSUE: OTP table has phone_number but no device_id tracking
```

### 1.2 Critical Vulnerabilities

#### **VULNERABILITY #1: No Refresh Token Rotation** ⚠️ CRITICAL
**Risk Level**: HIGH  
**Attack Vector**: Token hijacking, session fixation

**Current Problem**:
- JWT expires after 24 hours
- No refresh token endpoint
- Frontend must re-authenticate with credentials after expiry
- Stale tokens remain valid until expiry even if device is lost

**Attack Scenario**:
1. Attacker steals JWT token from compromised device
2. Token remains valid for 24 hours
3. Attacker transfers funds without any detection
4. Victim discovers attack only after initial discovery

**Production Fix Required**:
```go
// Implement 2-token system
type TokenPair struct {
    AccessToken  string    // 15-minute expiry, stateless
    RefreshToken string    // 7-day expiry, stored in DB
    ExpiresAt    int64
}

// Refresh token table schema
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    device_id VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    user_agent TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'active', -- active, revoked, expired
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP,
    rotated_at TIMESTAMP
);

// Refresh token rotation on each use
POST /auth/refresh
- Input: refresh_token, device_id
- Verify token exists + not revoked
- Generate new access token + new refresh token
- Invalidate old refresh token (mark as rotated)
- Return new token pair
```

**Why This Matters**:
- Short-lived access tokens limit damage from token theft
- Device tracking prevents token reuse on different devices
- Rotation audit trail enables detecting compromised tokens
- Revocation support for immediate token invalidation after device loss

---

#### **VULNERABILITY #2: No Brute-Force Protection** ⚠️ CRITICAL
**Risk Level**: HIGH  
**Attack Vector**: Credential stuffing, rainbow table attacks

**Current Problem**:
- No rate limiting on `/login` endpoint (only soft rate limit: 5 req/min)
- No account lockout after failed attempts
- No CAPTCHA after N failed tries
- MPIN attempts not tracked in code

**Attack Scenario**:
1. Attacker has list of 100,000 usernames/emails (from data breach)
2. Runs distributed brute-force attack from 1,000 IPs
3. Rate limiter allows 5 requests per IP per minute = 5,000 attempts/min
4. System breached in ~20 minutes
5. No alerts triggered, attacker drains wallet silently

**Production Fix Required**:
```go
// Implement distributed rate limiter with exponential backoff
type BruteForceTracker struct {
    mu sync.RWMutex
    attempts map[string]*AttemptsRecord // key: email+ip_hash
}

type AttemptsRecord struct {
    Count           int
    FirstAttemptAt  time.Time
    LastAttemptAt   time.Time
    LastBlockedUntil time.Time
    Locked          bool
    LockedUntil     time.Time
}

// Login endpoint with protection
func Login(w http.ResponseWriter, r *http.Request) {
    email := req.Email
    ipAddr := getClientIP(r)
    
    // 1. Check if account/IP is locked
    key := fmt.Sprintf("%s:%s", email, hashIP(ipAddr))
    if tracker.IsLocked(key) {
        return ErrorTooManyAttempts(w)
    }
    
    // 2. Validate credentials
    if !validateCredentials(email, password) {
        tracker.RecordFailure(key)
        
        // Exponential backoff: 2s, 4s, 8s, 16s, 32s...
        attempts := tracker.GetAttempts(key)
        if attempts >= 5 {
            backoffSeconds := math.Min(32, math.Pow(2, float64(attempts-4)))
            tracker.Lock(key, time.Duration(backoffSeconds)*time.Second)
            
            // Send alert
            notifySecurityTeam(email, ipAddr, attempts)
        }
        
        return ErrorInvalidCredentials(w)
    }
    
    // 3. Clear tracking on successful login
    tracker.ClearAttempts(key)
    return SuccessLogin(w, tokenPair)
}

// Redis-backed rate limiter for distributed systems
func RateLimitCheck(email, ip string) error {
    key := fmt.Sprintf("login:attempts:%s:%s", email, hashIP(ip))
    
    count, err := redis.Incr(key)
    if count == 1 {
        redis.Expire(key, 15*time.Minute)
    }
    
    if count > 5 {
        lockKey := fmt.Sprintf("login:locked:%s:%s", email, hashIP(ip))
        backoff := math.Min(32, math.Pow(2, float64(count-5)))
        redis.SetEX(lockKey, "1", time.Duration(backoff)*time.Second)
        return ErrTooManyAttempts
    }
    return nil
}
```

**Monitoring & Alerting**:
```go
// Track suspicious patterns
- 5+ failed attempts on same email in 15 min
- Login from 3+ countries in 1 hour (geo-velocity check)
- Successful login after repeated failed attempts
- Login from new device without OTP verification

// Alert channels:
- Email alert to user: "New login attempt from [City, Country] at [Time]"
- SMS OTP challenge if high risk
- Security team dashboard with real-time threat visualization
```

---

#### **VULNERABILITY #3: No Idempotency Key Enforcement** ⚠️ CRITICAL
**Risk Level**: CRITICAL  
**Attack Vector**: Duplicate transactions, double-spending

**Current Problem**:
```go
// Current SendPayment doesn't check for duplicate requests
// If network fails after commit but before response reaches client:
// Client retries → Another transaction is created
// User's balance reduced twice
// VIOLATION: Not idempotent
```

**Attack Scenario**:
1. User sends `500 NPR` to friend over weak network
2. Payment succeeds in database but response lost
3. Mobile app retries after 5 seconds (network reconnected)
4. Two transactions created, user loses `1000 NPR`
5. No duplicate detection mechanism in code

**Production Fix Required**:
```go
// Idempotency key pattern
type SendPaymentRequest struct {
    IdempotencyKey string  // NEW: unique request identifier
    SenderID       string
    ReceiverID     string
    Amount         float64
    Currency       string
    Note           string
}

// Database schema
CREATE TABLE idempotency_keys (
    id UUID PRIMARY KEY,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users(id),
    request_hash VARCHAR(255) NOT NULL, -- hash of full request
    response JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL -- 24h TTL
);

CREATE INDEX idx_idempotency_key ON idempotency_keys(idempotency_key, user_id);

// Implementation
func (s *Server) SendPayment(ctx context.Context, req *pb.SendPaymentRequest) (*pb.SendPaymentResponse, error) {
    if req.IdempotencyKey == "" {
        return nil, fmt.Errorf("idempotency_key is required")
    }
    
    // 1. Check if request was already processed
    var cachedResponse []byte
    err := db.DB.QueryRowContext(ctx,
        `SELECT response FROM idempotency_keys 
         WHERE idempotency_key = $1 AND user_id = $2 AND expires_at > NOW()`,
        req.IdempotencyKey, req.SenderID,
    ).Scan(&cachedResponse)
    
    if err == nil {
        // Cache hit: return previous response
        utils.Info("Idempotent request", map[string]interface{}{
            "idempotency_key": req.IdempotencyKey,
            "cached": true,
        })
        var resp pb.SendPaymentResponse
        json.Unmarshal(cachedResponse, &resp)
        return &resp, nil
    }
    
    // 2. Process new request
    response, err := s.processPayment(ctx, req)
    
    // 3. Cache response
    respBytes, _ := json.Marshal(response)
    db.DB.ExecContext(ctx,
        `INSERT INTO idempotency_keys 
         (idempotency_key, user_id, request_hash, response, expires_at) 
         VALUES ($1, $2, $3, $4::jsonb, NOW() + INTERVAL '24 hours')`,
        req.IdempotencyKey, req.SenderID, hashRequest(req), string(respBytes),
    )
    
    return response, err
}

// Client-side implementation
import "github.com/google/uuid"

func SendPaymentWithIdempotency(senderID, receiverID string, amount float64) {
    idempotencyKey := uuid.New().String()
    
    req := &pb.SendPaymentRequest{
        IdempotencyKey: idempotencyKey,
        SenderID:       senderID,
        ReceiverID:     receiverID,
        Amount:         amount,
    }
    
    // First attempt
    resp1, err := paymentClient.SendPayment(ctx, req)
    if err != nil && isNetworkError(err) {
        // Same request, same key → guaranteed same response
        resp2, _ := paymentClient.SendPayment(ctx, req)
        // resp2.TransactionID == resp1.TransactionID
    }
}
```

**Why This Matters**:
- Financial transactions MUST be idempotent (banking standard)
- Network failures are common in mobile apps
- Without idempotency, retries create duplicate transactions
- Payment reconciliation becomes impossible with duplicate transactions

---

#### **VULNERABILITY #4: No Device Fingerprinting or Session Binding** ⚠️ HIGH
**Risk Level**: HIGH  
**Attack Vector**: Session hijacking, token reuse on different devices

**Current Problem**:
```go
// Current token only contains:
// { user_id, email, iat, exp }
// No device identification
// Token stolen from Device A can be used on Device B
```

**Attack Scenario**:
1. User logs in on iPhone
2. Attacker steals JWT token via malware/MITM
3. Attacker uses token on Android phone
4. System sees valid token, allows access
5. User doesn't know their account is compromised until funds are gone

**Production Fix Required**:
```go
// Device Fingerprinting Module
type DeviceFingerprint struct {
    DeviceID      string // Hardware ID / UUID
    OSType        string // iOS, Android, Web
    OSVersion     string // e.g., "14.5"
    AppVersion    string // e.g., "1.2.3"
    BrowserAgent  string
    IPAddress     string
    DeviceModel   string
    Timezone      string
    Language      string
}

// Enhanced JWT payload
type TokenPayload struct {
    UserID        string
    Email         string
    DeviceID      string
    DeviceFP      string // hash(device attributes)
    IssuedAt      int64
    ExpiresAt     int64
    SessionID     string // unique per login session
}

// Database schema
CREATE TABLE device_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    device_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255) NOT NULL UNIQUE,
    device_fingerprint VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    user_agent TEXT NOT NULL,
    os_type VARCHAR(20) NOT NULL,
    os_version VARCHAR(50) NOT NULL,
    app_version VARCHAR(50) NOT NULL,
    is_trusted BOOLEAN DEFAULT FALSE,
    last_activity TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_device_sessions_user ON device_sessions(user_id);
CREATE INDEX idx_device_sessions_device ON device_sessions(user_id, device_id);

// Login flow with device binding
func Login(w http.ResponseWriter, r *http.Request) {
    // 1. Get device fingerprint from request headers
    fp := extractDeviceFingerprint(r)
    
    // 2. Verify credentials (unchanged)
    user := validateCredentials(req.Email, req.Password)
    
    // 3. Create session record
    sessionID := uuid.New().String()
    deviceFPHash := hashFingerprint(fp)
    
    db.DB.Exec(`
        INSERT INTO device_sessions 
        (user_id, device_id, session_id, device_fingerprint, ip_address, 
         user_agent, os_type, os_version, app_version, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW() + INTERVAL '7 days')
    `, user.ID, fp.DeviceID, sessionID, deviceFPHash, getClientIP(r),
       fp.BrowserAgent, fp.OSType, fp.OSVersion, fp.AppVersion)
    
    // 4. Generate token with device binding
    token := generateToken(user.ID, sessionID, deviceFPHash)
    
    return TokenResponse{
        AccessToken: token,
        SessionID:   sessionID,
        DeviceID:    fp.DeviceID,
    }
}

// API Middleware: Verify device on every request
func DeviceBindingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := extractToken(r)
        claims := parseToken(token)
        
        // Get current device fingerprint
        currentFP := extractDeviceFingerprint(r)
        currentFPHash := hashFingerprint(currentFP)
        
        // Verify device hasn't changed
        if currentFPHash != claims.DeviceFP {
            // Device mismatch - high risk indicator
            logSecurityEvent(claims.UserID, "DEVICE_MISMATCH", currentFP)
            
            // Challenge with additional verification
            return sendMFAChallenge(w, claims.UserID)
        }
        
        // Update last activity
        db.DB.Exec(`
            UPDATE device_sessions 
            SET last_activity = NOW() 
            WHERE session_id = $1
        `, claims.SessionID)
        
        next.ServeHTTP(w, r)
    })
}
```

**Mobile App Implementation**:
```typescript
// expo-device for hardware fingerprinting
import * as Device from 'expo-device';
import * as SecureStore from 'expo-secure-store';

async function getDeviceFingerprint() {
    const stored = await SecureStore.getItemAsync('device_id');
    if (!stored) {
        // First-time setup
        const newID = uuid.v4();
        await SecureStore.setItemAsync('device_id', newID);
    }
    
    return {
        deviceID: stored || uuid.v4(),
        osType: Device.platformVersion ? 'iOS' : 'Android',
        osVersion: Device.platformVersion,
        appVersion: '1.0.0',
        deviceModel: Device.modelName,
    };
}

// Send with every request
const axiosInstance = axios.create({
    baseURL: API_URL,
});

axiosInstance.interceptors.request.use(async (config) => {
    const fp = await getDeviceFingerprint();
    config.headers['X-Device-ID'] = fp.deviceID;
    config.headers['X-Device-FP'] = hashFingerprint(fp);
    return config;
});
```

---

#### **VULNERABILITY #5: No Input Validation/Sanitization** ⚠️ HIGH
**Risk Level**: HIGH  
**Attack Vector**: SQL injection, XSS, NoSQL injection

**Current Problem**:
```go
// Current: Minimal validation
if req.Email == "" {
    // ...
}

// NO checks for:
// - SQL injection in string fields
// - Email format validation (RFC 5321)
// - Name XSS payloads (e.g., "<script>alert('xss')</script>")
// - Phone number format validation
// - Request size limits
// - Unicode normalization attacks
```

**Attack Scenarios**:
1. **SQL Injection**:
   ```
   email: admin'--
   password: anything
   // Could bypass authentication
   ```

2. **XSS Attack**:
   ```
   name: <img src=x onerror="alert('hacked')">
   // Stored in database, executed when name is displayed
   ```

3. **Phone Number Attack**:
   ```
   phone: +977 9800000000 || TRUNCATE users; --
   // Depending on implementation
   ```

**Production Fix Required**:
```go
// Input validation middleware
type ValidationRules struct {
    Email struct {
        Required bool
        Format   string // regex pattern
        MinLen   int
        MaxLen   int
    }
    Password struct {
        Required bool
        MinLen   int
        MaxLen   int
        Rules    []PasswordRule // complexity rules
    }
    Name struct {
        Required bool
        MinLen   int
        MaxLen   int
        AllowedCharacters string
    }
    Phone struct {
        Required bool
        Format   string // E.164 format
    }
}

// Comprehensive validation library
package validation

import (
    "regexp"
    "unicode"
    "github.com/asaskevich/govalidator"
)

// Validate email
func ValidateEmail(email string) error {
    if len(email) > 254 { // RFC 5321
        return fmt.Errorf("email too long")
    }
    if !govalidator.IsEmail(email) {
        return fmt.Errorf("invalid email format")
    }
    // Check for disposable email domains (optional)
    return nil
}

// Validate password
func ValidatePassword(password string) error {
    if len(password) < 12 {
        return fmt.Errorf("password must be at least 12 characters")
    }
    if len(password) > 128 {
        return fmt.Errorf("password too long")
    }
    
    // Check complexity
    hasUpper := false
    hasLower := false
    hasDigit := false
    hasSpecial := false
    
    for _, r := range password {
        switch {
        case unicode.IsUpper(r):
            hasUpper = true
        case unicode.IsLower(r):
            hasLower = true
        case unicode.IsDigit(r):
            hasDigit = true
        case unicode.IsPunct(r) || unicode.IsSymbol(r):
            hasSpecial = true
        }
    }
    
    if !(hasUpper && hasLower && hasDigit && hasSpecial) {
        return fmt.Errorf(
            "password must contain: uppercase, lowercase, digit, special char",
        )
    }
    
    // Check for common passwords
    if isCommonPassword(password) {
        return fmt.Errorf("password too common (appears in breach database)")
    }
    
    return nil
}

// Validate phone (E.164 format)
func ValidatePhoneNumber(phone string) error {
    // E.164: +[country code][number] (up to 15 digits)
    pattern := `^\+[1-9]\d{1,14}$`
    matched, err := regexp.MatchString(pattern, phone)
    if err != nil || !matched {
        return fmt.Errorf("phone must be E.164 format: +1234567890")
    }
    return nil
}

// Sanitize name and string fields
func SanitizeInput(input string, maxLen int) (string, error) {
    if len(input) > maxLen {
        return "", fmt.Errorf("input exceeds max length of %d", maxLen)
    }
    
    // Trim whitespace
    sanitized := strings.TrimSpace(input)
    
    // Remove control characters
    sanitized = strings.Map(func(r rune) rune {
        if unicode.IsControl(r) {
            return -1
        }
        return r
    }, sanitized)
    
    // Check for XSS patterns
    if containsXSSPattern(sanitized) {
        return "", fmt.Errorf("invalid characters detected")
    }
    
    return sanitized, nil
}

// Centralized validation in handler
func Register(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // Validate all inputs
    if errs := validateRegisterRequest(req); len(errs) > 0 {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "errors": errs,
        })
        return
    }
    
    // Sanitize inputs
    name, _ := SanitizeInput(req.Name, 100)
    email, _ := SanitizeInput(req.Email, 254)
    
    // Proceed with sanitized inputs
    // ...
}
```

---

### 1.3 Production-Grade Authentication Architecture

```
┌─────────────┐
│  Mobile App │
│  (iOS/And)  │
└──────┬──────┘
       │ 1. REGISTER: name, email, password, phone
       ▼
┌──────────────────────────────────────────┐
│      USER SERVICE (Port 8080, REST)      │
│  ┌────────────────────────────────────┐  │
│  │ Input Validation Layer             │  │
│  │ - Email format (RFC 5321)          │  │
│  │ - Password strength (12+ chars)    │  │
│  │ - Phone E.164 format               │  │
│  │ - XSS/SQL injection prevention     │  │
│  └────────────────────────────────────┘  │
│  ┌────────────────────────────────────┐  │
│  │ Rate Limiting Layer (Redis)        │  │
│  │ - 5 login attempts/IP/15min        │  │
│  │ - Exponential backoff              │  │
│  │ - Device-based rate limiting       │  │
│  └────────────────────────────────────┘  │
│  ┌────────────────────────────────────┐  │
│  │ Authentication Handler             │  │
│  │ - Bcrypt verify (cost=12)          │  │
│  │ - Device fingerprinting            │  │
│  │ - Session creation                 │  │
│  │ - MFA/OTP if risk detected         │  │
│  └────────────────────────────────────┘  │
│  ┌────────────────────────────────────┐  │
│  │ Token Generation                   │  │
│  │ - Access Token (15min, stateless)  │  │
│  │ - Refresh Token (7d, DB stored)    │  │
│  │ - Device binding                   │  │
│  └────────────────────────────────────┘  │
└──────┬───────────────────────────────────┘
       │ 2. RESPONSE: { accessToken, refreshToken, expiresIn, sessionId }
       ▼
┌─────────────┐
│  Mobile App │
│ Store Token │
│ Securely    │
└─────────────┘

3. SUBSEQUENT REQUESTS:
┌─────────────┐
│  Mobile App │
│ + Headers:  │
│ - JWT Token │
│ - Device ID │
│ - Session ID│
└──────┬──────┘
       │
       ▼
┌──────────────────────────────────────────┐
│      API GATEWAY (Port 8000)             │
│  ┌────────────────────────────────────┐  │
│  │ Request Authentication Middleware  │  │
│  │ - Verify JWT signature & expiry    │  │
│  │ - Verify device binding            │  │
│  │ - Check session validity           │  │
│  │ - Geo-velocity checks              │  │
│  └────────────────────────────────────┘  │
└──────┬───────────────────────────────────┘
       │ Valid request
       ▼
┌──────────────────┐
│ Payment Service  │
│ Wallet Service   │
│ etc.             │
└──────────────────┘
```

---

## 2. TRANSACTION ENGINE AUDIT

### 2.1 Current Implementation Analysis

**Good Aspects**:
```go
// ✅ SERIALIZABLE isolation prevents race conditions
tx, err := db.DB.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelSerializable,
})

// ✅ Pessimistic locking prevents concurrent updates
SELECT balance FROM users WHERE id = $1 FOR UPDATE

// ✅ Outbox pattern ensures reliable event publishing
INSERT INTO outbox_events (topic, payload, status)
```

### 2.2 Critical Transaction Issues

#### **ISSUE #1: No Double-Entry Ledger** ⚠️ CRITICAL
**Risk Level**: CRITICAL  
**Problem**: Balance kept only in `users.balance` column

**Current Model**:
```
users table:
├─ id: UUID
├─ name: VARCHAR
├─ balance: NUMERIC(15,2)  ❌ Single source of truth
├─ ...

// Transaction: just updates balance
UPDATE users SET balance = balance - 500 WHERE id = sender_id
UPDATE users SET balance = balance + 500 WHERE id = receiver_id
```

**Risks**:
- Balance corruption from query errors
- No immutable history for audits
- Reconciliation impossible after DB corruption
- Bank reconciliation nightmare (no detailed ledger)

**Production-Grade Solution**:
```sql
-- Double-Entry Ledger (Accounting Standard)
CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    account_id VARCHAR(50) NOT NULL,
    amount NUMERIC(15, 4) NOT NULL, -- Can be negative (debit)
    currency VARCHAR(3) NOT NULL DEFAULT 'NPR',
    entry_type VARCHAR(20) NOT NULL, -- 'DEBIT' or 'CREDIT'
    description TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    posted_at TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, posted, reversed
    
    -- Immutable contract
    CONSTRAINT amount_non_zero CHECK (amount != 0),
    CONSTRAINT entry_type_valid CHECK (entry_type IN ('DEBIT', 'CREDIT')),
    CONSTRAINT debit_negative_credit_positive CHECK (
        (entry_type = 'DEBIT' AND amount < 0) OR
        (entry_type = 'CREDIT' AND amount > 0)
    )
);

-- Account Master (GL Accounts)
CREATE TABLE accounts (
    account_id VARCHAR(50) PRIMARY KEY,
    account_name VARCHAR(100) NOT NULL,
    account_type VARCHAR(50) NOT NULL, -- ASSET, LIABILITY, EQUITY, INCOME, EXPENSE
    normal_balance_side VARCHAR(10) NOT NULL, -- DEBIT or CREDIT
    currency VARCHAR(3) NOT NULL DEFAULT 'NPR',
    is_active BOOLEAN DEFAULT TRUE
);

-- Account Master Data (Simplified)
INSERT INTO accounts VALUES
    ('1010', 'User Wallet - Active', 'ASSET', 'DEBIT', 'NPR'),
    ('1020', 'User Wallet - Dormant', 'ASSET', 'DEBIT', 'NPR'),
    ('2010', 'Platform Fee Reserve', 'LIABILITY', 'CREDIT', 'NPR'),
    ('3010', 'Partner Payout Liability', 'LIABILITY', 'CREDIT', 'NPR');

-- User Account Ledger (Real-time Balance)
CREATE TABLE user_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id),
    account_id VARCHAR(50) NOT NULL REFERENCES accounts(account_id),
    currency VARCHAR(3) NOT NULL DEFAULT 'NPR',
    balance NUMERIC(15, 4) NOT NULL DEFAULT 0,
    last_transaction_id UUID,
    last_updated_at TIMESTAMP DEFAULT NOW()
);

-- Transaction with full ledger entries
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_type VARCHAR(50) NOT NULL, -- 'P2P_TRANSFER', 'MERCHANT_PAYMENT', 'WITHDRAWAL'
    sender_id UUID REFERENCES users(id),
    receiver_id UUID REFERENCES users(id),
    amount NUMERIC(15, 4) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'NPR',
    fee_amount NUMERIC(15, 4) DEFAULT 0,
    total_amount NUMERIC(15, 4) GENERATED ALWAYS AS (amount + fee_amount) STORED,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, completed, failed, reversed
    idempotency_key VARCHAR(255) UNIQUE,
    
    -- Traceability
    initiated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    reference_number VARCHAR(100),
    description TEXT,
    metadata JSONB DEFAULT '{}',
    
    -- Audit
    created_by_device_id VARCHAR(255),
    initiated_from_ip INET,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_txn_idempotency ON transactions(idempotency_key);
CREATE INDEX idx_txn_sender ON transactions(sender_id, created_at DESC);
CREATE INDEX idx_txn_receiver ON transactions(receiver_id, created_at DESC);

CREATE INDEX idx_ledger_user ON ledger_entries(user_id, posted_at DESC);
CREATE INDEX idx_ledger_txn ON ledger_entries(transaction_id);
CREATE INDEX idx_ledger_account ON ledger_entries(account_id, posted_at DESC);
```

**Transactional Logic with Double-Entry**:
```go
// Payment: P2P Transfer (Ledger-based)
func (s *Server) SendPayment(ctx context.Context, req *pb.SendPaymentRequest) error {
    tx, err := db.DB.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelSerializable,
    })
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()
    
    // 1. Create transaction record
    txnID := uuid.New().String()
    _, err = tx.ExecContext(ctx, `
        INSERT INTO transactions 
        (id, transaction_type, sender_id, receiver_id, amount, currency, 
         idempotency_key, status)
        VALUES ($1, 'P2P_TRANSFER', $2, $3, $4, $5, $6, 'pending')
    `, txnID, req.SenderID, req.ReceiverID, req.Amount, req.Currency, req.IdempotencyKey)
    
    // 2. Create balanced ledger entries (MUST balance)
    // Debit sender's account
    _, err = tx.ExecContext(ctx, `
        INSERT INTO ledger_entries 
        (transaction_id, user_id, account_id, amount, currency, entry_type, 
         description, status)
        VALUES ($1, $2, '1010', $3, $4, 'DEBIT', $5, 'pending')
    `, txnID, req.SenderID, -req.Amount, req.Currency, 
       fmt.Sprintf("P2P Transfer to %s", req.ReceiverID))
    
    // Credit receiver's account
    _, err = tx.ExecContext(ctx, `
        INSERT INTO ledger_entries 
        (transaction_id, user_id, account_id, amount, currency, entry_type, 
         description, status)
        VALUES ($1, $2, '1010', $3, $4, 'CREDIT', $5, 'pending')
    `, txnID, req.ReceiverID, req.Amount, req.Currency,
       fmt.Sprintf("P2P Transfer from %s", req.SenderID))
    
    // 3. Verify double-entry balances
    var debitSum, creditSum float64
    err = tx.QueryRowContext(ctx, `
        SELECT 
            COALESCE(SUM(CASE WHEN entry_type = 'DEBIT' THEN amount ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN entry_type = 'CREDIT' THEN amount ELSE 0 END), 0)
        FROM ledger_entries WHERE transaction_id = $1
    `, txnID).Scan(&debitSum, &creditSum)
    
    if debitSum+creditSum != 0 {
        return fmt.Errorf("ledger entries must balance")
    }
    
    // 4. Update user balances
    _, err = tx.ExecContext(ctx, `
        UPDATE user_accounts 
        SET balance = balance - $1, last_transaction_id = $2, last_updated_at = NOW()
        WHERE user_id = $3
    `, req.Amount, txnID, req.SenderID)
    
    _, err = tx.ExecContext(ctx, `
        UPDATE user_accounts 
        SET balance = balance + $1, last_transaction_id = $2, last_updated_at = NOW()
        WHERE user_id = $3
    `, req.Amount, txnID, req.ReceiverID)
    
    // 5. Mark ledger entries as posted
    _, err = tx.ExecContext(ctx, `
        UPDATE ledger_entries 
        SET status = 'posted', posted_at = NOW()
        WHERE transaction_id = $1
    `, txnID)
    
    // 6. Update transaction status
    _, err = tx.ExecContext(ctx, `
        UPDATE transactions 
        SET status = 'completed', completed_at = NOW()
        WHERE id = $1
    `, txnID)
    
    // 7. Create outbox event
    _, err = tx.ExecContext(ctx, `
        INSERT INTO outbox_events (topic, event_key, payload, status)
        VALUES ($1, $2, $3::jsonb, 'pending')
    `, "payment.completed", txnID, eventPayload)
    
    if err = tx.Commit(); err != nil {
        return err
    }
    
    return nil
}

// Reconciliation Query (Ensures ledger integrity)
func ReconcileAccounts(ctx context.Context, userID string) error {
    // Calculate balance from ledger
    var ledgerBalance float64
    err := db.DB.QueryRowContext(ctx, `
        SELECT COALESCE(SUM(
            CASE 
                WHEN entry_type = 'DEBIT' THEN amount
                WHEN entry_type = 'CREDIT' THEN amount
                ELSE 0
            END
        ), 0)
        FROM ledger_entries
        WHERE user_id = $1 AND status = 'posted'
    `, userID).Scan(&ledgerBalance)
    
    // Get stored balance
    var storedBalance float64
    err = db.DB.QueryRowContext(ctx, `
        SELECT balance FROM user_accounts WHERE user_id = $1
    `, userID).Scan(&storedBalance)
    
    // Verify match
    if math.Abs(ledgerBalance-storedBalance) > 0.01 { // Allow 0.01 NPR tolerance
        return fmt.Errorf(
            "reconciliation failed for user %s: ledger=%.2f vs stored=%.2f",
            userID, ledgerBalance, storedBalance,
        )
    }
    
    return nil
}
```

---

#### **ISSUE #2: No Transaction Reversal/Chargeback Support** ⚠️ HIGH
**Current Problem**: No way to reverse transactions after completion

**Production Solution**:
```go
// Transaction Reversal (Chargebacks, Disputes)
func ReverseTransaction(ctx context.Context, txnID string, reason string) error {
    tx, _ := db.DB.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelSerializable,
    })
    
    // 1. Fetch original transaction
    var origTxn Transaction
    db.DB.QueryRowContext(ctx, `
        SELECT id, sender_id, receiver_id, amount, currency, status
        FROM transactions WHERE id = $1
    `, txnID).Scan(&origTxn.ID, &origTxn.SenderID, 
       &origTxn.ReceiverID, &origTxn.Amount, &origTxn.Currency, &origTxn.Status)
    
    // 2. Create reverse transaction
    reverseID := uuid.New().String()
    tx.ExecContext(ctx, `
        INSERT INTO transactions 
        (id, transaction_type, sender_id, receiver_id, amount, currency, 
         status, reference_number)
        VALUES ($1, 'REVERSAL', $2, $3, $4, $5, 'completed', $6)
    `, reverseID, origTxn.ReceiverID, origTxn.SenderID, 
       origTxn.Amount, origTxn.Currency, 
       fmt.Sprintf("Reversal of %s: %s", txnID, reason))
    
    // 3. Create reversing ledger entries (mirror original but opposite)
    // Debit receiver (return funds)
    tx.ExecContext(ctx, `
        INSERT INTO ledger_entries 
        (transaction_id, user_id, account_id, amount, currency, 
         entry_type, description, status)
        VALUES ($1, $2, '1010', $3, $4, 'DEBIT', $5, 'posted')
    `, reverseID, origTxn.ReceiverID, -origTxn.Amount, origTxn.Currency, 
       fmt.Sprintf("Reversal debit: %s", reason))
    
    // Credit sender (return funds)
    tx.ExecContext(ctx, `
        INSERT INTO ledger_entries 
        (transaction_id, user_id, account_id, amount, currency, 
         entry_type, description, status)
        VALUES ($1, $2, '1010', $3, $4, 'CREDIT', $5, 'posted')
    `, reverseID, origTxn.SenderID, origTxn.Amount, origTxn.Currency,
       fmt.Sprintf("Reversal credit: %s", reason))
    
    // 4. Update original transaction status
    tx.ExecContext(ctx, `
        UPDATE transactions SET status = 'reversed' WHERE id = $1
    `, txnID)
    
    return tx.Commit()
}
```

---

### 2.3 Transaction Consistency Guarantees

```
TRANSACTION LIFECYCLE DIAGRAM

┌─────────────┐
│   INITIATED │ (Request received, idempotency key checked)
│   (Pending) │
└──────┬──────┘
       │ Validate sender balance, receiver exists
       ▼
┌──────────────────┐
│   DEBIT PENDING  │ (Sender balance locked)
│  (FOR UPDATE)    │
└──────┬───────────┘
       │ SERIALIZABLE isolation begins
       │ Sender balance = balance - amount
       ▼
┌───────────────────────┐
│ LEDGER ENTRIES POSTED │ (Immutable record created)
│ (Debit + Credit)      │
└──────┬────────────────┘
       │ Verify double-entry balances
       ▼
┌──────────────────┐
│  COMPLETED       │ (Atomic commit)
│ (Committed)      │
└──────┬───────────┘
       │ All-or-nothing guarantee
       ▼
┌──────────────────┐
│ OUTBOX EVENT     │ (Async event published)
│ (Kafka message)  │
└──────────────────┘

FAILURE HANDLING:

If ANY step fails:
├─ Validation fails → Reject immediately (no DB change)
├─ Balance insufficient → Rollback all changes
├─ Ledger insert fails → Rollback + Error
├─ Commit fails → Retry or mark as "RETRY_PENDING"
└─ Event publish fails → Outbox manager retries

GUARANTEES:
✅ Atomicity: All-or-nothing
✅ Consistency: Double-entry always balanced
✅ Isolation: SERIALIZABLE (no race conditions)
✅ Durability: Written to WAL before commit
✅ Idempotency: Same request = same response
✅ Auditability: Full ledger history
```

---

## 3. DATABASE ARCHITECTURE AUDIT

### 3.1 Current Schema Issues

**Current Users Table**:
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(150) UNIQUE,
    password VARCHAR(255),
    balance NUMERIC(15, 2),  -- ❌ Denormalized, should be in ledger
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    phone_number VARCHAR(20),
    phone_verified BOOLEAN,
    mpin_hash VARCHAR(255)    -- ❌ Should have timestamp tracking
);
```

**Issues**:
- No soft delete (deleted users referenced in transactions)
- No password history (users can reuse old passwords)
- No MPIN attempt tracking
- No account status column (active/suspended/closed)
- No KYC status tracking
- No last login timestamp

### 3.2 Production-Grade Schema

**User Management**:
```sql
-- Enhanced Users Table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Profile
    email VARCHAR(254) NOT NULL UNIQUE,
    phone_number VARCHAR(20) UNIQUE,
    full_name VARCHAR(200) NOT NULL,
    date_of_birth DATE,
    
    -- Account Status
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, suspended, closed, dormant
    account_closed_at TIMESTAMP,
    close_reason VARCHAR(255),
    
    -- KYC/AML Status
    kyc_status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, submitted, approved, rejected
    kyc_submission_date TIMESTAMP,
    kyc_approved_date TIMESTAMP,
    kyc_rejection_reason TEXT,
    kyc_level INTEGER DEFAULT 1, -- 1=basic, 2=intermediate, 3=full
    aml_risk_score NUMERIC(5, 2) DEFAULT 0,
    aml_last_checked_at TIMESTAMP,
    
    -- Security
    password_hash VARCHAR(255) NOT NULL,
    password_changed_at TIMESTAMP DEFAULT NOW(),
    password_reset_token VARCHAR(255),
    password_reset_token_expires_at TIMESTAMP,
    mpin_hash VARCHAR(255),
    mpin_set_at TIMESTAMP,
    mpin_attempts INTEGER DEFAULT 0,
    mpin_locked_until TIMESTAMP,
    
    -- 2FA/MFA
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_method VARCHAR(20), -- 'SMS', 'EMAIL', 'TOTP', 'BIOMETRIC'
    mfa_verified_at TIMESTAMP,
    
    -- Biometric
    biometric_enabled BOOLEAN DEFAULT FALSE,
    biometric_templates JSONB, -- Encrypted biometric data
    biometric_enabled_at TIMESTAMP,
    
    -- Device
    trusted_devices TEXT[], -- Array of device IDs
    
    -- Activity
    last_login_at TIMESTAMP,
    last_login_ip INET,
    last_login_device_id VARCHAR(255),
    last_login_location GEOGRAPHY,
    login_count INTEGER DEFAULT 0,
    
    -- Preferences
    preferred_language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50),
    notification_preferences JSONB DEFAULT '{"email":true,"sms":true,"push":true}',
    
    -- Compliance
    is_politically_exposed_person BOOLEAN DEFAULT FALSE,
    sanctions_list_checked_at TIMESTAMP,
    document_verification_status VARCHAR(20), -- pending, verified, rejected
    
    -- Soft Delete
    deleted_at TIMESTAMP,
    deletion_reason VARCHAR(255),
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Indexing
    CONSTRAINT email_not_empty CHECK (email != ''),
    CONSTRAINT phone_format CHECK (phone_number IS NULL OR phone_number ~ '^\+[1-9]\d{1,14}$')
);

-- Indices for performance
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_phone ON users(phone_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON users(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_kyc_status ON users(kyc_status);
CREATE INDEX idx_users_created_at ON users(created_at DESC);
CREATE INDEX idx_users_last_login ON users(last_login_at DESC);

-- Password History (Prevents reuse)
CREATE TABLE password_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password_hash VARCHAR(255) NOT NULL,
    set_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_password_check UNIQUE (user_id, password_hash)
);

CREATE INDEX idx_password_history_user ON password_history(user_id);

-- Audit Log (All user changes)
CREATE TABLE user_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL, -- 'LOGIN', 'PASSWORD_CHANGE', 'MFA_ENABLE', etc.
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    device_id VARCHAR(255),
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_user ON user_audit_log(user_id, timestamp DESC);
CREATE INDEX idx_audit_action ON user_audit_log(action, timestamp DESC);

-- Login Attempts (For security analysis)
CREATE TABLE login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    email VARCHAR(254) NOT NULL,
    status VARCHAR(20) NOT NULL, -- 'success', 'failed', 'blocked'
    failure_reason VARCHAR(100), -- 'invalid_credentials', 'account_locked', 'mfa_failed'
    ip_address INET NOT NULL,
    user_agent TEXT,
    device_id VARCHAR(255),
    country_code VARCHAR(2),
    city VARCHAR(100),
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_login_attempts_user ON login_attempts(user_id);
CREATE INDEX idx_login_attempts_email ON login_attempts(email, timestamp DESC);
CREATE INDEX idx_login_attempts_ip ON login_attempts(ip_address, timestamp DESC);

-- Suspicious Activity Log
CREATE TABLE suspicious_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL, -- 'MULTIPLE_FAILED_LOGINS', 'UNUSUAL_LOCATION', 'VELOCITY_CHECK_FAILED'
    risk_score NUMERIC(5, 2) NOT NULL, -- 0-100
    description TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    action_taken VARCHAR(50), -- 'NONE', 'MFA_CHALLENGE', 'ACCOUNT_LOCKED', 'FLAGGED_FOR_REVIEW'
    resolved_at TIMESTAMP,
    resolution_notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_suspicious_user ON suspicious_activities(user_id, created_at DESC);
CREATE INDEX idx_suspicious_unresolved ON suspicious_activities(resolved_at) WHERE resolved_at IS NULL;
```

**Transaction & Ledger Tables**:
```sql
-- Immutable Transaction Log
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_type VARCHAR(50) NOT NULL, -- P2P_TRANSFER, MERCHANT_PAYMENT, WITHDRAWAL, DEPOSIT
    sender_id UUID REFERENCES users(id),
    receiver_id UUID REFERENCES users(id),
    merchant_id UUID REFERENCES merchants(id),
    
    amount NUMERIC(15, 4) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'NPR',
    fee_amount NUMERIC(15, 4) DEFAULT 0,
    total_amount NUMERIC(15, 4) GENERATED ALWAYS AS (amount + fee_amount) STORED,
    
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, completed, failed, reversed, disputed
    reversal_of_txn_id UUID REFERENCES transactions(id),
    
    idempotency_key VARCHAR(255) UNIQUE,
    reference_number VARCHAR(100) UNIQUE,
    description TEXT,
    
    initiated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    failed_at TIMESTAMP,
    failure_reason TEXT,
    
    device_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    country_code VARCHAR(2),
    
    metadata JSONB DEFAULT '{}', -- Custom fields per transaction type
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indices
CREATE INDEX idx_txn_sender ON transactions(sender_id, created_at DESC) WHERE status != 'failed';
CREATE INDEX idx_txn_receiver ON transactions(receiver_id, created_at DESC);
CREATE INDEX idx_txn_status ON transactions(status, created_at DESC);
CREATE INDEX idx_txn_idempotency ON transactions(idempotency_key);
CREATE INDEX idx_txn_reference ON transactions(reference_number);
CREATE INDEX idx_txn_merchant ON transactions(merchant_id, created_at DESC);

-- Double-Entry Ledger
CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    user_id UUID REFERENCES users(id),
    account_id VARCHAR(50) NOT NULL,
    
    amount NUMERIC(15, 4) NOT NULL CHECK (amount != 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'NPR',
    entry_type VARCHAR(20) NOT NULL CHECK (entry_type IN ('DEBIT', 'CREDIT')),
    
    description TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, posted, reversed
    
    posted_at TIMESTAMP,
    reversed_at TIMESTAMP,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_user ON ledger_entries(user_id, posted_at DESC);
CREATE INDEX idx_ledger_txn ON ledger_entries(transaction_id);
CREATE INDEX idx_ledger_account ON ledger_entries(account_id, posted_at DESC);

-- Dispute/Chargeback Management
CREATE TABLE disputes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    user_id UUID NOT NULL REFERENCES users(id),
    
    dispute_type VARCHAR(50) NOT NULL, -- 'UNAUTHORIZED', 'NOT_RECEIVED', 'DUPLICATE_CHARGE', 'FRAUD'
    status VARCHAR(20) NOT NULL DEFAULT 'open', -- open, investigating, resolved, closed
    resolution VARCHAR(50), -- 'merchant_refund', 'system_reversal', 'rejected'
    
    reason TEXT NOT NULL,
    evidence JSONB DEFAULT '{}', -- Supporting documents/screenshots
    
    opened_at TIMESTAMP NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMP,
    resolution_notes TEXT,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_disputes_user ON disputes(user_id, opened_at DESC);
CREATE INDEX idx_disputes_status ON disputes(status);
```

---

## 4. MICROSERVICE ARCHITECTURE REDESIGN

### 4.1 Recommended Service Breakdown

```
MONOLITHIC ARCHITECTURE (Current)
┌─────────────────────────────────────┐
│       Backend Services              │
├─────────┬─────────────┬─────────────┤
│  User   │  Payment    │ Notification│
│ Service │  Service    │  Service    │
└─────────┴─────────────┴─────────────┘


PRODUCTION-GRADE MICROSERVICES
┌──────────────────────────────────────────────────────────────┐
│                      API GATEWAY (Port 8000)                 │
│  - Authentication/Authorization                             │
│  - Rate Limiting (Token Bucket)                             │
│  - Request Validation                                        │
│  - Response Caching                                          │
│  - Logging/Tracing                                           │
│  - Circuit Breaker                                           │
└──────────────────────────────────────────────────────────────┘
         │           │           │           │           │
         ▼           ▼           ▼           ▼           ▼
    ┌────────┐ ┌─────────┐ ┌──────────┐ ┌────────┐ ┌─────────┐
    │ Auth   │ │ Wallet  │ │Transaction│ │ KYC    │ │ Fraud   │
    │Service │ │ Service │ │  Service  │ │Service │ │Detection│
    │:8001   │ │:8002    │ │:8003      │ │:8004   │ │ :8005   │
    └────────┘ └─────────┘ └──────────┘ └────────┘ └─────────┘
         │           │           │           │           │
         └───────────┼───────────┼───────────┼───────────┘
                     │
            ┌────────┴────────┐
            ▼                 ▼
       ┌─────────────┐  ┌──────────────┐
       │PostgreSQL   │  │Redis Cache   │
       │DB Cluster   │  │(Replicated)  │
       └─────────────┘  └──────────────┘
            │
    ┌───────┴───────┐
    ▼               ▼
 ┌──────┐       ┌──────┐
 │Primary│     │Standby│
 │(RW)   │     │(RO)   │
 └──────┘       └──────┘
       │
       ▼
  ┌─────────┐
  │Kafka    │ (Event Stream)
  │Cluster  │
  └─────────┘
       │
    ┌──┴──┬──┬──┐
    ▼     ▼  ▼  ▼
  [Notification][Reporting][Audit][Analytics]
```

### 4.2 Service Definitions

```
SERVICE: Auth Service (Port 8001)
├─ Endpoints:
│  ├─ POST /auth/register
│  ├─ POST /auth/login
│  ├─ POST /auth/verify-otp
│  ├─ POST /auth/mfa/setup
│  ├─ POST /auth/mfa/verify
│  ├─ POST /auth/token/refresh
│  ├─ POST /auth/token/revoke
│  ├─ POST /auth/password/change
│  ├─ POST /auth/password/reset
│  └─ POST /auth/logout
├─ Dependencies: PostgreSQL, Redis, SMS/Email service
├─ Responsibilities:
│  ├─ User registration & email verification
│  ├─ Password management (bcrypt/Argon2)
│  ├─ JWT generation & refresh
│  ├─ MFA/OTP generation & verification
│  ├─ Device fingerprinting
│  ├─ Session management
│  └─ Biometric authentication

SERVICE: Wallet Service (Port 8002)
├─ Endpoints:
│  ├─ GET /wallet/balance
│  ├─ GET /wallet/statement
│  ├─ POST /wallet/topup
│  ├─ POST /wallet/link-bank
│  ├─ GET /wallet/bank-accounts
│  └─ POST /wallet/initiate-withdrawal
├─ Dependencies: PostgreSQL (Ledger), Redis
├─ Responsibilities:
│  ├─ Balance queries (cached)
│  ├─ Transaction history
│  ├─ Wallet topup processing
│  ├─ Bank account linking
│  └─ Withdrawal initiation

SERVICE: Transaction Service (Port 8003)
├─ Endpoints:
│  ├─ POST /transactions/send (P2P transfer)
│  ├─ POST /transactions/merchant (Merchant payment)
│  ├─ GET /transactions/:id
│  ├─ POST /transactions/:id/cancel
│  ├─ POST /transactions/:id/dispute
│  └─ GET /transactions/history
├─ Dependencies: PostgreSQL (Ledger), Kafka
├─ Responsibilities:
│  ├─ P2P transfers
│  ├─ Merchant payments
│  ├─ Transaction validation
│  ├─ Idempotency key checking
│  ├─ Double-entry ledger posting
│  ├─ Event publishing
│  └─ Dispute handling

SERVICE: KYC Service (Port 8004)
├─ Endpoints:
│  ├─ POST /kyc/submit
│  ├─ GET /kyc/status
│  ├─ POST /kyc/upload-document
│  ├─ POST /kyc/verify-manual
│  └─ GET /kyc/limits
├─ Dependencies: PostgreSQL, Document storage, ML verification
├─ Responsibilities:
│  ├─ KYC document management
│  ├─ OCR document verification
│  ├─ Identity verification (manual/automatic)
│  ├─ Transaction limit assignment
│  ├─ AML/Sanctions checks
│  └─ PEP (Politically Exposed Person) identification

SERVICE: Fraud Detection Service (Port 8005)
├─ Real-time Streaming (Kafka consumer)
├─ Responsibilities:
│  ├─ Velocity checks (N transactions in X minutes)
│  ├─ Geo-location analysis (impossible travel)
│  ├─ Device fingerprint analysis
│  ├─ Transaction amount anomalies
│  ├─ Behavioral analysis
│  └─ Risk scoring
├─ Output: Kafka topic "fraud.alerts"

SERVICE: Notification Service
├─ Kafka Consumer
├─ Channels: Email, SMS, Push notifications
├─ Events:
│  ├─ payment.completed → Send receipt
│  ├─ payment.failed → Alert user
│  ├─ kyc.approved → Welcome email
│  ├─ fraud.alert → Security warning
│  └─ security.event → Unusual activity alert
```

---

## 5. API STANDARDS & SECURITY

### 5.1 API Request/Response Format

**Standard Request Format**:
```json
{
  "request_id": "req_550e8400e29b41d4a716446655440000",
  "timestamp": "2026-05-20T10:30:00Z",
  "version": "2.0",
  "idempotency_key": "550e8400-e29b-41d4-a716-446655440000",
  "payload": {
    "receiver_id": "user_550e8400...",
    "amount": 500.00,
    "currency": "NPR"
  }
}

Headers:
- Authorization: Bearer <access_token>
- Content-Type: application/json
- X-Request-ID: req_550e8400e29b41d4a716446655440000
- X-Device-ID: device_id
- X-Device-Fingerprint: hash
- X-Session-ID: session_id
- User-Agent: <standard user agent>
- X-Client-Version: 1.2.3
```

**Standard Response Format**:
```json
{
  "success": true,
  "request_id": "req_550e8400e29b41d4a716446655440000",
  "timestamp": "2026-05-20T10:30:05Z",
  "data": {
    "transaction_id": "txn_550e8400...",
    "status": "completed",
    "new_balance": 4500.00,
    "timestamp": "2026-05-20T10:30:05Z"
  },
  "meta": {
    "rate_limit": {
      "limit": 100,
      "remaining": 87,
      "reset": 1653045605
    }
  }
}
```

**Error Response Format**:
```json
{
  "success": false,
  "request_id": "req_550e8400e29b41d4a716446655440000",
  "timestamp": "2026-05-20T10:30:05Z",
  "error": {
    "code": "INSUFFICIENT_FUNDS",
    "message": "Your wallet balance is insufficient for this transaction",
    "details": {
      "available_balance": 400.00,
      "requested_amount": 500.00,
      "shortfall": 100.00
    },
    "user_action": "Please add funds to your wallet and try again"
  },
  "trace_id": "trace_550e8400e29b41d4a716446655440000"
}
```

### 5.2 Request Signature & Hmac for Payment APIs

```go
// For high-security endpoints, implement HMAC signing

type SignedRequest struct {
    Payload   []byte
    Timestamp int64
    Signature string // HMAC-SHA256
}

func SignRequest(payload []byte, secret string) string {
    timestamp := time.Now().Unix()
    message := fmt.Sprintf("%d:%s", timestamp, base64.StdEncoding.EncodeToString(payload))
    
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(message))
    signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
    
    return signature
}

// On server side, verify:
func VerifyRequestSignature(payload []byte, timestamp int64, signature string, secret string) bool {
    // Verify timestamp is within 5 minute window (prevent replay)
    if time.Now().Unix()-timestamp > 300 {
        return false
    }
    
    message := fmt.Sprintf("%d:%s", timestamp, base64.StdEncoding.EncodeToString(payload))
    
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(message))
    expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))
    
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

---

## 6. FRAUD DETECTION FRAMEWORK

### 6.1 Real-Time Risk Scoring

```go
type FraudDetector struct {
    riskThreshold float64 // 70.0 = 70% risk triggers MFA challenge
}

type TransactionRiskAssessment struct {
    TransactionID string
    RiskScore     float64 // 0-100
    RiskFactors   []RiskFactor
    Recommendation FraudAction // ALLOW, CHALLENGE, BLOCK
}

type RiskFactor struct {
    Name    string
    Weight  float64
    Value   float64
    Reason  string
}

type FraudAction string
const (
    FraudActionAllow     FraudAction = "ALLOW"
    FraudActionChallenge FraudAction = "CHALLENGE_MFA"
    FraudActionBlock     FraudAction = "BLOCK"
    FraudActionReview    FraudAction = "MANUAL_REVIEW"
)

// Risk Scoring Algorithm
func (fd *FraudDetector) AssessTransactionRisk(
    ctx context.Context, 
    user *User, 
    txn *Transaction,
) *TransactionRiskAssessment {
    
    assessment := &TransactionRiskAssessment{
        TransactionID: txn.ID,
        RiskScore: 0,
        RiskFactors: []RiskFactor{},
    }
    
    // 1. Velocity Check (50 points max)
    velocityFactor := fd.checkTransactionVelocity(user.ID, txn.Amount)
    assessment.RiskScore += velocityFactor.Value * velocityFactor.Weight
    assessment.RiskFactors = append(assessment.RiskFactors, velocityFactor)
    
    // 2. Amount Anomaly Check (30 points max)
    amountFactor := fd.checkTransactionAmount(user.ID, txn.Amount)
    assessment.RiskScore += amountFactor.Value * amountFactor.Weight
    assessment.RiskFactors = append(assessment.RiskFactors, amountFactor)
    
    // 3. Geo-Location Check (40 points max)
    geoFactor := fd.checkGeoLocation(user.LastLoginLocation, txn.InitiatedFrom)
    assessment.RiskScore += geoFactor.Value * geoFactor.Weight
    assessment.RiskFactors = append(assessment.RiskFactors, geoFactor)
    
    // 4. Device Check (25 points max)
    deviceFactor := fd.checkDeviceBinding(user.ID, txn.DeviceID)
    assessment.RiskScore += deviceFactor.Value * deviceFactor.Weight
    assessment.RiskFactors = append(assessment.RiskFactors, deviceFactor)
    
    // 5. Time-Based Check (20 points max)
    timeFactor := fd.checkTransactionTime(user.ID)
    assessment.RiskScore += timeFactor.Value * timeFactor.Weight
    assessment.RiskFactors = append(assessment.RiskFactors, timeFactor)
    
    // 6. Receiver Risk Check (35 points max)
    receiverFactor := fd.checkReceiverRisk(txn.ReceiverID)
    assessment.RiskScore += receiverFactor.Value * receiverFactor.Weight
    assessment.RiskFactors = append(assessment.RiskFactors, receiverFactor)
    
    // Determine action
    if assessment.RiskScore >= 80 {
        assessment.Recommendation = FraudActionBlock
    } else if assessment.RiskScore >= 60 {
        assessment.Recommendation = FraudActionChallenge
    } else if assessment.RiskScore >= 40 {
        assessment.Recommendation = FraudActionReview
    } else {
        assessment.Recommendation = FraudActionAllow
    }
    
    return assessment
}

// Individual Risk Checks

// 1. VELOCITY CHECK: More than N transactions in X minutes
func (fd *FraudDetector) checkTransactionVelocity(userID string, amount float64) RiskFactor {
    // Count transactions in last 15 minutes
    var txnCount int
    var totalAmount float64
    err := db.QueryRow(`
        SELECT COUNT(*), COALESCE(SUM(amount), 0)
        FROM transactions
        WHERE sender_id = $1 AND status = 'completed' AND created_at > NOW() - INTERVAL '15 minutes'
    `, userID).Scan(&txnCount, &totalAmount)
    
    risk := RiskFactor{
        Name: "VELOCITY_CHECK",
        Weight: 0.5,
        Reason: fmt.Sprintf("%d transactions in 15 min", txnCount),
    }
    
    switch {
    case txnCount >= 10:
        risk.Value = 100 // Max points - suspicious
    case txnCount >= 5:
        risk.Value = 75
    case txnCount >= 3:
        risk.Value = 50
    case txnCount >= 1:
        risk.Value = 25
    default:
        risk.Value = 0
    }
    
    return risk
}

// 2. AMOUNT ANOMALY: Transaction significantly larger than average
func (fd *FraudDetector) checkTransactionAmount(userID string, amount float64) RiskFactor {
    // Get user's transaction statistics
    var avgAmount, maxAmount float64
    err := db.QueryRow(`
        SELECT 
            AVG(amount),
            MAX(amount)
        FROM transactions
        WHERE sender_id = $1 AND status = 'completed' AND created_at > NOW() - INTERVAL '30 days'
    `, userID).Scan(&avgAmount, &maxAmount)
    
    risk := RiskFactor{
        Name: "AMOUNT_ANOMALY",
        Weight: 0.4,
        Reason: fmt.Sprintf("Amount %.2f vs avg %.2f", amount, avgAmount),
    }
    
    if avgAmount == 0 {
        risk.Value = 0
        return risk
    }
    
    ratio := amount / avgAmount
    switch {
    case ratio > 10:
        risk.Value = 100
    case ratio > 5:
        risk.Value = 75
    case ratio > 3:
        risk.Value = 50
    case ratio > 1.5:
        risk.Value = 25
    default:
        risk.Value = 0
    }
    
    return risk
}

// 3. GEO-LOCATION CHECK: Impossible travel (too far too fast)
func (fd *FraudDetector) checkGeoLocation(lastLoc, currentLoc *geo.Location) RiskFactor {
    risk := RiskFactor{
        Name: "GEO_VELOCITY",
        Weight: 0.6,
        Reason: "Geo-location check",
    }
    
    if lastLoc == nil || currentLoc == nil {
        risk.Value = 0
        return risk
    }
    
    distance := geo.Distance(lastLoc, currentLoc) // km
    lastLoginTime := time.Now().Add(-1 * time.Minute) // Assume 1 min ago
    timeDiff := time.Since(lastLoginTime).Minutes()
    
    maxPossibleSpeed := 900 // km/h (airline speed)
    requiredSpeed := distance / (timeDiff / 60)
    
    switch {
    case requiredSpeed > maxPossibleSpeed*2:
        risk.Value = 100 // Impossible travel
    case requiredSpeed > maxPossibleSpeed:
        risk.Value = 75
    case distance > 500:
        risk.Value = 50
    case distance > 200:
        risk.Value = 25
    default:
        risk.Value = 0
    }
    
    return risk
}

// 4. DEVICE BINDING CHECK
func (fd *FraudDetector) checkDeviceBinding(userID, deviceID string) RiskFactor {
    var isTrustedDevice bool
    db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM device_sessions 
            WHERE user_id = $1 AND device_id = $2 AND is_trusted = TRUE
        )
    `, userID, deviceID).Scan(&isTrustedDevice)
    
    risk := RiskFactor{
        Name: "DEVICE_BINDING",
        Weight: 0.4,
        Reason: fmt.Sprintf("Device trusted: %v", isTrustedDevice),
    }
    
    if !isTrustedDevice {
        risk.Value = 50
    }
    
    return risk
}

// 5. TIME-BASED CHECK: Transactions outside normal hours
func (fd *FraudDetector) checkTransactionTime(userID string) RiskFactor {
    // Get user's normal transaction hours
    var avgHour float64
    db.QueryRow(`
        SELECT AVG(EXTRACT(HOUR FROM created_at))
        FROM transactions
        WHERE sender_id = $1 AND created_at > NOW() - INTERVAL '30 days'
    `, userID).Scan(&avgHour)
    
    currentHour := float64(time.Now().Hour())
    
    risk := RiskFactor{
        Name: "TIME_ANOMALY",
        Weight: 0.3,
        Reason: fmt.Sprintf("Txn at hour %v, avg %v", currentHour, avgHour),
    }
    
    hourDiff := math.Abs(currentHour - avgHour)
    switch {
    case hourDiff > 12:
        risk.Value = 40
    case hourDiff > 6:
        risk.Value = 20
    default:
        risk.Value = 0
    }
    
    return risk
}

// 6. RECEIVER RISK: Is receiver flagged/high-risk?
func (fd *FraudDetector) checkReceiverRisk(receiverID string) RiskFactor {
    var amlScore float64
    var isBlacklisted bool
    
    db.QueryRow(`
        SELECT aml_risk_score, is_blacklisted
        FROM users WHERE id = $1
    `, receiverID).Scan(&amlScore, &isBlacklisted)
    
    risk := RiskFactor{
        Name: "RECEIVER_RISK",
        Weight: 0.35,
        Reason: fmt.Sprintf("Receiver AML score: %.2f", amlScore),
    }
    
    if isBlacklisted {
        risk.Value = 100
    } else if amlScore > 75 {
        risk.Value = 80
    } else if amlScore > 50 {
        risk.Value = 50
    } else {
        risk.Value = 0
    }
    
    return risk
}
```

### 6.2 Fraud Response Actions

```go
// Based on risk assessment, take action

func (fd *FraudDetector) executeAction(
    txn *Transaction,
    assessment *TransactionRiskAssessment,
) error {
    
    switch assessment.Recommendation {
    case FraudActionAllow:
        // Low risk - proceed normally
        return nil
    
    case FraudActionChallenge:
        // Medium risk - challenge with MFA
        return fd.challengeWithMFA(txn.SenderID)
    
    case FraudActionBlock:
        // High risk - block immediately
        return fd.blockAndAlert(txn, "SUSPECTED_FRAUD", assessment.RiskScore)
    
    case FraudActionReview:
        // Manual review needed
        return fd.flagForManualReview(txn, assessment)
    }
    
    return nil
}

func (fd *FraudDetector) challengeWithMFA(userID string) error {
    // Send OTP to user
    otp := generateOTP(6)
    redis.SetEX(fmt.Sprintf("mfa_challenge:%s", userID), otp, 5*time.Minute)
    
    // Send to email/SMS
    sendMFAChallenge(userID, otp)
    
    // User must verify before transaction proceeds
    return nil
}

func (fd *FraudDetector) blockAndAlert(
    txn *Transaction,
    reason string,
    riskScore float64,
) error {
    // 1. Block transaction
    db.Exec(`
        UPDATE transactions SET status = 'blocked' WHERE id = $1
    `, txn.ID)
    
    // 2. Create security alert
    db.Exec(`
        INSERT INTO suspicious_activities 
        (user_id, activity_type, risk_score, description, action_taken)
        VALUES ($1, 'FRAUD_SUSPECTED', $2, $3, 'ACCOUNT_LOCKED')
    `, txn.SenderID, riskScore, reason)
    
    // 3. Lock account temporarily
    db.Exec(`
        UPDATE users SET status = 'suspended' WHERE id = $1
    `, txn.SenderID)
    
    // 4. Send alerts
    sendUserAlert(txn.SenderID, "Your account has been temporarily locked due to suspicious activity")
    sendSecurityTeamAlert(fmt.Sprintf(
        "FRAUD ALERT: User %s, Risk Score: %.2f, Transaction: %s",
        txn.SenderID, riskScore, txn.ID,
    ))
    
    return nil
}
```

---

## 7. COMPLIANCE & AUDIT REQUIREMENTS

### 7.1 Audit Trail

```sql
-- Comprehensive Audit Log for Regulatory Compliance
CREATE TABLE compliance_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Event Details
    event_type VARCHAR(100) NOT NULL, -- 'LOGIN', 'TRANSACTION', 'KYC_SUBMISSION', etc
    event_category VARCHAR(50) NOT NULL, -- 'AUTH', 'FINANCIAL', 'SECURITY', 'COMPLIANCE'
    
    -- User Context
    user_id UUID REFERENCES users(id),
    actor_id UUID, -- Who triggered (admin, system, user)
    actor_type VARCHAR(20) NOT NULL, -- 'USER', 'ADMIN', 'SYSTEM'
    
    -- Transaction Context
    transaction_id UUID REFERENCES transactions(id),
    related_user_id UUID, -- Other user involved
    
    -- Request Details
    request_id VARCHAR(255),
    api_endpoint VARCHAR(255),
    http_method VARCHAR(10),
    http_status_code INT,
    
    -- Data Change
    old_values JSONB,
    new_values JSONB,
    changed_fields TEXT[], -- Which fields changed
    
    -- Security Context
    ip_address INET,
    user_agent TEXT,
    device_id VARCHAR(255),
    session_id VARCHAR(255),
    
    -- Geographic
    country_code VARCHAR(2),
    city VARCHAR(100),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    
    -- Compliance Metadata
    requires_approval BOOLEAN DEFAULT FALSE,
    approved_by_admin_id UUID,
    approval_reason TEXT,
    
    -- Timestamps (Immutable)
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    retention_until TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '7 years'),
    
    -- Data Integrity
    log_hash VARCHAR(255) NOT NULL, -- SHA-256 hash for tamper detection
    previous_log_hash VARCHAR(255), -- Link to previous log (blockchain-like)
    
    CONSTRAINT log_integrity UNIQUE (log_hash)
);

-- Indices
CREATE INDEX idx_audit_user ON compliance_audit_log(user_id, timestamp DESC);
CREATE INDEX idx_audit_event ON compliance_audit_log(event_type, timestamp DESC);
CREATE INDEX idx_audit_txn ON compliance_audit_log(transaction_id);
CREATE INDEX idx_audit_timestamp ON compliance_audit_log(timestamp DESC);
CREATE INDEX idx_audit_retention ON compliance_audit_log(retention_until);
```

### 7.2 Regulatory Compliance Checklist

```
========================
KYC (Know Your Customer)
========================
✅ REQUIRED IMPLEMENTATION
- [ ] ID verification (Passport, License, National ID)
- [ ] Address verification (Proof of residence)
- [ ] Identity matching (Photo ID verification)
- [ ] Birth date verification
- [ ] Tax ID number collection
- [ ] Business registration (if applicable)
- [ ] UBO (Beneficial Owner) identification
- [ ] PEP (Politically Exposed Person) check
- [ ] Sanctions list screening
- [ ] Quarterly KYC review
- [ ] Enhanced KYC for high-risk users
- [ ] Document expiration tracking
- [ ] Re-verification on limit increase

========================
AML (Anti-Money Laundering)
========================
✅ REQUIRED IMPLEMENTATION
- [ ] Transaction monitoring (Real-time)
- [ ] Velocity checks (Unusual activity patterns)
- [ ] Geographic anomaly detection
- [ ] Cash-in/cash-out tracking
- [ ] Large transaction reporting (>5L NPR)
- [ ] Suspicious transaction reports (STRs)
- [ ] Currency conversion monitoring
- [ ] Beneficial owner tracking
- [ ] Third-party remittance monitoring
- [ ] Quarterly AML risk assessment
- [ ] AML training for employees
- [ ] AML policy documentation
- [ ] Suspicious activity investigation workflow
- [ ] Report to FIU (Financial Intelligence Unit) within 24 hours

========================
Data Protection & Privacy
========================
✅ REQUIRED IMPLEMENTATION
- [ ] Data minimization (Collect only necessary data)
- [ ] Purpose limitation (Use data only for stated purpose)
- [ ] Storage limitation (Delete after purpose served)
- [ ] Encryption at rest (AES-256)
- [ ] Encryption in transit (TLS 1.3)
- [ ] Access controls (Role-based, principle of least privilege)
- [ ] Data subject consent management
- [ ] Right to be forgotten implementation
- [ ] Data breach notification (72 hours to regulator)
- [ ] Privacy impact assessment
- [ ] Data Processing Agreement (DPA)
- [ ] Subprocessor management
- [ ] Cross-border data transfer compliance

========================
Transaction Reporting
========================
✅ REQUIRED IMPLEMENTATION
- [ ] Large Transaction Report (LTR): >5,000,000 NPR
- [ ] Suspicious Transaction Report (STR): Unusual patterns
- [ ] Currency Transaction Report (CTR): >1,000,000 NPR
- [ ] Report submission to FIU within 24 hours
- [ ] Record retention for 7 years
- [ ] Quarterly reporting compliance status
- [ ] Documentation of rationale for each report
- [ ] Audit trail of all reports submitted
- [ ] Customer notification (if applicable)

========================
Sanctions Compliance
========================
✅ REQUIRED IMPLEMENTATION
- [ ] Daily OFAC (Office of Foreign Assets Control) list check
- [ ] Hit score calculation (Match confidence)
- [ ] False positive workflow
- [ ] Transaction blocking for blacklisted entities
- [ ] Escalation to compliance team
- [ ] Investigation and documentation
- [ ] Legal review before unblocking
- [ ] Quarterly sanctions audit

========================
Record Keeping
========================
✅ REQUIRED IMPLEMENTATION
- [ ] 7-year retention for all financial records
- [ ] 5-year retention for customer identification records
- [ ] Tamper-evident logging (Blockchain-like hashing)
- [ ] Immutable transaction history
- [ ] Backup and recovery procedures
- [ ] Disaster recovery plan
- [ ] Access logs for all system activities
- [ ] Export capability for regulatory requests
- [ ] Archive strategy (Hot/Warm/Cold storage)

========================
Reporting & Disclosure
========================
✅ REQUIRED IMPLEMENTATION
- [ ] Annual compliance report
- [ ] Quarterly risk assessment update
- [ ] Monthly AML/KYC status dashboard
- [ ] Incident reporting to regulators
- [ ] Breach notification to affected users
- [ ] Regulatory examination readiness
- [ ] Internal audit schedule
- [ ] Audit findings remediation tracking
```

---

## 8. PRODUCTION-GRADE SECURITY CHECKLIST

```
╔═══════════════════════════════════════════════════════════════╗
║             FINTECH SECURITY IMPLEMENTATION                   ║
║                    Production Checklist                        ║
╚═══════════════════════════════════════════════════════════════╝

TIER 1: CRITICAL (Must have before production)
═══════════════════════════════════════════════════════════════

Authentication & Session Management:
  ✅ JWT with 15-minute expiry + 7-day refresh tokens
  ✅ Refresh token rotation on every use
  ✅ Device fingerprinting & binding
  ✅ Session invalidation on device change
  ✅ Brute-force protection (5 attempts, exponential backoff)
  ✅ Account lockout after failed attempts
  ✅ CAPTCHA after N failed attempts
  ✅ Secure password hashing (bcrypt/Argon2, cost=12+)
  ✅ Password complexity enforcement (12+ chars, mixed case, symbols)
  ✅ Password history tracking (No reuse of last 5)

API Security:
  ✅ HTTPS/TLS 1.3 for all connections
  ✅ Input validation on all endpoints
  ✅ Output encoding (XSS prevention)
  ✅ SQL injection prevention (parameterized queries)
  ✅ CSRF tokens for state-changing operations
  ✅ Rate limiting (Token bucket algorithm)
  ✅ API authentication (OAuth 2.0 / JWT)
  ✅ Request signing (HMAC-SHA256 for payment APIs)
  ✅ Idempotency keys for all transactions
  ✅ Request size limits
  ✅ Timeout enforcement
  ✅ Secure headers (HSTS, CSP, X-Content-Type-Options)

Data Security:
  ✅ Encryption at rest (AES-256)
  ✅ Encryption in transit (TLS 1.3)
  ✅ Database encryption (pgcrypto)
  ✅ Sensitive field encryption (SSN, Card data)
  ✅ Key rotation schedule (90-day rotation)
  ✅ Secret management (HashiCorp Vault)
  ✅ No hardcoded secrets in code/configs
  ✅ Audit logging of all data access
  ✅ PII (Personally Identifiable Information) masking
  ✅ Data minimization (Collect only necessary data)

Transaction Security:
  ✅ SERIALIZABLE isolation for transactions
  ✅ Pessimistic locking (FOR UPDATE)
  ✅ Double-entry ledger accounting
  ✅ Immutable transaction history
  ✅ Transaction reversal workflow
  ✅ Dispute management system
  ✅ Reconciliation procedures
  ✅ Balance integrity checks
  ✅ Transaction limit enforcement
  ✅ Velocity checks (unusual patterns)

Fraud Prevention:
  ✅ Real-time fraud detection (Risk scoring)
  ✅ Geo-location velocity checks
  ✅ Device fingerprinting
  ✅ Transaction amount anomaly detection
  ✅ Time-based anomaly detection
  ✅ Receiver risk assessment
  ✅ Account takeover detection
  ✅ Synthetic identity detection
  ✅ Automated fraud response (Block/Challenge/Allow)
  ✅ Manual review queue for edge cases

Compliance & Audit:
  ✅ Comprehensive audit logging
  ✅ Immutable audit trail (7-year retention)
  ✅ Tamper detection (Hashing audit logs)
  ✅ KYC/AML implementation
  ✅ Sanctions screening
  ✅ Large transaction reporting
  ✅ Suspicious transaction reporting
  ✅ Data protection compliance
  ✅ Regulatory documentation
  ✅ Incident response plan
  ✅ Breach notification procedures

TIER 2: HIGH (Implement in first production release)
═══════════════════════════════════════════════════════════════

MFA/2FA:
  ✅ SMS OTP (One-Time Password)
  ✅ Email OTP
  ✅ TOTP (Time-based OTP)
  ✅ Biometric authentication
  ✅ MFA for login
  ✅ MFA for sensitive operations (Withdrawal, Account changes)
  ✅ Recovery codes for MFA bypass
  ✅ MFA device trust management
  ✅ MFA enforcement policies

Advanced Security:
  ✅ API Gateway (Kong, AWS API Gateway)
  ✅ Web Application Firewall (WAF)
  ✅ DDoS protection (CloudFlare, AWS Shield)
  ✅ Intrusion Detection System (IDS)
  ✅ Intrusion Prevention System (IPS)
  ✅ VPN for employee access
  ✅ Network segmentation
  ✅ Zero-trust architecture
  ✅ Privileged Access Management (PAM)

Monitoring & Alerting:
  ✅ Real-time alerting for security events
  ✅ Centralized logging (ELK, Datadog)
  ✅ Distributed tracing (Jaeger, DataDog)
  ✅ Security Information and Event Management (SIEM)
  ✅ Anomaly detection
  ✅ Alert escalation workflow
  ✅ On-call security team
  ✅ Incident response runbooks
  ✅ Regular penetration testing

TIER 3: MEDIUM (Implement gradually)
═══════════════════════════════════════════════════════════════

Advanced Fraud Detection:
  ✅ Machine Learning models (Behavioral analysis)
  ✅ Graph analysis (Network fraud detection)
  ✅ Synthetic fraud detection
  ✅ Money laundering pattern detection
  ✅ Predictive modeling
  ✅ Anomaly scoring (Statistical)
  ✅ Ensemble models (Multiple algorithms)
  ✅ Continuous model retraining
  ✅ Model explainability (LIME, SHAP)

Performance & Optimization:
  ✅ Query optimization
  ✅ Index optimization
  ✅ Caching strategy (Redis)
  ✅ Database replication (Read replicas)
  ✅ Database sharding
  ✅ Connection pooling
  ✅ Circuit breaker pattern
  ✅ Load balancing
  ✅ Auto-scaling policies
  ✅ CDN for static assets

Infrastructure:
  ✅ Containerization (Docker)
  ✅ Orchestration (Kubernetes)
  ✅ Infrastructure as Code (Terraform)
  ✅ Blue-green deployment
  ✅ Canary deployment
  ✅ Disaster recovery plan
  ✅ Regular backups (Point-in-time recovery)
  ✅ Monitoring and observability
  ✅ Cost optimization
  ✅ Security scanning in CI/CD

TIER 4: LONG-TERM (Roadmap for 6-12 months)
═══════════════════════════════════════════════════════════════

Advanced Features:
  ✅ Blockchain-based audit trail (Optional)
  ✅ Multi-signature transactions
  ✅ Smart contract integration
  ✅ Cryptocurrency wallet support
  ✅ Cross-border payments (SWIFT, RippleNet)
  ✅ Currency exchange optimization
  ✅ Liquidity management
  ✅ Algorithmic trading (If applicable)
  ✅ AI-powered customer support
  ✅ Predictive analytics
```

---

## 9. DEPLOYMENT & INFRASTRUCTURE ARCHITECTURE

```
PRODUCTION DEPLOYMENT ARCHITECTURE

┌────────────────────────────────────────────────────────────┐
│                    CLIENT LAYER                             │
├─────────────┬──────────────┬──────────────┬────────────────┤
│ Mobile App  │ Web Browser  │ Third-party  │ Admin Dashboard│
│ (iOS/And)   │ (React)      │ API Clients  │ (Operators)    │
└─────────────┴──────────────┴──────────────┴────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              EDGE LAYER (CloudFlare / AWS)                  │
│  - DDoS Protection                                          │
│  - WAF (Web Application Firewall)                           │
│  - Bot Protection                                           │
│  - Rate Limiting (Global)                                   │
│  - CDN (Static assets)                                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              LOAD BALANCER (AWS ALB / K8s)                  │
│  - SSL/TLS Termination                                      │
│  - Round-robin load balancing                               │
│  - Health checks                                            │
│  - Auto-scaling based on metrics                            │
│  - Session stickiness (if needed)                           │
└─────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
    ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
    │ API Gateway  │  │ API Gateway  │  │ API Gateway  │
    │ (Pod 1)      │  │ (Pod 2)      │  │ (Pod 3)      │
    └──────────────┘  └──────────────┘  └──────────────┘
            │                │                 │
            ▼                ▼                 ▼
    ┌─────────────────────────────────────────────────┐
    │         KUBERNETES CLUSTER                       │
    │  ┌────────────────────────────────────────────┐  │
    │  │      SERVICES (Multiple replicas)          │  │
    │  ├──────────────┬──────────────┬──────────────┤  │
    │  │ Auth Service │ Wallet Svc   │ Transaction  │  │
    │  │   (3 pods)   │   (3 pods)   │ Svc (3 pods) │  │
    │  ├──────────────┼──────────────┼──────────────┤  │
    │  │ KYC Service  │ Fraud Det.   │ Notification │  │
    │  │  (2 pods)    │  (3 pods)    │  (2 pods)    │  │
    │  └──────────────┴──────────────┴──────────────┘  │
    │                                                   │
    │  ┌──────────────────────────────────────────────┐ │
    │  │    Ingress Controller / Service Mesh (Istio) │ │
    │  │  - Service discovery                         │ │
    │  │  - Load balancing                            │ │
    │  │  - Circuit breaking                          │ │
    │  │  - Distributed tracing                       │ │
    │  │  - Mutual TLS                                │ │
    │  └──────────────────────────────────────────────┘ │
    └─────────────────────────────────────────────────┘
            │                │                 │
            ▼                ▼                 ▼
    ┌─────────────────┬──────────────────┬───────────┐
    │ PostgreSQL      │  Redis Cluster   │   Kafka   │
    │ Cluster         │  (HA, Sentinel)  │  Cluster  │
    │ (3 nodes)       │  - Cache         │ (3 nodes) │
    │ - Primary (RW)  │  - Session Store │ - Event   │
    │ - Standby (RO)  │  - Rate Limiting │   Stream  │
    │ - DR            │  - Queue         │ - Topic   │
    │                 │  - Pub/Sub       │   Storage │
    └─────────────────┴──────────────────┴───────────┘
            │                │                 │
            ▼                ▼                 ▼
    ┌───────────────────────────────────────────────┐
    │          MONITORING & OBSERVABILITY            │
    ├──────────────┬──────────────┬────────────────┤
    │ Prometheus   │ Grafana      │ ELK Stack      │
    │ (Metrics)    │ (Dashboards) │ (Logs)         │
    ├──────────────┼──────────────┼────────────────┤
    │ Jaeger       │ DataDog      │ PagerDuty      │
    │ (Tracing)    │ (APM)        │ (Alerts)       │
    └──────────────┴──────────────┴────────────────┘
```

---

## 10. PRODUCTION-GRADE FOLDER STRUCTURE

```
payment-system/
│
├── backend/                        # Backend code
│   ├── services/
│   │   ├── auth-service/
│   │   │   ├── cmd/
│   │   │   │   └── main.go
│   │   │   ├── internal/
│   │   │   │   ├── handler/
│   │   │   │   │   ├── register.go
│   │   │   │   │   ├── login.go
│   │   │   │   │   ├── mfa.go
│   │   │   │   │   └── token.go
│   │   │   │   ├── service/
│   │   │   │   │   ├── auth_service.go
│   │   │   │   │   ├── token_service.go
│   │   │   │   │   ├── mfa_service.go
│   │   │   │   │   └── password_service.go
│   │   │   │   ├── repository/
│   │   │   │   │   ├── user_repo.go
│   │   │   │   │   ├── session_repo.go
│   │   │   │   │   └── password_history_repo.go
│   │   │   │   ├── middleware/
│   │   │   │   │   ├── auth_middleware.go
│   │   │   │   │   ├── rate_limit.go
│   │   │   │   │   └── logging.go
│   │   │   │   ├── security/
│   │   │   │   │   ├── crypto.go (Encryption/Decryption)
│   │   │   │   │   ├── password.go (Hashing/Validation)
│   │   │   │   │   ├── jwt.go (JWT generation/validation)
│   │   │   │   │   └── otp.go (OTP generation)
│   │   │   │   ├── model/
│   │   │   │   │   ├── user.go
│   │   │   │   │   ├── session.go
│   │   │   │   │   └── device.go
│   │   │   │   ├── utils/
│   │   │   │   │   ├── logger.go
│   │   │   │   │   ├── validator.go
│   │   │   │   │   └── helpers.go
│   │   │   │   └── config/
│   │   │   │       └── config.go
│   │   │   ├── pkg/
│   │   │   │   ├── db/
│   │   │   │   │   ├── postgres.go
│   │   │   │   │   └── migrations/
│   │   │   │   ├── cache/
│   │   │   │   │   └── redis.go
│   │   │   │   ├── email/
│   │   │   │   │   └── sendgrid.go
│   │   │   │   └── sms/
│   │   │   │       └── twilio.go
│   │   │   ├── tests/
│   │   │   │   ├── unit/
│   │   │   │   ├── integration/
│   │   │   │   └── e2e/
│   │   │   ├── go.mod
│   │   │   ├── go.sum
│   │   │   ├── Dockerfile
│   │   │   └── .env.example
│   │   │
│   │   ├── wallet-service/         (Similar structure)
│   │   ├── transaction-service/    (Similar structure)
│   │   ├── kyc-service/            (Similar structure)
│   │   ├── fraud-detection-service/(Similar structure)
│   │   └── notification-service/   (Similar structure)
│   │
│   ├── shared/                     # Shared libraries
│   │   ├── pb/                     # gRPC proto files
│   │   │   ├── auth.proto
│   │   │   ├── payment.proto
│   │   │   └── ...
│   │   ├── errors/
│   │   │   └── error_codes.go
│   │   ├── middleware/
│   │   │   └── common_middleware.go
│   │   ├── constants/
│   │   │   └── constants.go
│   │   └── utils/
│   │       ├── crypto.go
│   │       └── validation.go
│   │
│   ├── api-gateway/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── middleware/
│   │   │   ├── router/
│   │   │   └── proxy/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   └── docker-compose.prod.yml
│
├── frontend/                       # Frontend code
│   ├── mobile/                     # React Native (iOS/Android)
│   │   ├── src/
│   │   │   ├── screens/
│   │   │   ├── components/
│   │   │   ├── navigation/
│   │   │   ├── api/
│   │   │   ├── utils/
│   │   │   ├── store/              # Redux or Zustand
│   │   │   └── security/
│   │   ├── package.json
│   │   └── .env.example
│   │
│   └── web/                        # React web dashboard
│       ├── src/
│       ├── package.json
│       └── .env.example
│
├── infrastructure/                 # Infrastructure as Code
│   ├── kubernetes/
│   │   ├── base/
│   │   │   ├── namespace.yaml
│   │   │   ├── rbac.yaml
│   │   │   ├── network-policy.yaml
│   │   │   └── storage-class.yaml
│   │   ├── services/
│   │   │   ├── auth-service-deployment.yaml
│   │   │   ├── wallet-service-deployment.yaml
│   │   │   └── ...
│   │   ├── ingress/
│   │   │   └── ingress.yaml
│   │   ├── monitoring/
│   │   │   ├── prometheus.yaml
│   │   │   ├── grafana.yaml
│   │   │   └── ...
│   │   └── kustomization.yaml
│   │
│   ├── terraform/                  # Infrastructure as Code
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── networking.tf
│   │   ├── database.tf
│   │   ├── cache.tf
│   │   ├── kafka.tf
│   │   ├── monitoring.tf
│   │   └── terraform.tfvars.example
│   │
│   └── docker/
│       ├── Dockerfile.base       # Base image with common setup
│       └── docker-compose.prod.yml
│
├── database/
│   ├── migrations/
│   │   ├── 001_users.sql
│   │   ├── 002_transactions.sql
│   │   ├── 003_ledger.sql
│   │   ├── 004_audit_log.sql
│   │   └── ...
│   ├── schemas/
│   │   ├── users_schema.sql
│   │   ├── transactions_schema.sql
│   │   └── ...
│   ├── stored_procedures/
│   │   ├── reconcile_accounts.sql
│   │   ├── generate_reports.sql
│   │   └── ...
│   └── backups/
│       └── backup_strategy.md
│
├── scripts/
│   ├── deploy.sh                   # Deployment script
│   ├── backup.sh                   # Backup script
│   ├── restore.sh                  # Restore script
│   ├── health-check.sh
│   └── incident-response.sh
│
├── config/
│   ├── .env.production
│   ├── .env.staging
│   ├── logging.yaml
│   ├── metrics.yaml
│   └── security-policies.yaml
│
├── docs/
│   ├── README.md
│   ├── ARCHITECTURE.md
│   ├── API_SPECIFICATION.md
│   ├── SECURITY_GUIDELINES.md
│   ├── DEPLOYMENT.md
│   ├── OPERATIONAL_RUNBOOKS.md
│   ├── INCIDENT_RESPONSE.md
│   ├── COMPLIANCE.md
│   ├── DISASTER_RECOVERY.md
│   └── RUN_BOOK.md
│
├── monitoring/
│   ├── alerts.yaml                # AlertManager rules
│   ├── dashboards/
│   │   ├── system-health.json
│   │   ├── security-dashboard.json
│   │   ├── fraud-detection.json
│   │   └── ...
│   └── slos/
│       ├── availability.md
│       ├── latency.md
│       └── error_rate.md
│
├── security/
│   ├── secrets-vault/
│   │   └── vault-config.hcl
│   ├── security-policies.md
│   ├── incident-response-plan.md
│   ├── threat-model.md
│   └── penetration-test-results/
│
├── tests/
│   ├── load-tests/
│   │   └── load-test-config.js
│   ├── security-tests/
│   │   └── security-scan.yaml
│   └── integration-tests/
│
├── .github/
│   └── workflows/
│       ├── ci-cd.yml              # GitHub Actions
│       ├── security-scan.yml
│       ├── deploy.yml
│       └── ...
│
├── docker-compose.yml             # Local development
├── Makefile                        # Common commands
├── .gitignore
└── README.md                       # Project overview
```

---

## 11. NEXT STEPS & IMMEDIATE ACTIONS

### Phase 1 (Weeks 1-2): Critical Security Fixes
```
Priority 1 - MUST DO:
1. Implement refresh token rotation
2. Add brute-force protection & account lockout
3. Implement idempotency key mechanism
4. Add comprehensive input validation
5. Enable HTTPS/TLS on all endpoints
6. Implement device fingerprinting

Phase 2 (Weeks 3-4): Transaction Safety
1. Implement double-entry ledger
2. Add transaction reversal workflow
3. Implement audit logging
4. Add fraud detection framework (basic)
5. Implement rate limiting via Redis

Phase 3 (Months 2-3): Compliance & Infrastructure
1. Implement KYC/AML system
2. Setup centralized logging (ELK)
3. Implement monitoring/alerting
4. Setup backup/disaster recovery
5. Implement CI/CD pipeline security
```

---

## END OF AUDIT REPORT

**Assessment Date**: May 20, 2026  
**Recommendation**: System requires substantial security hardening before production use  
**Estimated Remediation Time**: 8-12 weeks for critical issues, 6 months for full production-grade setup

**For Complete Implementation, See**:
- `PRODUCTION_ARCHITECTURE_REDESIGN.md`
- `SECURITY_IMPLEMENTATION_GUIDE.md`
- `DATABASE_SCHEMA_GUIDE.md`
- `API_STANDARDS_GUIDE.md`
- `DEPLOYMENT_GUIDE.md`
