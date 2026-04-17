package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/sumitDon47/payment-system/user-service/cache"
	"github.com/sumitDon47/payment-system/user-service/db"
	"github.com/sumitDon47/payment-system/user-service/handler"
	"github.com/sumitDon47/payment-system/user-service/middleware"
)

func main() {
	_ = godotenv.Load()

	// Connect to PostgreSQL
	db.Connect()

	// Connect to Redis — optional, service works without it
	// If REDIS_URL is not set or Redis is down, Cache stays nil
	// and all cache operations in handler/user.go are silently skipped
	redisClient := cache.New()
	if redisClient != nil {
		defer redisClient.Close()
		handler.Cache = redisClient
		log.Println("✅ Redis cache enabled")
	} else {
		log.Println("⚠️  Redis unavailable — running without cache (slower but correct)")
	}

	// Setup routes
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/health", handler.HealthCheck)
	mux.HandleFunc("/register", handler.Register)
	mux.HandleFunc("/login", handler.Login)

	// Protected routes
	mux.HandleFunc("/profile", middleware.AuthMiddleware(handler.GetProfile))
	mux.HandleFunc("/wallet", middleware.AuthMiddleware(handler.GetWalletBalance))

	// Internal route — cache invalidation called by payment-service
	// In production this would be behind internal network only
	mux.HandleFunc("/internal/cache/invalidate", handler.InvalidateUserCache)

	server := middleware.CORSMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 User Service running on port %s", port)
	if err := http.ListenAndServe(":"+port, server); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
