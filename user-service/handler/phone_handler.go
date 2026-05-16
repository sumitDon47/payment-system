package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/sumitDon47/payment-system/user-service/db"
	"github.com/sumitDon47/payment-system/user-service/models"
	"github.com/sumitDon47/payment-system/user-service/utils"
	"golang.org/x/crypto/bcrypt"
)

// Phone number validation (E.164 format: +{country_code}{number})
var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

func validatePhoneNumber(phone string) error {
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("invalid phone number format, use E.164 format: +{country_code}{number}")
	}
	return nil
}

// LoginByPhone godoc
// POST /login/phone
// Authenticate user with phone number and password
func LoginByPhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req models.PhoneLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request body"})
		return
	}

	// Validate phone number format
	if err := validatePhoneNumber(req.PhoneNumber); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()})
		return
	}

	if req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "password is required"})
		return
	}

	// Query user by phone number
	var user models.User
	var passwordHash string
	err := db.DB.QueryRowContext(r.Context(),
		`SELECT id, name, email, phone_number, balance, password, created_at, updated_at 
		 FROM users WHERE phone_number = $1`,
		req.PhoneNumber,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PhoneNumber, &user.Balance, &passwordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid phone number or password"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "database error"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid phone number or password"})
		return
	}

	// Verify MPIN if provided
	if req.MPIN != "" {
		if err := validateMPIN(req.MPIN); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()})
			return
		}
		// Additional MPIN verification logic here if needed
	}

	// Generate JWT token
	tokenString, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to generate token"})
		return
	}

	// Clear cache after login
	if Cache != nil {
		Cache.Delete(r.Context(), "user:"+user.ID)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.LoginResponse{
		Token: tokenString,
		User:  user,
	})
}

// LookupUserByPhone godoc
// GET /lookup/phone/:phone_number
// Look up user by phone number (basic info only)
func LookupUserByPhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "method not allowed"})
		return
	}

	// Extract phone number from URL parameter
	phoneNumber := r.URL.Query().Get("phone")
	if phoneNumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "phone number parameter is required"})
		return
	}

	// Validate phone number format
	if err := validatePhoneNumber(phoneNumber); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()})
		return
	}

	// Try cache first
	cacheKey := "phone:lookup:" + phoneNumber
	if Cache != nil {
		if cached, err := Cache.Get(r.Context(), cacheKey); err == nil && cached != "" {
			var resp models.PhoneLookupResponse
			if json.Unmarshal([]byte(cached), &resp) == nil {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
				return
			}
		}
	}

	// Query user by phone number
	var userID, name string
	err := db.DB.QueryRowContext(r.Context(),
		`SELECT id, name FROM users WHERE phone_number = $1 AND phone_verified = true`,
		phoneNumber,
	).Scan(&userID, &name)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "user not found"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "database error"})
		return
	}

	resp := models.PhoneLookupResponse{
		UserID:      userID,
		Name:        name,
		PhoneNumber: phoneNumber,
	}

	// Cache the result for 5 minutes
	if Cache != nil {
		if data, err := json.Marshal(resp); err == nil {
			Cache.Set(r.Context(), cacheKey, string(data), 5*time.Minute)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// SendPhoneOTP godoc
// POST /send-phone-otp
// Send OTP to phone number
func SendPhoneOTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req models.SendPhoneOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := validatePhoneNumber(req.PhoneNumber); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()})
		return
	}

	// Check rate limiting: 3 attempts per 15 minutes
	rateLimitKey := "otp:rate:" + req.PhoneNumber
	if Cache != nil {
		attempts, _ := Cache.Get(r.Context(), rateLimitKey)
		attemptCount := 0
		if attempts != "" {
			fmt.Sscanf(attempts, "%d", &attemptCount)
		}
		if attemptCount >= 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "too many OTP requests, try again later"})
			return
		}
	}

	// Generate 6-digit OTP
	otp, err := utils.GenerateOTP()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to generate OTP"})
		return
	}

	// Store OTP in database (expires in 10 minutes)
	_, err = db.DB.ExecContext(r.Context(),
		`INSERT INTO phone_otps (phone_number, code, expires_at, attempts, created_at)
		 VALUES ($1, $2, $3, 0, NOW())`,
		req.PhoneNumber, otp, time.Now().Add(10*time.Minute),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to send OTP"})
		return
	}

	// TODO: Send OTP via SMS (Twilio, Sparrow, Nexmo, etc.)
	// For now, log it for testing
	fmt.Printf("OTP for %s: %s\n", req.PhoneNumber, otp)

	// Update rate limiting counter
	if Cache != nil {
		if attempts, _ := Cache.Get(r.Context(), rateLimitKey); attempts != "" {
			var count int
			fmt.Sscanf(attempts, "%d", &count)
			Cache.Set(r.Context(), rateLimitKey, fmt.Sprintf("%d", count+1), 15*time.Minute)
		} else {
			Cache.Set(r.Context(), rateLimitKey, "1", 15*time.Minute)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":      "OTP sent to phone number",
		"phone_number": req.PhoneNumber,
	})
}

// VerifyPhoneOTP godoc
// POST /verify-phone-otp
// Verify OTP sent to phone number
func VerifyPhoneOTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req models.VerifyPhoneOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := validatePhoneNumber(req.PhoneNumber); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()})
		return
	}

	if req.Code == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "OTP code is required"})
		return
	}

	// Verify OTP
	var verified bool
	err := db.DB.QueryRowContext(r.Context(),
		`SELECT 1 FROM phone_otps 
		 WHERE phone_number = $1 AND code = $2 AND expires_at > NOW()`,
		req.PhoneNumber, req.Code,
	).Scan(&verified)

	if err == sql.ErrNoRows {
		// Increment attempt counter
		_, _ = db.DB.ExecContext(r.Context(),
			`UPDATE phone_otps SET attempts = attempts + 1 
			 WHERE phone_number = $1 AND code = $2`,
			req.PhoneNumber, req.Code,
		)

		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid or expired OTP"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "verification failed"})
		return
	}

	// Update phone verification status for existing user
	_, err = db.DB.ExecContext(r.Context(),
		`UPDATE users SET phone_verified = true WHERE phone_number = $1`,
		req.PhoneNumber,
	)

	// Delete used OTP
	_, _ = db.DB.ExecContext(r.Context(),
		`DELETE FROM phone_otps WHERE phone_number = $1`,
		req.PhoneNumber,
	)

	// Clear rate limiting
	if Cache != nil {
		Cache.Delete(r.Context(), "otp:rate:"+req.PhoneNumber)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":      "Phone number verified successfully",
		"phone_number": req.PhoneNumber,
	})
}
