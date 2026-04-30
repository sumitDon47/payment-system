package handler

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sumitDon47/payment-system/user-service/cache"
	"github.com/sumitDon47/payment-system/user-service/db"
	"github.com/sumitDon47/payment-system/user-service/email"
	"github.com/sumitDon47/payment-system/user-service/models"
	"github.com/sumitDon47/payment-system/user-service/utils"
	"golang.org/x/crypto/bcrypt"
)

// Cache is the Redis client — set from main.go after initializing Redis.
// If Redis is unavailable, this stays nil and all cache operations are skipped.
// The service degrades gracefully — slower but still correct.
var Cache *cache.Client

// HealthCheck godoc
// GET /health
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := map[string]string{
		"status":  "ok",
		"service": "user-service",
	}

	// Include Redis health if available
	if Cache != nil {
		if err := Cache.HealthCheck(r.Context()); err != nil {
			status["redis"] = "unavailable"
		} else {
			status["redis"] = "ok"
		}
	} else {
		status["redis"] = "disabled"
	}

	json.NewEncoder(w).Encode(status)
}

// Register godoc
// POST /register
func Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "name, email, and password are required"})
		return
	}

	if len(req.Password) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "password must be at least 8 characters"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to process password"})
		return
	}

	var user models.User
	query := `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, balance, created_at, updated_at
	`
	err = db.DB.QueryRowContext(r.Context(), query, req.Name, req.Email, string(hashedPassword)).Scan(
		&user.ID, &user.Name, &user.Email, &user.Balance, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		log.Printf("Register error: %v", err)
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "email already exists"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to generate token"})
		return
	}

	// Cache the initial balance (0) so first wallet check is fast
	if Cache != nil {
		Cache.SetBalance(r.Context(), user.ID, user.Balance)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.AuthResponse{Token: token, User: user})
}

// Login godoc
// POST /login
func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request body"})
		return
	}

	log.Printf("[LOGIN] Login attempt for email: %s", req.Email)

	var user models.User
	var hashedPassword string
	query := `SELECT id, name, email, password, balance, created_at, updated_at FROM users WHERE email = $1`
	err := db.DB.QueryRowContext(r.Context(), query, req.Email).Scan(
		&user.ID, &user.Name, &user.Email, &hashedPassword,
		&user.Balance, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		log.Printf("[LOGIN] User not found: %s", req.Email)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid email or password"})
		return
	}
	if err != nil {
		log.Printf("[LOGIN] Database error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	log.Printf("[LOGIN] User found: %s, comparing passwords", req.Email)
	log.Printf("[LOGIN] Password from request length: %d chars", len(req.Password))
	log.Printf("[LOGIN] Hashed password in DB length: %d chars", len(hashedPassword))

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		log.Printf("[LOGIN] PASSWORD MISMATCH for user %s. Error: %v", req.Email, err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid email or password"})
		return
	}

	log.Printf("[LOGIN] PASSWORD MATCHED for user: %s", req.Email)

	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		log.Printf("[LOGIN] Error generating token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to generate token"})
		return
	}

	log.Printf("[LOGIN] Login SUCCESSFUL for user: %s with token", req.Email)
	json.NewEncoder(w).Encode(models.AuthResponse{Token: token, User: user})
}

// GetProfile godoc
// GET /profile  (requires Authorization: Bearer <token>)
func GetProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Header.Get("X-User-ID")

	var user models.User
	query := `SELECT id, name, email, balance, created_at, updated_at FROM users WHERE id = $1`
	err := db.DB.QueryRowContext(r.Context(), query, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Balance, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "user not found"})
		return
	}

	json.NewEncoder(w).Encode(models.SuccessResponse{Data: user})
}

// GetWalletBalance godoc
// GET /wallet  (requires Authorization: Bearer <token>)
//
// Cache-aside pattern:
// 1. Check Redis first — if hit, return immediately (fast path)
// 2. On miss, query PostgreSQL (slow path)
// 3. Store result in Redis for next time
func GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Header.Get("X-User-ID")

	// ── Fast path: check Redis cache ─────────────────────────────────────────
	if Cache != nil {
		if balance, hit := Cache.GetBalance(r.Context(), userID); hit {
			log.Printf("🎯 Cache HIT  wallet user=%s balance=%.2f", userID, balance)
			json.NewEncoder(w).Encode(models.SuccessResponse{
				Message: "wallet balance fetched",
				Data:    map[string]float64{"balance": balance},
			})
			return
		}
		log.Printf("💨 Cache MISS wallet user=%s — querying DB", userID)
	}

	// ── Slow path: query PostgreSQL ───────────────────────────────────────────
	var balance float64
	err := db.DB.QueryRowContext(r.Context(),
		`SELECT balance FROM users WHERE id = $1`, userID,
	).Scan(&balance)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "could not fetch balance"})
		return
	}

	// ── Store in cache for next request ──────────────────────────────────────
	if Cache != nil {
		Cache.SetBalance(r.Context(), userID, balance)
		log.Printf("💾 Cached wallet user=%s balance=%.2f", userID, balance)
	}

	json.NewEncoder(w).Encode(models.SuccessResponse{
		Message: "wallet balance fetched",
		Data:    map[string]float64{"balance": balance},
	})
}

// InvalidateUserCache godoc
// POST /internal/cache/invalidate
// Called by payment-service after a payment commits to bust the cache.
// This endpoint is internal — not exposed to the public.
func InvalidateUserCache(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "method not allowed"})
		return
	}

	apiKey := strings.TrimSpace(os.Getenv("INTERNAL_API_KEY"))
	if apiKey == "" {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal endpoint not configured"})
		return
	}

	providedKey := strings.TrimSpace(r.Header.Get("X-Internal-API-Key"))
	if subtle.ConstantTimeCompare([]byte(providedKey), []byte(apiKey)) != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "unauthorized"})
		return
	}

	if Cache == nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "cache disabled"})
		return
	}

	var req struct {
		UserIDs []string `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.UserIDs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "user_ids required"})
		return
	}

	Cache.InvalidateMultiple(r.Context(), req.UserIDs...)
	log.Printf("🗑️  Cache invalidated for users: %v", req.UserIDs)

	json.NewEncoder(w).Encode(map[string]string{"status": "invalidated"})
}

// ── JWT claims struct needed for token validation ─────────────────────────────

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// ── Helper: context with timeout for DB queries ───────────────────────────────

func withTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 5*time.Second)
}

// ForgotPassword godoc
// POST /forgot-password
// Generates a password reset token that expires in 15 minutes and sends it via email
func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "email is required"})
		return
	}

	// Check if user exists and get their name
	var userID, userName string
	query := `SELECT id, name FROM users WHERE email = $1`
	err := db.DB.QueryRowContext(r.Context(), query, req.Email).Scan(&userID, &userName)
	if err == sql.ErrNoRows {
		// Return success even if email doesn't exist (security best practice)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.SuccessResponse{
			Message: "If the email exists, a password reset link has been sent",
		})
		return
	}
	if err != nil {
		log.Printf("Error checking user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	// Generate random token (32 bytes = 64 hex chars)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		log.Printf("Error generating token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to generate reset token"})
		return
	}
	resetToken := hex.EncodeToString(tokenBytes)

	// Store token in database with 15-minute expiry
	expiresAt := time.Now().Add(15 * time.Minute)
	insertQuery := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (token) DO NOTHING
	`
	_, err = db.DB.ExecContext(r.Context(), insertQuery, userID, resetToken, expiresAt)
	if err != nil {
		log.Printf("Error storing reset token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to create reset token"})
		return
	}

	// Send email with reset link
	emailClient := email.NewSendGridClient()
	if emailClient != nil {
		go func() {
			if err := emailClient.SendPasswordResetEmail(req.Email, userName, resetToken); err != nil {
				log.Printf("❌ Error sending password reset email to %s: %v", req.Email, err)
			}
		}()
	}

	log.Printf("🔑 Reset token generated for user %s (expires: %v)", userID, expiresAt)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Message: "If the email is registered, a password reset link has been sent to your inbox",
	})
}

// ResetPassword godoc
// POST /reset-password
// Resets password using a valid reset token
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "token and new_password are required"})
		return
	}

	if len(req.NewPassword) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "password must be at least 8 characters"})
		return
	}

	// Verify token exists and is not expired
	var userID string
	query := `
		SELECT user_id FROM password_reset_tokens
		WHERE token = $1 AND expires_at > NOW()
	`
	err := db.DB.QueryRowContext(r.Context(), query, req.Token).Scan(&userID)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid or expired reset token"})
		return
	}
	if err != nil {
		log.Printf("Error validating token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to process password"})
		return
	}

	log.Printf("[RESET] Password reset for user %s", userID)
	log.Printf("[RESET] New password length: %d chars", len(req.NewPassword))
	log.Printf("[RESET] Hashed password length: %d chars", len(hashedPassword))

	// Update user password
	updateQuery := `UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`
	result, err := db.DB.ExecContext(r.Context(), updateQuery, string(hashedPassword), userID)
	if err != nil {
		log.Printf("Error updating password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to reset password"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("[RESET] Database update complete - rows affected: %d", rowsAffected)

	// Delete used reset token
	deleteQuery := `DELETE FROM password_reset_tokens WHERE token = $1`
	_, _ = db.DB.ExecContext(r.Context(), deleteQuery, req.Token)

	// Invalidate cache if available
	if Cache != nil {
		Cache.InvalidateBalance(r.Context(), userID)
	}

	log.Printf("[RESET] Password reset SUCCESSFUL for user %s", userID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Message: "Password has been reset successfully",
	})
}

// RegisterWithOTP godoc
// POST /register-otp
// Generates an OTP and sends it to the user's email for verification
func RegisterWithOTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req models.RegisterWithOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "name, email, and password are required"})
		return
	}

	if len(req.Password) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "password must be at least 8 characters"})
		return
	}

	// Check if user already exists
	var existingUserID string
	query := `SELECT id FROM users WHERE email = $1`
	err := db.DB.QueryRowContext(r.Context(), query, req.Email).Scan(&existingUserID)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "email already registered"})
		return
	}
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error checking user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	// Generate OTP
	otp, err := utils.GenerateOTP()
	if err != nil {
		log.Printf("Error generating OTP: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to generate OTP"})
		return
	}

	// Hash password for storage
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to process password"})
		return
	}

	// Delete any existing OTP for this email
	deleteQuery := `DELETE FROM otp_codes WHERE email = $1 AND verified = false`
	_, _ = db.DB.ExecContext(r.Context(), deleteQuery, req.Email)

	// Store OTP with 10-minute expiry
	expiresAt := time.Now().Add(10 * time.Minute)
	insertQuery := `
		INSERT INTO otp_codes (email, code, name, password_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = db.DB.ExecContext(r.Context(), insertQuery, req.Email, otp, req.Name, string(hashedPassword), expiresAt)
	if err != nil {
		log.Printf("Error storing OTP: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to generate OTP"})
		return
	}

	// Send OTP via email
	emailClient := email.NewSendGridClient()
	if emailClient != nil {
		go func() {
			subject := "PaymentApp Verification Code 🔐"
			htmlContent := utils.FormatOTPMessage(req.Name, otp)
			if err := emailClient.SendEmail(req.Email, subject, htmlContent); err != nil {
				log.Printf("❌ Error sending OTP email to %s: %v", req.Email, err)
			} else {
				log.Printf("✅ OTP email sent to %s", req.Email)
			}
		}()
	}

	log.Printf("📧 OTP sent to %s (expires: %v)", req.Email, expiresAt)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.OTPResponse{
		Message: "Verification code sent to your email",
		Email:   req.Email,
	})
}

// VerifyOTP godoc
// POST /verify-otp
// Verifies OTP and creates the user account
func VerifyOTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req models.VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Email == "" || req.Code == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "email and code are required"})
		return
	}

	// Find and validate OTP
	var otpID, name, passwordHash string
	var expiresAt time.Time
	var attempts int

	otpQuery := `
		SELECT id, name, password_hash, expires_at, attempts
		FROM otp_codes
		WHERE email = $1 AND code = $2 AND verified = false
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := db.DB.QueryRowContext(r.Context(), otpQuery, req.Email, req.Code).Scan(
		&otpID, &name, &passwordHash, &expiresAt, &attempts,
	)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid verification code"})
		return
	}
	if err != nil {
		log.Printf("Error fetching OTP: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	// Check if OTP is expired
	if time.Now().After(expiresAt) {
		// Delete expired OTP
		_, _ = db.DB.ExecContext(r.Context(), `DELETE FROM otp_codes WHERE id = $1`, otpID)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "verification code has expired"})
		return
	}

	// Check attempt limit (max 5 attempts)
	if attempts >= 5 {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "too many failed attempts"})
		return
	}

	// Create user account with email from OTP record
	var user models.User
	userQuery := `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, balance, created_at, updated_at
	`
	err = db.DB.QueryRowContext(r.Context(), userQuery, name, req.Email, passwordHash).Scan(
		&user.ID, &user.Name, &user.Email, &user.Balance, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to create account"})
		return
	}

	// Mark OTP as verified
	_, _ = db.DB.ExecContext(r.Context(),
		`UPDATE otp_codes SET verified = true WHERE id = $1`, otpID)

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to generate token"})
		return
	}

	// Cache initial balance
	if Cache != nil {
		Cache.SetBalance(r.Context(), user.ID, user.Balance)
	}

	log.Printf("✅ User account created and verified: %s (ID: %s)", user.Email, user.ID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.AuthResponse{Token: token, User: user})
}

// ResendOTP godoc
// POST /resend-otp
// Resends OTP to the user's email
func ResendOTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "email is required"})
		return
	}

	// Find the most recent unverified OTP for this email
	var otpID, code, name string
	var expiresAt time.Time

	otpQuery := `
		SELECT id, code, name, expires_at
		FROM otp_codes
		WHERE email = $1 AND verified = false
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := db.DB.QueryRowContext(r.Context(), otpQuery, req.Email).Scan(
		&otpID, &code, &name, &expiresAt,
	)
	if err == sql.ErrNoRows {
		// Return success even if no OTP exists (security)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.OTPResponse{
			Message: "If an OTP exists, it has been resent to your email",
			Email:   req.Email,
		})
		return
	}
	if err != nil {
		log.Printf("Error fetching OTP: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	// Send OTP via email
	emailClient := email.NewSendGridClient()
	if emailClient != nil {
		go func() {
			subject := "PaymentApp Verification Code 🔐"
			htmlContent := utils.FormatOTPMessage(name, code)
			if err := emailClient.SendEmail(req.Email, subject, htmlContent); err != nil {
				log.Printf("❌ Error resending OTP to %s: %v", req.Email, err)
			} else {
				log.Printf("✅ OTP resent to %s", req.Email)
			}
		}()
	}

	log.Printf("📧 OTP resent to %s", req.Email)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.OTPResponse{
		Message: "Verification code has been resent to your email",
		Email:   req.Email,
	})
}
