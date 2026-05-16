package models

import "time"

type User struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Password      string    `json:"-"` // never send password in JSON response
	PhoneNumber   string    `json:"phone_number,omitempty"`
	PhoneVerified bool      `json:"phone_verified"`
	Balance       float64   `json:"balance"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	MPIN        string `json:"mpin"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	MPIN     string `json:"mpin"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type RegisterWithOTPRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	MPIN        string `json:"mpin"`
}

type VerifyOTPRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type OTPResponse struct {
	Message string `json:"message"`
	Email   string `json:"email"`
}

type OTPCode struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Code         string    `json:"-"` // never send code in response
	Name         string    `json:"-"`
	PasswordHash string    `json:"-"`
	ExpiresAt    time.Time `json:"expires_at"`
	Attempts     int       `json:"attempts"`
	Verified     bool      `json:"verified"`
	CreatedAt    time.Time `json:"created_at"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PhoneLoginRequest struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	MPIN        string `json:"mpin"`
}

type PhoneLookupResponse struct {
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
}

type SendPhoneOTPRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type VerifyPhoneOTPRequest struct {
	PhoneNumber string `json:"phone_number"`
	Code        string `json:"code"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type WalletResponse struct {
	Balance float64 `json:"balance"`
	UserID  string  `json:"user_id"`
}

type CacheInvalidationRequest struct {
	UserID string `json:"user_id"`
}
