package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/sumitDon47/payment-system/user-service/db"
	"github.com/sumitDon47/payment-system/user-service/models"
)

// LookupUserByEmail godoc
// GET /users/lookup?email=example@email.com
// Allows authenticated users to look up another user's ID by email
// Requires: Authorization header with JWT token
func LookupUserByEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "method not allowed"})
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "email query parameter required"})
		return
	}

	var user models.User
	query := `SELECT id, name, email FROM users WHERE email = $1`
	err := db.DB.QueryRowContext(r.Context(), query, email).Scan(
		&user.ID, &user.Name, &user.Email,
	)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "user not found"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	json.NewEncoder(w).Encode(models.SuccessResponse{
		Message: "user found",
		Data:    map[string]string{"id": user.ID, "name": user.Name, "email": user.Email},
	})
}
