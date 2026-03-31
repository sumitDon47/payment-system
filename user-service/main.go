package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/yourname/payment-system/user-service/db"
	"github.com/yourname/payment-system/user-service/handler"
	"github.com/yourname/payment-system/user-service/middleware"
)

func main() {
	// Load .env file (won't fail if not found — useful for Docker)
	_ = godotenv.Load()

	// Connect to PostgreSQL
	db.Connect()

	// Setup routes
	mux := http.NewServeMux()

	// Public routes — no auth needed
	mux.HandleFunc("/health", handler.HealthCheck)
	mux.HandleFunc("/register", handler.Register)
	mux.HandleFunc("/login", handler.Login)

	// Protected routes — require valid JWT
	mux.HandleFunc("/profile", middleware.AuthMiddleware(handler.GetProfile))
	mux.HandleFunc("/wallet", middleware.AuthMiddleware(handler.GetWalletBalance))

	// Wrap everything with CORS middleware
	server := middleware.CORSMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("User Service running on port %s", port)
	if err := http.ListenAndServe(":"+port, server); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
