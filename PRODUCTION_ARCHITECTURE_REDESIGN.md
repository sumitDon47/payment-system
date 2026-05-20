# 🏗️ PRODUCTION-GRADE ARCHITECTURE REDESIGN
**System**: Digital Wallet Payment Platform  
**Target Deployment**: Kubernetes on AWS  
**Scale**: 1M+ daily active users  
**Compliance**: PCI DSS, KYC/AML, Data Protection

---

## PART 1: REFINED MICROSERVICES ARCHITECTURE

### Service Design Principles

```
┌─────────────────────────────────────────────────────────┐
│              DESIGN PRINCIPLES                           │
├─────────────────────────────────────────────────────────┤
│ 1. SINGLE RESPONSIBILITY                                │
│    - Each service has ONE business capability            │
│    - Clear, well-defined domain                          │
│                                                          │
│ 2. INDEPENDENT DEPLOYMENT                               │
│    - Services deployable without others                  │
│    - Versioned API contracts                             │
│    - Backward-compatible changes                         │
│                                                          │
│ 3. DECENTRALIZED DATA                                    │
│    - Each service owns its data                          │
│    - No shared databases                                 │
│    - Event-driven communication                          │
│                                                          │
│ 4. RESILIENT COMMUNICATION                               │
│    - Async messaging (Kafka) > Sync (gRPC)              │
│    - Circuit breakers for failures                       │
│    - Timeout & retry policies                            │
│                                                          │
│ 5. OBSERVABILITY                                         │
│    - Distributed tracing (Jaeger)                        │
│    - Centralized logging (ELK)                           │
│    - Metrics (Prometheus)                                │
└─────────────────────────────────────────────────────────┘
```

### Core Services

#### 1. **AUTH SERVICE** (Port 8001)

```go
// auth-service/cmd/main.go

package main

import (
    "github.com/payment-system/auth-service/internal/handler"
    "github.com/payment-system/auth-service/internal/service"
    "github.com/payment-system/auth-service/pkg/db"
    "github.com/payment-system/auth-service/pkg/cache"
)

func main() {
    // Initialize dependencies
    dbConn := db.NewPostgresConnection()
    redisCache := cache.NewRedisCache()
    
    // Initialize service layer
    authService := service.NewAuthService(dbConn, redisCache)
    tokenService := service.NewTokenService()
    mfaService := service.NewMFAService()
    passwordService := service.NewPasswordService()
    
    // Initialize handlers
    authHandler := handler.NewAuthHandler(
        authService, tokenService, mfaService, passwordService,
    )
    
    // Register routes
    http.HandleFunc("/auth/register", authHandler.Register)
    http.HandleFunc("/auth/login", authHandler.Login)
    http.HandleFunc("/auth/verify-otp", authHandler.VerifyOTP)
    http.HandleFunc("/auth/token/refresh", authHandler.RefreshToken)
    http.HandleFunc("/auth/token/revoke", authHandler.RevokeToken)
    http.HandleFunc("/auth/mfa/setup", authHandler.SetupMFA)
    http.HandleFunc("/auth/mfa/verify", authHandler.VerifyMFA)
    
    // Start server
    log.Fatal(http.ListenAndServe(":8001", nil))
}
```

**Database Schema**:
```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(254) UNIQUE NOT NULL,
    phone_number VARCHAR(20) UNIQUE,
    full_name VARCHAR(200) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    kyc_status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Refresh tokens
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    rotated_at TIMESTAMP
);

-- Device sessions
CREATE TABLE device_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255) UNIQUE NOT NULL,
    device_fingerprint VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    is_trusted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);

-- MFA settings
CREATE TABLE mfa_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_method VARCHAR(20), -- SMS, EMAIL, TOTP, BIOMETRIC
    mfa_secret VARCHAR(255), -- For TOTP
    backup_codes TEXT[], -- Recovery codes
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Login audit trail
CREATE TABLE login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    email VARCHAR(254),
    status VARCHAR(20), -- success, failed, blocked
    failure_reason VARCHAR(100),
    ip_address INET,
    device_id VARCHAR(255),
    country_code VARCHAR(2),
    timestamp TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_login_user ON login_attempts(user_id, timestamp DESC);
CREATE INDEX idx_login_email ON login_attempts(email, timestamp DESC);
```

---

#### 2. **WALLET SERVICE** (Port 8002)

```go
// wallet-service/cmd/main.go

package main

import (
    "github.com/payment-system/wallet-service/internal/handler"
    "github.com/payment-system/wallet-service/internal/service"
)

func main() {
    walletHandler := handler.NewWalletHandler()
    
    http.HandleFunc("/wallet/balance", authMiddleware(walletHandler.GetBalance))
    http.HandleFunc("/wallet/statement", authMiddleware(walletHandler.GetStatement))
    http.HandleFunc("/wallet/link-bank", authMiddleware(walletHandler.LinkBankAccount))
    http.HandleFunc("/wallet/withdraw", authMiddleware(walletHandler.InitiateWithdrawal))
    
    log.Fatal(http.ListenAndServe(":8002", nil))
}
```

**Key Implementation**:
```go
// internal/service/wallet_service.go

type WalletService struct {
    db    *sql.DB
    cache Cache
    kafka KafkaProducer
}

// GetBalance returns cached balance if fresh, else query ledger
func (ws *WalletService) GetBalance(ctx context.Context, userID string) (float64, error) {
    // Check cache (5-minute TTL)
    key := fmt.Sprintf("wallet:balance:%s", userID)
    if cached, err := ws.cache.Get(ctx, key); err == nil {
        var balance float64
        json.Unmarshal(cached, &balance)
        return balance, nil
    }
    
    // Query ledger
    var balance float64
    err := ws.db.QueryRowContext(ctx, `
        SELECT COALESCE(SUM(
            CASE 
                WHEN entry_type = 'DEBIT' THEN amount
                WHEN entry_type = 'CREDIT' THEN amount
            END
        ), 0)
        FROM ledger_entries
        WHERE user_id = $1 AND status = 'posted'
    `, userID).Scan(&balance)
    
    // Cache result
    bs, _ := json.Marshal(balance)
    ws.cache.SetEX(ctx, key, bs, 5*time.Minute)
    
    return balance, err
}
```

---

#### 3. **TRANSACTION SERVICE** (Port 8003)

```go
// transaction-service/internal/handler/payment_handler.go

type PaymentHandler struct {
    txnService    *TransactionService
    fraudService  *FraudDetectionClient
    walletService *WalletServiceClient
}

// POST /transactions/send - P2P transfer
func (ph *PaymentHandler) SendPayment(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    
    var req SendPaymentRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 1. Validate request
    if err := validatePaymentRequest(req); err != nil {
        respondError(w, err, http.StatusBadRequest)
        return
    }
    
    // 2. Assess fraud risk
    riskAssessment := ph.fraudService.AssessRisk(userID, req.ReceiverID, req.Amount)
    
    if riskAssessment.Recommendation == FraudActionBlock {
        respondError(w, errors.New("transaction blocked"), http.StatusForbidden)
        return
    }
    
    if riskAssessment.Recommendation == FraudActionChallenge {
        // Require MFA verification
        // Return 428 Precondition Required with MFA challenge
        respondMFAChallenge(w, userID)
        return
    }
    
    // 3. Process transaction
    txn, err := ph.txnService.SendPayment(ctx, SendPaymentParams{
        SenderID:   userID,
        ReceiverID: req.ReceiverID,
        Amount:     req.Amount,
        Currency:   req.Currency,
        IdempotencyKey: req.IdempotencyKey,
    })
    
    if err != nil {
        respondError(w, err, http.StatusInternalServerError)
        return
    }
    
    respondSuccess(w, txn)
}
```

**Transaction Processing Logic**:
```go
// internal/service/transaction_service.go

func (ts *TransactionService) SendPayment(ctx context.Context, params SendPaymentParams) error {
    // Check idempotency
    if cached := ts.getIdempotentResponse(ctx, params.IdempotencyKey); cached != nil {
        return cached
    }
    
    // Begin transaction
    tx, err := ts.db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelSerializable,
    })
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()
    
    // 1. Lock sender and receiver balances
    var senderBalance, receiverBalance float64
    tx.QueryRowContext(ctx, `
        SELECT balance FROM user_accounts WHERE user_id = $1 FOR UPDATE
    `, params.SenderID).Scan(&senderBalance)
    
    tx.QueryRowContext(ctx, `
        SELECT balance FROM user_accounts WHERE user_id = $1 FOR UPDATE
    `, params.ReceiverID).Scan(&receiverBalance)
    
    // 2. Validate balance
    if senderBalance < params.Amount {
        return fmt.Errorf("insufficient funds")
    }
    
    // 3. Create transaction record
    txnID, _ := uuid.NewV4()
    tx.ExecContext(ctx, `
        INSERT INTO transactions 
        (id, transaction_type, sender_id, receiver_id, amount, currency, 
         idempotency_key, status)
        VALUES ($1, 'P2P_TRANSFER', $2, $3, $4, $5, $6, 'pending')
    `, txnID.String(), params.SenderID, params.ReceiverID, 
       params.Amount, params.Currency, params.IdempotencyKey)
    
    // 4. Create ledger entries (MUST BALANCE)
    // Debit sender
    tx.ExecContext(ctx, `
        INSERT INTO ledger_entries 
        (transaction_id, user_id, account_id, amount, currency, entry_type, status)
        VALUES ($1, $2, '1010', $3, $4, 'DEBIT', 'pending')
    `, txnID.String(), params.SenderID, -params.Amount, params.Currency)
    
    // Credit receiver
    tx.ExecContext(ctx, `
        INSERT INTO ledger_entries 
        (transaction_id, user_id, account_id, amount, currency, entry_type, status)
        VALUES ($1, $2, '1010', $3, $4, 'CREDIT', 'pending')
    `, txnID.String(), params.ReceiverID, params.Amount, params.Currency)
    
    // 5. Update user account balances
    tx.ExecContext(ctx, `
        UPDATE user_accounts 
        SET balance = balance - $1 WHERE user_id = $2
    `, params.Amount, params.SenderID)
    
    tx.ExecContext(ctx, `
        UPDATE user_accounts 
        SET balance = balance + $1 WHERE user_id = $2
    `, params.Amount, params.ReceiverID)
    
    // 6. Mark ledger entries as posted
    tx.ExecContext(ctx, `
        UPDATE ledger_entries 
        SET status = 'posted', posted_at = NOW()
        WHERE transaction_id = $1
    `, txnID.String())
    
    // 7. Update transaction status
    tx.ExecContext(ctx, `
        UPDATE transactions 
        SET status = 'completed', completed_at = NOW()
        WHERE id = $1
    `, txnID.String())
    
    // 8. Create outbox event
    event := map[string]interface{}{
        "transaction_id": txnID.String(),
        "sender_id":      params.SenderID,
        "receiver_id":    params.ReceiverID,
        "amount":         params.Amount,
        "currency":       params.Currency,
        "timestamp":      time.Now().UTC(),
    }
    eventJSON, _ := json.Marshal(event)
    
    tx.ExecContext(ctx, `
        INSERT INTO outbox_events 
        (topic, event_key, payload, status)
        VALUES ('payment.completed', $1, $2::jsonb, 'pending')
    `, txnID.String(), string(eventJSON))
    
    // 9. Commit
    if err = tx.Commit(); err != nil {
        return err
    }
    
    // 10. Cache idempotent response
    ts.cacheIdempotentResponse(ctx, params.IdempotencyKey, response, 24*time.Hour)
    
    return nil
}
```

---

#### 4. **FRAUD DETECTION SERVICE** (Port 8005)

```go
// fraud-service/cmd/main.go

package main

import (
    "github.com/payment-system/fraud-service/internal/detector"
    "github.com/segmentio/kafka-go"
)

func main() {
    // Kafka consumer
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: []string{"kafka-1:9092", "kafka-2:9092", "kafka-3:9092"},
        Topic:   "payment.initiated",
        GroupID: "fraud-detection-group",
    })
    
    fraudDetector := detector.NewFraudDetector()
    
    for {
        msg, err := reader.ReadMessage(context.Background())
        if err != nil {
            log.Printf("Read error: %v", err)
            continue
        }
        
        // Parse transaction
        var txn Transaction
        json.Unmarshal(msg.Value, &txn)
        
        // Assess risk
        assessment := fraudDetector.AssessTransaction(txn)
        
        // Publish decision
        if assessment.Recommendation != FraudActionAllow {
            fraudWriter.WriteMessages(context.Background(),
                kafka.Message{
                    Key:   []byte(txn.ID),
                    Value: marshalAssessment(assessment),
                },
            )
        }
    }
}
```

**Fraud Detection Rules**:
```go
// internal/detector/fraud_rules.go

type FraudRuleSet struct {
    velocityCheck     *VelocityRule
    amountAnomaly     *AmountAnomalyRule
    geoVelocity       *GeoVelocityRule
    deviceBinding     *DeviceBindingRule
    timeAnomaly       *TimeAnomalyRule
    receiverRisk      *ReceiverRiskRule
}

func (fr *FraudRuleSet) EvaluateRules(txn Transaction) *RiskAssessment {
    assessment := &RiskAssessment{
        TransactionID: txn.ID,
        RiskScore:     0,
        Factors:       []RiskFactor{},
    }
    
    // Rule 1: Velocity check (5+ txns in 15 min = suspicious)
    velocityScore := fr.velocityCheck.Evaluate(txn)
    assessment.RiskScore += velocityScore * 0.5
    assessment.Factors = append(assessment.Factors, velocityScore)
    
    // Rule 2: Amount anomaly (3x average = suspicious)
    amountScore := fr.amountAnomaly.Evaluate(txn)
    assessment.RiskScore += amountScore * 0.4
    assessment.Factors = append(assessment.Factors, amountScore)
    
    // Rule 3: Geo-location velocity (impossible travel)
    geoScore := fr.geoVelocity.Evaluate(txn)
    assessment.RiskScore += geoScore * 0.6
    assessment.Factors = append(assessment.Factors, geoScore)
    
    // Rule 4: Device binding (new device = moderate risk)
    deviceScore := fr.deviceBinding.Evaluate(txn)
    assessment.RiskScore += deviceScore * 0.4
    assessment.Factors = append(assessment.Factors, deviceScore)
    
    // Rule 5: Time anomaly (transaction outside normal hours)
    timeScore := fr.timeAnomaly.Evaluate(txn)
    assessment.RiskScore += timeScore * 0.3
    assessment.Factors = append(assessment.Factors, timeScore)
    
    // Rule 6: Receiver risk (high-risk receiver)
    receiverScore := fr.receiverRisk.Evaluate(txn)
    assessment.RiskScore += receiverScore * 0.35
    assessment.Factors = append(assessment.Factors, receiverScore)
    
    // Determine recommendation
    switch {
    case assessment.RiskScore >= 80:
        assessment.Recommendation = FraudActionBlock
        assessment.Message = "Transaction blocked due to high fraud risk"
    case assessment.RiskScore >= 60:
        assessment.Recommendation = FraudActionChallenge
        assessment.Message = "Additional verification required"
    case assessment.RiskScore >= 40:
        assessment.Recommendation = FraudActionReview
        assessment.Message = "Flagged for manual review"
    default:
        assessment.Recommendation = FraudActionAllow
        assessment.Message = "Transaction approved"
    }
    
    return assessment
}
```

---

#### 5. **NOTIFICATION SERVICE**

```go
// notification-service/cmd/main.go

package main

import (
    "github.com/payment-system/notification-service/internal/consumer"
    "github.com/segmentio/kafka-go"
)

func main() {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: []string{"kafka-1:9092"},
        Topic:   "payment.completed",
        GroupID: "notification-service",
    })
    
    notificationConsumer := consumer.NewNotificationConsumer()
    
    for {
        msg, _ := reader.ReadMessage(context.Background())
        
        var event PaymentEvent
        json.Unmarshal(msg.Value, &event)
        
        // Route to appropriate handler
        switch event.EventType {
        case "payment.completed":
            notificationConsumer.HandlePaymentCompleted(event)
        case "payment.failed":
            notificationConsumer.HandlePaymentFailed(event)
        case "fraud.alert":
            notificationConsumer.HandleFraudAlert(event)
        }
        
        reader.CommitMessages(context.Background(), msg)
    }
}
```

**Notification Channels**:
```go
// internal/notifier/email_notifier.go

type EmailNotifier struct {
    sgClient *mail.Client
}

func (en *EmailNotifier) SendPaymentReceipt(txn Transaction) error {
    message := &mail.SGMailV3{
        From: &mail.Email{Address: "noreply@payment-system.com"},
        To: []*mail.Email{{Address: txn.SenderEmail}},
        Subject: fmt.Sprintf("Payment Receipt #%s", txn.ID),
        PlainTextContent: fmt.Sprintf(`
Payment Completed
================
Amount: %.2f %s
Receiver: %s
Reference: %s
Time: %s
        `, txn.Amount, txn.Currency, txn.ReceiverName, txn.ID, txn.CompletedAt),
    }
    
    _, err := en.sgClient.Send(message)
    return err
}

// internal/notifier/sms_notifier.go

type SMSNotifier struct {
    twilioClient *twilio.Client
}

func (sn *SMSNotifier) SendPaymentAlert(userPhone string, txn Transaction) error {
    message := fmt.Sprintf(
        "Payment of %.2f %s sent to %s. Ref: %s",
        txn.Amount, txn.Currency, txn.ReceiverName, txn.ID,
    )
    
    _, err := sn.twilioClient.Messages.SendMessage(
        os.Getenv("TWILIO_PHONE"),
        userPhone,
        message,
        nil,
    )
    
    return err
}
```

---

## PART 2: API GATEWAY & SECURITY LAYER

```go
// api-gateway/cmd/main.go

package main

import (
    "github.com/payment-system/api-gateway/internal/middleware"
    "github.com/payment-system/api-gateway/internal/router"
)

func main() {
    mux := http.NewServeMux()
    
    // Apply middleware stack
    handler := mux
    handler = middleware.LoggingMiddleware(handler)
    handler = middleware.RateLimitMiddleware(handler)
    handler = middleware.SecurityHeadersMiddleware(handler)
    handler = middleware.AuthenticationMiddleware(handler)
    handler = middleware.RequestValidationMiddleware(handler)
    
    // Route requests to services
    router.RegisterRoutes(mux)
    
    log.Fatal(http.ListenAndServe(":8000", handler))
}
```

**Security Middleware**:
```go
// internal/middleware/auth_middleware.go

func AuthenticationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip for public endpoints
        if isPublicPath(r.URL.Path) {
            next.ServeHTTP(w, r)
            return
        }
        
        // Extract token
        token := extractBearerToken(r.Header.Get("Authorization"))
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        // Verify token
        claims, err := verifyJWT(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // Verify device binding
        deviceID := r.Header.Get("X-Device-ID")
        if !verifyDeviceBinding(claims.UserID, claims.SessionID, deviceID) {
            http.Error(w, "Device mismatch", http.StatusForbidden)
            return
        }
        
        // Add claims to request context
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "claims", claims)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// internal/middleware/rate_limit.go

func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := extractUserID(r)
        
        // Check rate limit (100 requests/minute per user)
        limit := 100
        window := 60 * time.Second
        
        key := fmt.Sprintf("ratelimit:%s", userID)
        count, err := redis.Incr(key)
        if count == 1 {
            redis.Expire(key, window)
        }
        
        if count > int64(limit) {
            w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
            w.Header().Set("X-RateLimit-Remaining", "0")
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        
        w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
        w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-int(count)))
        
        next.ServeHTTP(w, r)
    })
}
```

---

## PART 3: DATABASE & LEDGER ARCHITECTURE

**Complete Ledger Schema**:
```sql
-- General Ledger Accounts
CREATE TABLE accounts (
    account_id VARCHAR(50) PRIMARY KEY,
    account_name VARCHAR(100) NOT NULL,
    account_type VARCHAR(50) NOT NULL, -- ASSET, LIABILITY, EQUITY, INCOME, EXPENSE
    normal_balance_side VARCHAR(10) NOT NULL, -- DEBIT or CREDIT
    currency VARCHAR(3) NOT NULL DEFAULT 'NPR',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Sample GL accounts
INSERT INTO accounts VALUES
    -- Assets
    ('1010', 'User Wallets - Active', 'ASSET', 'DEBIT', 'NPR', TRUE, NOW()),
    ('1020', 'User Wallets - Dormant', 'ASSET', 'DEBIT', 'NPR', TRUE, NOW()),
    ('1030', 'Merchant Payouts Pending', 'ASSET', 'DEBIT', 'NPR', TRUE, NOW()),
    -- Liabilities
    ('2010', 'Bank Transfer Payables', 'LIABILITY', 'CREDIT', 'NPR', TRUE, NOW()),
    ('2020', 'Merchant Settlement Payables', 'LIABILITY', 'CREDIT', 'NPR', TRUE, NOW()),
    -- Equity
    ('3010', 'Retained Earnings', 'EQUITY', 'CREDIT', 'NPR', TRUE, NOW()),
    -- Income
    ('4010', 'Transaction Fees', 'INCOME', 'CREDIT', 'NPR', TRUE, NOW()),
    ('4020', 'Merchant Fees', 'INCOME', 'CREDIT', 'NPR', TRUE, NOW()),
    -- Expense
    ('5010', 'Bank Charges', 'EXPENSE', 'DEBIT', 'NPR', TRUE, NOW());

-- User account mapping
CREATE TABLE user_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id),
    account_id VARCHAR(50) NOT NULL REFERENCES accounts(account_id),
    currency VARCHAR(3) NOT NULL DEFAULT 'NPR',
    balance NUMERIC(15, 4) NOT NULL DEFAULT 0,
    last_transaction_id UUID,
    last_updated_at TIMESTAMP DEFAULT NOW()
);

-- Immutable ledger entries
CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL,
    journal_entry_id UUID NOT NULL, -- Group related debits/credits
    user_id UUID REFERENCES users(id),
    account_id VARCHAR(50) NOT NULL REFERENCES accounts(account_id),
    
    amount NUMERIC(15, 4) NOT NULL CHECK (amount != 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'NPR',
    entry_type VARCHAR(20) NOT NULL CHECK (entry_type IN ('DEBIT', 'CREDIT')),
    
    description TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, posted, reversed
    
    posted_at TIMESTAMP,
    reversed_at TIMESTAMP,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Immutability: No updates allowed after posting
    CONSTRAINT ledger_immutable CHECK (status != 'posted' OR reversed_at IS NULL)
);

CREATE INDEX idx_ledger_user ON ledger_entries(user_id, posted_at DESC);
CREATE INDEX idx_ledger_account ON ledger_entries(account_id, posted_at DESC);
CREATE INDEX idx_ledger_txn ON ledger_entries(transaction_id);
CREATE INDEX idx_ledger_journal ON ledger_entries(journal_entry_id);

-- Immutable transaction log
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_type VARCHAR(50) NOT NULL, -- P2P_TRANSFER, MERCHANT_PAYMENT, WITHDRAWAL, DEPOSIT
    
    sender_id UUID REFERENCES users(id),
    receiver_id UUID REFERENCES users(id),
    merchant_id UUID,
    
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
    
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_txn_sender ON transactions(sender_id, created_at DESC) WHERE status != 'failed';
CREATE INDEX idx_txn_receiver ON transactions(receiver_id, created_at DESC);
CREATE INDEX idx_txn_status ON transactions(status, created_at DESC);
CREATE INDEX idx_txn_idempotency ON transactions(idempotency_key);

-- Reconciliation audit trail
CREATE TABLE reconciliation_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reconciliation_date DATE NOT NULL,
    account_id VARCHAR(50) NOT NULL,
    ledger_balance NUMERIC(15, 4) NOT NULL,
    stored_balance NUMERIC(15, 4) NOT NULL,
    variance NUMERIC(15, 4),
    status VARCHAR(20) NOT NULL, -- passed, failed, manual_review
    reviewed_by VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Reconciliation Stored Procedure**:
```sql
-- Verify all accounts balance
CREATE OR REPLACE FUNCTION reconcile_all_accounts()
RETURNS TABLE(account_id VARCHAR, status VARCHAR, variance NUMERIC) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        a.account_id,
        CASE 
            WHEN ABS(ledger_balance - stored_balance) < 0.01 THEN 'PASS'
            ELSE 'FAIL'
        END,
        ledger_balance - stored_balance as variance
    FROM (
        SELECT 
            a.account_id,
            COALESCE(SUM(CASE WHEN l.entry_type = 'DEBIT' THEN l.amount WHEN l.entry_type = 'CREDIT' THEN l.amount ELSE 0 END), 0) as ledger_balance,
            COALESCE(ua.balance, 0) as stored_balance
        FROM accounts a
        LEFT JOIN ledger_entries l ON a.account_id = l.account_id AND l.status = 'posted'
        LEFT JOIN user_accounts ua ON a.account_id = ua.account_id
        GROUP BY a.account_id, ua.balance
    ) AS reconciliation_check;
END;
$$ LANGUAGE plpgsql;

-- Run daily reconciliation
SELECT * FROM reconcile_all_accounts();
```

---

## PART 4: DEPLOYMENT ON KUBERNETES

```yaml
# kubernetes/services/auth-service-deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: payment-system
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - auth-service
              topologyKey: kubernetes.io/hostname
      
      containers:
      - name: auth-service
        image: registry.example.com/auth-service:1.0.0
        imagePullPolicy: IfNotPresent
        
        ports:
        - name: http
          containerPort: 8001
        - name: metrics
          containerPort: 9001
        
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: database-config
              key: host
        - name: DB_PORT
          value: "5432"
        - name: DB_NAME
          value: "payment_db"
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: database-credentials
              key: username
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: database-credentials
              key: password
        - name: REDIS_HOST
          valueFrom:
            configMapKeyRef:
              name: redis-config
              key: host
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret
        - name: KAFKA_BROKERS
          value: "kafka-0.kafka-headless:9092,kafka-1.kafka-headless:9092,kafka-2.kafka-headless:9092"
        
        livenessProbe:
          httpGet:
            path: /health
            port: 8001
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        
        readinessProbe:
          httpGet:
            path: /ready
            port: 8001
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2
        
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 1000
          capabilities:
            drop:
            - ALL

      securityContext:
        fsGroup: 1000

---
apiVersion: v1
kind: Service
metadata:
  name: auth-service
  namespace: payment-system
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 8001
    targetPort: 8001
  - name: metrics
    port: 9001
    targetPort: 9001
  selector:
    app: auth-service

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: auth-service-hpa
  namespace: payment-system
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: auth-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
```

---

## PART 5: MONITORING & OBSERVABILITY

```yaml
# monitoring/prometheus-rules.yaml

apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: payment-system-alerts
  namespace: payment-system
spec:
  groups:
  - name: payment-system.rules
    interval: 30s
    rules:
    
    # Alert: High error rate
    - alert: PaymentServiceHighErrorRate
      expr: |
        (
          sum(rate(http_requests_total{service="payment-service",status=~"5.."}[5m]))
          /
          sum(rate(http_requests_total{service="payment-service"}[5m]))
        ) > 0.05
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "High error rate on payment service"
        description: "Error rate is {{ $value | humanizePercentage }} for payment service"
    
    # Alert: High latency
    - alert: PaymentServiceHighLatency
      expr: |
        histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{service="payment-service"}[5m])) > 1
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High latency on payment service"
        description: "P99 latency is {{ $value }}s"
    
    # Alert: Database connection pool exhausted
    - alert: DatabaseConnectionPoolExhausted
      expr: |
        db_connections_in_use / db_connections_max > 0.9
      for: 2m
      labels:
        severity: critical
      annotations:
        summary: "Database connection pool nearing capacity"
    
    # Alert: Fraud detection backlog
    - alert: FraudDetectionBacklog
      expr: |
        kafka_consumergroup_lag{group="fraud-detection-group"} > 1000
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "Fraud detection consumer lag is high"
        description: "Consumer lag is {{ $value }} messages"

---
# monitoring/dashboard-alerts.json

{
  "dashboard": {
    "title": "Payment System Real-Time Dashboard",
    "panels": [
      {
        "title": "Transaction Success Rate",
        "targets": [
          {
            "expr": "sum(rate(transactions_completed_total[5m])) / sum(rate(transactions_initiated_total[5m]))"
          }
        ],
        "gauge": {
          "maxValue": 1,
          "minValue": 0,
          "thresholds": ["0.95", "0.99"]
        }
      },
      {
        "title": "Fraud Detection Accuracy",
        "targets": [
          {
            "expr": "fraud_detection_tp / (fraud_detection_tp + fraud_detection_fp)"
          }
        ],
        "gauge": {
          "maxValue": 1,
          "thresholds": ["0.90", "0.95"]
        }
      },
      {
        "title": "Transaction Volume (Per Minute)",
        "targets": [
          {
            "expr": "sum(rate(transactions_completed_total[1m]))"
          }
        ],
        "graph": {}
      },
      {
        "title": "API Response Time (P50, P95, P99)",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "P50"
          },
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "P95"
          },
          {
            "expr": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "P99"
          }
        ],
        "graph": {}
      }
    ]
  }
}
```

---

## PART 6: SECURITY IMPLEMENTATION CHECKLIST

```
IMPLEMENTATION PRIORITY:

WEEK 1-2 (CRITICAL):
═══════════════════════════════════════════════════════════════
✅ JWT + Refresh Token Rotation
   - Implement refresh token table
   - Add token rotation on refresh
   - Add device binding to tokens
   - Implement token revocation

✅ Brute-Force Protection
   - Implement rate limiter (Redis-backed)
   - Add account lockout (15min after 5 failures)
   - Add CAPTCHA challenge
   - Add login attempt logging

✅ Input Validation
   - Add email format validation (RFC 5321)
   - Add password complexity enforcement
   - Add XSS/SQL injection prevention
   - Add request size limits

✅ IDEMPOTENCY:
   - Add idempotency_key to all payment requests
   - Store request hash in database
   - Return cached response on retry

WEEK 3-4 (HIGH):
═══════════════════════════════════════════════════════════════
✅ Fraud Detection
   - Velocity checks
   - Amount anomaly detection
   - Geo-location velocity
   - Receiver risk scoring

✅ Device Fingerprinting
   - Capture device attributes on login
   - Bind device to session
   - Verify device on each request
   - Alert on device mismatch

✅ Audit Logging
   - Log all user actions
   - Store in immutable audit table
   - 7-year retention
   - Log integrity verification (hashing)

✅ Secrets Management
   - Move to HashiCorp Vault
   - Rotate credentials every 90 days
   - Remove hardcoded secrets
   - Implement secret access logging

MONTH 2 (MEDIUM):
═══════════════════════════════════════════════════════════════
✅ KYC/AML Implementation
   - Document verification workflow
   - Sanctions list screening
   - PEP identification
   - Risk scoring

✅ MFA/2FA
   - SMS OTP
   - Email OTP
   - TOTP support
   - Recovery codes

✅ Monitoring & Alerting
   - Centralized logging (ELK)
   - Metrics (Prometheus)
   - Alerting (PagerDuty)
   - Distributed tracing (Jaeger)

MONTH 3+ (LONG-TERM):
═══════════════════════════════════════════════════════════════
✅ Advanced Fraud Detection
   - ML models (behavioral analysis)
   - Graph fraud detection
   - Real-time decision engine

✅ Performance Optimization
   - Database query optimization
   - Caching strategy
   - Read replicas
   - Sharding

✅ Compliance
   - Regulatory documentation
   - Audit readiness
   - Incident response plan
   - DR/BC plan
```

---

**Next Document**: See `SECURITY_IMPLEMENTATION_GUIDE.md` for detailed code examples of each security component.

