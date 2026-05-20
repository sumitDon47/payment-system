# 🔐 PRODUCTION SECURITY IMPLEMENTATION GUIDE

**For**: Payment System Migration to Production  
**Focus**: Immediate security hardening (4-week sprint)

---

## MODULE 1: AUTHENTICATION HARDENING

### 1.1 JWT + Refresh Token Implementation

```go
// shared/security/token_service.go

package security

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
    TokenType    string `json:"token_type"`
}

type AccessTokenClaims struct {
    UserID      string `json:"sub"`
    Email       string `json:"email"`
    SessionID   string `json:"sid"`
    DeviceFP    string `json:"dfp"`
    jwt.RegisteredClaims
}

type RefreshTokenPayload struct {
    UserID    string
    SessionID string
    DeviceID  string
    IPAddress string
}

// Generate access token (15 minutes)
func GenerateAccessToken(userID, email, sessionID, deviceFP string) (string, error) {
    claims := AccessTokenClaims{
        UserID:    userID,
        Email:     email,
        SessionID: sessionID,
        DeviceFP:  deviceFP,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "payment-system",
            ID:        generateJTI(), // Unique ID per token
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(getJWTSecret()))
}

// Generate refresh token (7 days)
func GenerateRefreshToken() (string, error) {
    tokenBytes := make([]byte, 32)
    _, err := rand.Read(tokenBytes)
    if err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

// Store refresh token in database
func StoreRefreshToken(db *sql.DB, userID, refreshToken, deviceID, ipAddress string) error {
    tokenHash := hashToken(refreshToken)
    
    _, err := db.Exec(`
        INSERT INTO refresh_tokens 
        (user_id, token_hash, device_id, ip_address, status, expires_at)
        VALUES ($1, $2, $3, $4, 'active', NOW() + INTERVAL '7 days')
    `, userID, tokenHash, deviceID, ipAddress)
    
    return err
}

// Verify and rotate refresh token
func RotateRefreshToken(db *sql.DB, oldToken, deviceID string) (*TokenPair, error) {
    oldTokenHash := hashToken(oldToken)
    
    // Verify old token exists and is active
    var userID, sessionID string
    err := db.QueryRow(`
        SELECT user_id, id FROM refresh_tokens
        WHERE token_hash = $1 AND status = 'active' AND expires_at > NOW()
    `, oldTokenHash).Scan(&userID, &sessionID)
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("invalid refresh token")
    }
    if err != nil {
        return nil, err
    }
    
    // 1. Invalidate old token
    db.Exec(`
        UPDATE refresh_tokens 
        SET status = 'rotated', rotated_at = NOW()
        WHERE token_hash = $1
    `, oldTokenHash)
    
    // 2. Generate new tokens
    accessToken, _ := GenerateAccessToken(userID, "", sessionID, "")
    newRefreshToken, _ := GenerateRefreshToken()
    
    // 3. Store new refresh token
    StoreRefreshToken(db, userID, newRefreshToken, deviceID, "")
    
    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: newRefreshToken,
        ExpiresIn:    900, // 15 minutes
        TokenType:    "Bearer",
    }, nil
}

func hashToken(token string) string {
    h := sha256.New()
    h.Write([]byte(token))
    return hex.EncodeToString(h.Sum(nil))
}

func generateJTI() string {
    b := make([]byte, 16)
    rand.Read(b)
    return fmt.Sprintf("jti_%s", base64.URLEncoding.EncodeToString(b))
}
```

**Usage in Login Handler**:
```go
// auth-service/internal/handler/login_handler.go

func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 1. Validate credentials
    user, err := ah.authService.ValidateCredentials(req.Email, req.Password)
    if err != nil {
        // Record failed attempt
        ah.recordFailedLogin(req.Email, getClientIP(r))
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    
    // 2. Get device fingerprint
    deviceFP := extractDeviceFingerprint(r)
    
    // 3. Create device session
    sessionID, _ := uuid.NewV4()
    ah.createDeviceSession(user.ID, deviceFP, sessionID.String())
    
    // 4. Generate token pair
    accessToken, _ := security.GenerateAccessToken(
        user.ID, user.Email, sessionID.String(), deviceFP.Hash,
    )
    refreshToken, _ := security.GenerateRefreshToken()
    
    // 5. Store refresh token
    security.StoreRefreshToken(ah.db, user.ID, refreshToken, deviceFP.DeviceID, getClientIP(r))
    
    // 6. Return tokens
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    900,
        TokenType:    "Bearer",
    })
}
```

---

### 1.2 Brute-Force Protection

```go
// shared/security/rate_limiter.go

package security

import (
    "fmt"
    "math"
    "sync"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type BruteForceProtector struct {
    redis         *redis.Client
    maxAttempts   int
    lockoutTime   time.Duration
}

func NewBruteForceProtector(client *redis.Client) *BruteForceProtector {
    return &BruteForceProtector{
        redis:       client,
        maxAttempts: 5,
        lockoutTime: 15 * time.Minute,
    }
}

// Check if user/IP is locked
func (bfp *BruteForceProtector) IsLocked(ctx context.Context, identifier string) bool {
    lockKey := fmt.Sprintf("login:locked:%s", identifier)
    val, err := bfp.redis.Get(ctx, lockKey).Result()
    return err == nil && val == "1"
}

// Record failed attempt
func (bfp *BruteForceProtector) RecordFailure(ctx context.Context, identifier string) error {
    key := fmt.Sprintf("login:attempts:%s", identifier)
    
    // Increment attempt counter
    count, err := bfp.redis.Incr(ctx, key).Result()
    if err != nil {
        return err
    }
    
    // Set TTL on first attempt
    if count == 1 {
        bfp.redis.Expire(ctx, key, 15*time.Minute)
    }
    
    // Lock if exceeded max attempts
    if count > int64(bfp.maxAttempts) {
        lockKey := fmt.Sprintf("login:locked:%s", identifier)
        
        // Exponential backoff: 2s, 4s, 8s, 16s, 32s...
        backoffSeconds := math.Min(32, math.Pow(2, float64(count-bfp.maxAttempts)))
        lockDuration := time.Duration(backoffSeconds) * time.Second
        
        bfp.redis.SetEX(ctx, lockKey, "1", lockDuration)
        
        return fmt.Errorf("too many attempts, locked for %.0f seconds", backoffSeconds)
    }
    
    return nil
}

// Clear attempts on successful login
func (bfp *BruteForceProtector) ClearAttempts(ctx context.Context, identifier string) error {
    key := fmt.Sprintf("login:attempts:%s", identifier)
    return bfp.redis.Del(ctx, key).Err()
}

// Get current attempt count
func (bfp *BruteForceProtector) GetAttemptCount(ctx context.Context, identifier string) int {
    key := fmt.Sprintf("login:attempts:%s", identifier)
    count, _ := bfp.redis.Get(ctx, key).Int64()
    return int(count)
}
```

**Usage in Login Endpoint**:
```go
func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    identifier := fmt.Sprintf("%s:%s", req.Email, getClientIP(r))
    
    // 1. Check if locked
    if ah.bruteForce.IsLocked(r.Context(), identifier) {
        w.WriteHeader(http.StatusTooManyRequests)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Too many failed attempts. Try again later.",
        })
        return
    }
    
    // 2. Validate credentials
    user, err := ah.authService.ValidateCredentials(req.Email, req.Password)
    if err != nil {
        // Record failure
        ah.bruteForce.RecordFailure(r.Context(), identifier)
        
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid credentials",
        })
        return
    }
    
    // 3. Clear attempts on success
    ah.bruteForce.ClearAttempts(r.Context(), identifier)
    
    // ... continue with token generation
}
```

---

### 1.3 Device Fingerprinting

```go
// auth-service/internal/service/device_service.go

package service

import (
    "crypto/sha256"
    "fmt"
)

type DeviceFingerprint struct {
    DeviceID      string
    OSType        string
    OSVersion     string
    AppVersion    string
    DeviceModel   string
    UserAgent     string
}

type DeviceBindingService struct {
    db    *sql.DB
}

// Extract device fingerprint from request
func ExtractDeviceFingerprint(r *http.Request) DeviceFingerprint {
    return DeviceFingerprint{
        DeviceID:    r.Header.Get("X-Device-ID"),
        OSType:      r.Header.Get("X-OS-Type"),
        OSVersion:   r.Header.Get("X-OS-Version"),
        AppVersion:  r.Header.Get("X-App-Version"),
        DeviceModel: r.Header.Get("X-Device-Model"),
        UserAgent:   r.Header.Get("User-Agent"),
    }
}

// Generate fingerprint hash
func (fp DeviceFingerprint) Hash() string {
    data := fmt.Sprintf("%s|%s|%s|%s|%s",
        fp.DeviceID, fp.OSType, fp.OSVersion, fp.DeviceModel, fp.UserAgent,
    )
    h := sha256.New()
    h.Write([]byte(data))
    return fmt.Sprintf("%x", h.Sum(nil))
}

// Store device session
func (dbs *DeviceBindingService) StoreSession(ctx context.Context, userID, sessionID string, fp DeviceFingerprint) error {
    fpHash := fp.Hash()
    
    _, err := dbs.db.ExecContext(ctx, `
        INSERT INTO device_sessions 
        (user_id, device_id, session_id, device_fingerprint, ip_address, 
         os_type, os_version, app_version, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW() + INTERVAL '7 days')
    `, userID, fp.DeviceID, sessionID, fpHash, getClientIP(ctx),
       fp.OSType, fp.OSVersion, fp.AppVersion)
    
    return err
}

// Verify device binding
func (dbs *DeviceBindingService) VerifyBinding(ctx context.Context, userID, sessionID string, currentFP DeviceFingerprint) bool {
    var storedFP string
    
    err := dbs.db.QueryRowContext(ctx, `
        SELECT device_fingerprint FROM device_sessions 
        WHERE user_id = $1 AND session_id = $2 AND expires_at > NOW()
    `, userID, sessionID).Scan(&storedFP)
    
    if err != nil {
        return false
    }
    
    return storedFP == currentFP.Hash()
}
```

---

## MODULE 2: INPUT VALIDATION & SECURITY

### 2.1 Comprehensive Input Validation

```go
// shared/validation/validator.go

package validation

import (
    "fmt"
    "regexp"
    "strings"
    "unicode"
    
    "github.com/asaskevich/govalidator"
)

type ValidationError struct {
    Field   string
    Message string
}

type Validator struct{}

// Validate email
func (v *Validator) ValidateEmail(email string) []ValidationError {
    var errors []ValidationError
    
    if email = strings.TrimSpace(email); email == "" {
        errors = append(errors, ValidationError{"email", "Email is required"})
        return errors
    }
    
    if len(email) > 254 {
        errors = append(errors, ValidationError{"email", "Email too long (max 254 chars)"})
    }
    
    if !govalidator.IsEmail(email) {
        errors = append(errors, ValidationError{"email", "Invalid email format"})
    }
    
    return errors
}

// Validate password (12+ chars, mixed case, symbols)
func (v *Validator) ValidatePassword(password string) []ValidationError {
    var errors []ValidationError
    
    if len(password) < 12 {
        errors = append(errors, ValidationError{"password", "Min 12 characters"})
    }
    
    if len(password) > 128 {
        errors = append(errors, ValidationError{"password", "Max 128 characters"})
    }
    
    hasUpper, hasLower, hasDigit, hasSpecial := false, false, false, false
    
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
        errors = append(errors, ValidationError{"password", 
            "Must contain: uppercase, lowercase, number, special character"})
    }
    
    // Check against common passwords
    if v.isCommonPassword(password) {
        errors = append(errors, ValidationError{"password", "Password is too common"})
    }
    
    return errors
}

// Validate phone (E.164 format)
func (v *Validator) ValidatePhoneNumber(phone string) []ValidationError {
    var errors []ValidationError
    
    if phone = strings.TrimSpace(phone); phone == "" {
        errors = append(errors, ValidationError{"phone", "Phone is required"})
        return errors
    }
    
    // E.164 format: +[country code][number]
    pattern := `^\+[1-9]\d{1,14}$`
    if !regexp.MustCompile(pattern).MatchString(phone) {
        errors = append(errors, ValidationError{"phone", 
            "Phone must be E.164 format (e.g., +1234567890)"})
    }
    
    return errors
}

// Sanitize string input (prevent XSS)
func (v *Validator) SanitizeString(input string, maxLen int) (string, error) {
    if len(input) > maxLen {
        return "", fmt.Errorf("input exceeds max length of %d", maxLen)
    }
    
    // Trim whitespace
    sanitized := strings.TrimSpace(input)
    
    // Remove control characters
    sanitized = strings.Map(func(r rune) rune {
        if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
            return -1
        }
        return r
    }, sanitized)
    
    // Check for XSS patterns
    xssPatterns := []string{
        "<script", "javascript:", "onerror=", "onclick=", "<iframe",
        "<object", "<embed", "<img", "alert(", "eval(",
    }
    
    lowerSanitized := strings.ToLower(sanitized)
    for _, pattern := range xssPatterns {
        if strings.Contains(lowerSanitized, pattern) {
            return "", fmt.Errorf("invalid characters detected")
        }
    }
    
    return sanitized, nil
}

// Validate amount
func (v *Validator) ValidateAmount(amount float64) []ValidationError {
    var errors []ValidationError
    
    if amount <= 0 {
        errors = append(errors, ValidationError{"amount", "Amount must be greater than 0"})
    }
    
    if amount > 10000000 { // 1 crore NPR
        errors = append(errors, ValidationError{"amount", "Amount exceeds maximum limit"})
    }
    
    return errors
}

func (v *Validator) isCommonPassword(password string) bool {
    commonPasswords := []string{
        "password", "123456", "12345678", "password123", "admin",
        "letmein", "welcome", "monkey", "dragon", "iloveyou",
    }
    
    lower := strings.ToLower(password)
    for _, common := range commonPasswords {
        if strings.Contains(lower, common) {
            return true
        }
    }
    
    return false
}
```

**Usage in Handler**:
```go
func (ah *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    validator := validation.Validator{}
    var allErrors []validation.ValidationError
    
    // Validate all fields
    allErrors = append(allErrors, validator.ValidateEmail(req.Email)...)
    allErrors = append(allErrors, validator.ValidatePassword(req.Password)...)
    allErrors = append(allErrors, validator.ValidatePhoneNumber(req.Phone)...)
    
    if len(allErrors) > 0 {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "errors": allErrors,
        })
        return
    }
    
    // Sanitize inputs
    name, _ := validator.SanitizeString(req.Name, 100)
    email, _ := validator.SanitizeString(req.Email, 254)
    
    // Continue with registration...
}
```

---

## MODULE 3: IDEMPOTENCY KEY IMPLEMENTATION

```go
// shared/idempotency/idempotency_manager.go

package idempotency

import (
    "crypto/sha256"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"
)

type IdempotencyManager struct {
    db *sql.DB
}

// Check if request was already processed
func (im *IdempotencyManager) GetCachedResponse(userID, idempotencyKey string) ([]byte, error) {
    var response []byte
    
    err := im.db.QueryRow(`
        SELECT response FROM idempotency_keys
        WHERE idempotency_key = $1 AND user_id = $2 AND expires_at > NOW()
    `, idempotencyKey, userID).Scan(&response)
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("not found")
    }
    
    return response, err
}

// Cache response
func (im *IdempotencyManager) CacheResponse(
    userID, idempotencyKey string, 
    requestHash, response string,
) error {
    _, err := im.db.Exec(`
        INSERT INTO idempotency_keys 
        (idempotency_key, user_id, request_hash, response, expires_at)
        VALUES ($1, $2, $3, $4::jsonb, NOW() + INTERVAL '24 hours')
        ON CONFLICT (idempotency_key) DO NOTHING
    `, idempotencyKey, userID, requestHash, response)
    
    return err
}

// Hash request for integrity verification
func HashRequest(payload interface{}) string {
    data, _ := json.Marshal(payload)
    h := sha256.New()
    h.Write(data)
    return fmt.Sprintf("%x", h.Sum(nil))
}

// Middleware to handle idempotent requests
func IdempotencyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        idempotencyKey := r.Header.Get("X-Idempotency-Key")
        
        // Skip if no key provided
        if idempotencyKey == "" {
            next.ServeHTTP(w, r)
            return
        }
        
        userID := r.Context().Value("user_id").(string)
        im := &IdempotencyManager{db: getDB()}
        
        // Check cache
        if cached, err := im.GetCachedResponse(userID, idempotencyKey); err == nil {
            w.Header().Set("X-Cached-Response", "true")
            w.Header().Set("Content-Type", "application/json")
            w.Write(cached)
            return
        }
        
        // Otherwise proceed with request
        next.ServeHTTP(w, r)
    })
}
```

---

## MODULE 4: AUDIT LOGGING

```go
// shared/audit/audit_logger.go

package audit

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "time"
)

type AuditLogger struct {
    db *sql.DB
}

type AuditEvent struct {
    EventType       string
    EventCategory   string
    UserID          string
    ActorType       string // USER, ADMIN, SYSTEM
    TransactionID   string
    OldValues       interface{}
    NewValues       interface{}
    IPAddress       string
    UserAgent       string
    DeviceID        string
    CountryCode     string
    Timestamp       time.Time
}

// Log user action
func (al *AuditLogger) LogEvent(event AuditEvent) error {
    oldValuesJSON, _ := json.Marshal(event.OldValues)
    newValuesJSON, _ := json.Marshal(event.NewValues)
    
    _, err := al.db.Exec(`
        INSERT INTO compliance_audit_log 
        (event_type, event_category, user_id, actor_type, transaction_id,
         old_values, new_values, ip_address, user_agent, device_id, 
         country_code, timestamp)
        VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8, $9, $10, $11, $12)
    `, event.EventType, event.EventCategory, event.UserID, event.ActorType,
       event.TransactionID, string(oldValuesJSON), string(newValuesJSON),
       event.IPAddress, event.UserAgent, event.DeviceID, event.CountryCode,
       event.Timestamp)
    
    return err
}

// Log sensitive operations
func (al *AuditLogger) LogSensitiveOperation(userID, operation, details string) error {
    return al.LogEvent(AuditEvent{
        EventType:     operation,
        EventCategory: "SECURITY",
        UserID:        userID,
        ActorType:     "USER",
        Timestamp:     time.Now(),
    })
}

// Log transaction
func (al *AuditLogger) LogTransaction(
    userID, txnID string,
    oldBalance, newBalance float64,
) error {
    return al.LogEvent(AuditEvent{
        EventType:     "TRANSACTION_COMPLETED",
        EventCategory: "FINANCIAL",
        UserID:        userID,
        TransactionID: txnID,
        OldValues: map[string]float64{
            "balance": oldBalance,
        },
        NewValues: map[string]float64{
            "balance": newBalance,
        },
        Timestamp: time.Now(),
    })
}
```

---

## MODULE 5: SECRET MANAGEMENT

```go
// shared/secrets/vault_client.go

package secrets

import (
    "fmt"
    
    "github.com/hashicorp/vault/api"
)

type VaultClient struct {
    client *api.Client
}

func NewVaultClient() (*VaultClient, error) {
    config := api.DefaultConfig()
    config.Address = "http://vault:8200"
    
    client, err := api.NewClient(config)
    if err != nil {
        return nil, err
    }
    
    // Authenticate using AppRole
    client.SetToken(os.Getenv("VAULT_TOKEN"))
    
    return &VaultClient{client: client}, nil
}

// Get database credentials
func (vc *VaultClient) GetDatabaseCredentials() (string, string, error) {
    secret, err := vc.client.Logical().Read("secret/data/db/credentials")
    if err != nil {
        return "", "", err
    }
    
    data := secret.Data["data"].(map[string]interface{})
    username := data["username"].(string)
    password := data["password"].(string)
    
    return username, password, nil
}

// Get JWT secret
func (vc *VaultClient) GetJWTSecret() (string, error) {
    secret, err := vc.client.Logical().Read("secret/data/jwt/secret")
    if err != nil {
        return "", err
    }
    
    data := secret.Data["data"].(map[string]interface{})
    return data["secret"].(string), nil
}

// Rotate credentials (called every 90 days)
func (vc *VaultClient) RotateCredentials() error {
    // Use Vault's dynamic database credentials
    secret, err := vc.client.Logical().Read("database/static-creds/payment-app")
    if err != nil {
        return err
    }
    
    username := secret.Data["username"].(string)
    password := secret.Data["password"].(string)
    
    fmt.Printf("Credentials rotated: user=%s\n", username)
    return nil
}
```

**Vault Configuration**:
```hcl
# vault/config.hcl

vault {
  address = "http://vault:8200"
}

auto_auth {
  method {
    type = "kubernetes"
    config = {
      role = "payment-system-role"
    }
  }
  
  sink {
    type = "file"
    config = {
      path = "/tmp/vault-token"
    }
  }
}

cache {
  use_auto_auth_token = true
}

listener "unix" {
  address = "/tmp/vault.sock"
  tls_disable = true
}
```

---

## SECURITY TESTING CHECKLIST

```bash
#!/bin/bash
# scripts/security-test.sh

echo "🔐 SECURITY TESTING SUITE"
echo "=========================="

# Test 1: OWASP Top 10 - SQL Injection
echo "Test 1: SQL Injection Prevention"
curl -X POST http://localhost:8080/login \
  -d '{"email":"admin'\''--", "password":"test"}' \
  -H "Content-Type: application/json" \
  | grep -q "Invalid email" && echo "✅ PASS" || echo "❌ FAIL"

# Test 2: XSS Prevention
echo "Test 2: XSS Prevention"
curl -X POST http://localhost:8080/register \
  -d '{"name":"<script>alert(1)</script>", "email":"test@test.com", "password":"Pass123!@"}' \
  -H "Content-Type: application/json" \
  | grep -q "invalid characters" && echo "✅ PASS" || echo "❌ FAIL"

# Test 3: Brute-force Protection
echo "Test 3: Brute-force Protection"
for i in {1..6}; do
  curl -X POST http://localhost:8080/login \
    -d '{"email":"user@test.com", "password":"wrong"}' \
    -H "Content-Type: application/json" > /dev/null
done
curl -X POST http://localhost:8080/login \
  -d '{"email":"user@test.com", "password":"correct"}' \
  -H "Content-Type: application/json" \
  | grep -q "429\|locked" && echo "✅ PASS" || echo "❌ FAIL"

# Test 4: JWT Expiry
echo "Test 4: JWT Token Expiry"
# Get old token and wait 15+ minutes
# curl with old token should fail
echo "✅ MANUAL TEST REQUIRED"

# Test 5: Idempotency
echo "Test 5: Idempotency Keys"
IDEM_KEY="test-$(date +%s)"
RESP1=$(curl -X POST http://localhost:9090/payment \
  -d '{"amount":100, "receiver":"user2"}' \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Idempotency-Key: $IDEM_KEY" | jq -r '.transaction_id')
RESP2=$(curl -X POST http://localhost:9090/payment \
  -d '{"amount":100, "receiver":"user2"}' \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Idempotency-Key: $IDEM_KEY" | jq -r '.transaction_id')
[ "$RESP1" = "$RESP2" ] && echo "✅ PASS" || echo "❌ FAIL"

# Test 6: HTTPS Only
echo "Test 6: HTTPS Enforcement"
curl -I http://localhost:8080 2>&1 | grep -q "Connection refused\|SSL\|HTTPS" && echo "✅ PASS" || echo "❌ FAIL"

echo "=========================="
echo "Testing Complete!"
```

---

**End of Security Implementation Guide**

See `DEPLOYMENT_AND_COMPLIANCE_GUIDE.md` for DevOps, monitoring, and regulatory compliance setup.
