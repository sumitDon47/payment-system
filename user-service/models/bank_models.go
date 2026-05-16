package models

import "time"

// Bank-related models for wallet loading

type BankWalletLoadRequest struct {
	PhoneNumber   string  `json:"phone_number"`
	Amount        float64 `json:"amount"`
	BankReference string  `json:"bank_reference"`
	BankCode      string  `json:"bank_code"`
	Description   string  `json:"description,omitempty"`
}

type BankWalletLoadResponse struct {
	Status           string    `json:"status"`
	TransactionID    string    `json:"transaction_id"`
	PhoneNumber      string    `json:"phone_number"`
	Amount           float64   `json:"amount"`
	WalletBalance    float64   `json:"wallet_balance"`
	BankReference    string    `json:"bank_reference"`
	Timestamp        time.Time `json:"timestamp"`
}

type BankWalletVerificationRequest struct {
	BankReference string `json:"bank_reference"`
	Status        string `json:"status"` // completed, failed
	Timestamp     time.Time `json:"timestamp"`
	Signature     string `json:"signature"`
}

type BankWalletStatusResponse struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	Amount        float64   `json:"amount"`
	PhoneNumber   string    `json:"phone_number"`
	BankReference string    `json:"bank_reference"`
	CreatedAt     time.Time `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

type BankWalletFailureRequest struct {
	BankReference string `json:"bank_reference"`
	Status        string `json:"status"` // failed
	Reason        string `json:"reason"`
	Timestamp     time.Time `json:"timestamp"`
}

type BankAPIError struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// SendPaymentByPhoneRequest for gRPC
type SendPaymentByPhoneRequest struct {
	SenderPhone   string  `json:"sender_phone"`
	ReceiverPhone string  `json:"receiver_phone"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Description   string  `json:"description,omitempty"`
}

type SendPaymentByPhoneResponse struct {
	TransactionID string  `json:"transaction_id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	Message       string  `json:"message"`
	Timestamp     time.Time `json:"timestamp"`
}

// BankWalletLoad database model
type BankWalletLoad struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	PhoneNumber   string     `json:"phone_number"`
	Amount        float64    `json:"amount"`
	BankReference string     `json:"bank_reference"`
	BankCode      string     `json:"bank_code"`
	Status        string     `json:"status"` // pending, completed, failed, reversed
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}
