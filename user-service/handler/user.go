package handler

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sumitDon47/payment-system/user-service/cache"
	"github.com/sumitDon47/payment-system/user-service/db"
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

	var user models.User
	var hashedPassword string
	query := `SELECT id, name, email, password, balance, created_at, updated_at FROM users WHERE email = $1`
	err := db.DB.QueryRowContext(r.Context(), query, req.Email).Scan(
		&user.ID, &user.Name, &user.Email, &hashedPassword,
		&user.Balance, &user.CreatedAt, &user.UpdatedAt,
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
