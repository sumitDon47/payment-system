package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/sumitDon47/payment-system/user-service/db"
	"github.com/sumitDon47/payment-system/user-service/models"
	"github.com/sumitDon47/payment-system/user-service/utils"
	"golang.org/x/crypto/bcrypt"
)

// Register godoc
// POST /register
// Body: { "name": "John", "email": "john@gmail.com", "password": "secret123" }
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

	// Basic validation
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

	// Hash the password — NEVER store plain text passwords
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to process password"})
		return
	}

	// Insert user into DB
	var user models.User
	query := `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, balance, created_at, updated_at
	`
	err = db.DB.QueryRow(query, req.Name, req.Email, string(hashedPassword)).Scan(
		&user.ID, &user.Name, &user.Email, &user.Balance, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		log.Printf("Register error: %v", err)
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "email already exists"})
		return
	}

	// Generate JWT
	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to generate token"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.AuthResponse{Token: token, User: user})
}

// Login godoc
// POST /login
// Body: { "email": "john@gmail.com", "password": "secret123" }
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

	// Find user by email
	var user models.User
	var hashedPassword string
	query := `SELECT id, name, email, password, balance, created_at, updated_at FROM users WHERE email = $1`
	err := db.DB.QueryRow(query, req.Email).Scan(
		&user.ID, &user.Name, &user.Email, &hashedPassword, &user.Balance, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid email or password"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid email or password"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to generate token"})
		return
	}

	json.NewEncoder(w).Encode(models.AuthResponse{Token: token, User: user})
}

// GetProfile godoc
// GET /profile  (requires Authorization: Bearer <token>)
func GetProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Header.Get("X-User-ID")

	var user models.User
	query := `SELECT id, name, email, balance, created_at, updated_at FROM users WHERE id = $1`
	err := db.DB.QueryRow(query, userID).Scan(
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
func GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Header.Get("X-User-ID")

	var balance float64
	err := db.DB.QueryRow(`SELECT balance FROM users WHERE id = $1`, userID).Scan(&balance)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "could not fetch balance"})
		return
	}

	json.NewEncoder(w).Encode(models.SuccessResponse{
		Message: "wallet balance fetched",
		Data:    map[string]float64{"balance": balance},
	})
}

// HealthCheck godoc
// GET /health
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "user-service"})
}
