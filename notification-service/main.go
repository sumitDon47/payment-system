package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sumitDon47/payment-system/notification-service/consumer"
	"github.com/sumitDon47/payment-system/notification-service/db"
)

func main() {
	_ = godotenv.Load()

	db.Connect()

	broker := getEnv("KAFKA_BROKER", "kafka:9092")
	topic := getEnv("KAFKA_TOPIC", "payment.completed")
	groupID := getEnv("KAFKA_GROUP_ID", "notification-service")

	log.Printf("🚀 Notification Service starting")
	log.Printf("   Broker:  %s", broker)
	log.Printf("   Topic:   %s", topic)
	log.Printf("   GroupID: %s", groupID)

	// Create consumer
	c := consumer.New(broker, topic, groupID)
	defer c.Close()

	// Context that cancels on OS signal — this is graceful shutdown.
	// When you press Ctrl+C or Docker stops the container, SIGTERM is sent.
	// The context cancels, FetchMessage returns, the loop exits cleanly.
	// No message is lost or left half-processed.
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	// Start consuming — this blocks until ctx is cancelled
	// Run in a goroutine so main can wait for the signal below
	go c.Start(ctx)

	// Start a separate HTTP server for Prometheus metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("📊 Prometheus metrics (HTTP) available on :8082/metrics")
		if err := http.ListenAndServe(":8082", nil); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start metrics server: %v", err)
		}
	}()

	// Block here until signal received
	<-ctx.Done()
	log.Println("🛑 Shutdown signal received — stopping gracefully")
}

// getEnv reads an environment variable with a fallback default.
// Same pattern used in both other services — now you know why it exists.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
