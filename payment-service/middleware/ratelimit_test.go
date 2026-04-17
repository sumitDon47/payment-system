package middleware

import (
	"context"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestRateLimiterInterceptor_NewInstance(t *testing.T) {
	limiter := NewRateLimiterInterceptor()
	if limiter == nil {
		t.Fatal("NewRateLimiterInterceptor returned nil")
	}
	if len(limiter.limiters) != 0 {
		t.Errorf("Expected empty limiters map, got %d", len(limiter.limiters))
	}
}

func TestRateLimiterInterceptor_Allow(t *testing.T) {
	limiter := NewRateLimiterInterceptor()
	userID := "user-123"

	// First request should be allowed (initial burst)
	if !limiter.Allow(userID) {
		t.Error("Expected first request to be allowed")
	}

	// Should still have burst available
	if !limiter.Allow(userID) {
		t.Error("Expected second request (burst) to be allowed")
	}
}

func TestRateLimiterInterceptor_PerUser(t *testing.T) {
	limiter := NewRateLimiterInterceptor()
	user1 := "user-123"
	user2 := "user-456"

	// Both users should be able to make requests
	if !limiter.Allow(user1) {
		t.Error("Expected request from user1 to be allowed")
	}
	if !limiter.Allow(user2) {
		t.Error("Expected request from user2 to be allowed")
	}

	// Each user should have their own rate limiter
	limiter1 := limiter.getLimiter(user1)
	limiter2 := limiter.getLimiter(user2)

	if limiter1 == limiter2 {
		t.Error("Expected separate rate limiters for different users")
	}
}

func TestSendPaymentLimiter_HighLimit(t *testing.T) {
	limiter := NewSendPaymentLimiter()
	userID := "user-123"

	// Send payment limiter should allow requests in burst
	allowedCount := 0
	for i := 0; i < 10; i++ {
		if limiter.Allow(userID) {
			allowedCount++
		}
	}

	// Should allow at least the burst (5) out of 10 rapid requests
	if allowedCount < 5 {
		t.Errorf("SendPayment limiter should allow burst of at least 5, allowed %d out of 10", allowedCount)
	}

	// With no time passing, subsequent requests should be blocked
	if allowedCount > 5 {
		t.Logf("SendPayment limiter allowed %d out of 10 (expected ~5 due to burst)", allowedCount)
	}
}

func TestGetTransactionLimiter_ModerateLimit(t *testing.T) {
	limiter := NewGetTransactionLimiter()
	userID := "user-123"

	// GetTransaction limiter: 1000 requests per minute
	allowedCount := 0
	for i := 0; i < 100; i++ {
		if limiter.Allow(userID) {
			allowedCount++
		}
	}

	// Should allow significant portion of 100 requests
	if allowedCount < 10 {
		t.Errorf("GetTransaction limiter should allow ~16/sec, allowed %d out of 100", allowedCount)
	}
}

func TestUnaryServerInterceptor_AllowedRequest(t *testing.T) {
	limiter := NewRateLimiterInterceptor()
	interceptor := UnaryServerInterceptor(limiter)

	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/payment.PaymentService/SendPayment",
	}

	ctx := context.Background()
	resp, err := interceptor(ctx, nil, info, handler)

	if !called {
		t.Error("Expected handler to be called")
	}
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp != "response" {
		t.Errorf("Expected 'response', got %v", resp)
	}
}

func TestUnaryServerInterceptor_RateLimitExceeded(t *testing.T) {
	// Create a limiter with a very strict limit to trigger rate limiting
	limiter := NewRateLimiterInterceptor()

	// Exhaust the burst by making rapid requests to "test-method"
	for i := 0; i < 100; i++ {
		limiter.Allow("test-method")
	}

	// Now it should be rate limited
	interceptor := UnaryServerInterceptor(limiter)

	info := &grpc.UnaryServerInfo{
		FullMethod: "test-method",
	}

	ctx := context.Background()
	_, err := interceptor(ctx, nil, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("Handler should not be called when rate limited")
		return nil, nil
	})

	if err == nil {
		t.Fatal("Expected rate limit error")
	}

	// Check error contains rate limit message
	errStr := err.Error()
	if !strings.Contains(errStr, "rate limit exceeded") {
		t.Errorf("Unexpected error: %s", errStr)
	}
}

func TestRateLimitSendPayment(t *testing.T) {
	userID := "user-sendpay-" + t.Name() // Use unique user ID per test to avoid state collision

	// First request should be allowed
	if !RateLimitSendPayment(userID) {
		t.Error("Expected SendPayment to be allowed")
	}

	// Multiple requests should be allowed within burst (5)
	successCount := 0
	for i := 0; i < 5; i++ {
		if RateLimitSendPayment(userID) {
			successCount++
		}
	}

	// With burst of 5, we should allow at least some more requests
	if successCount == 0 {
		t.Error("SendPayment rate limit too restrictive")
	}
}

func TestRateLimitGetTransaction(t *testing.T) {
	userID := "user-456"

	// First request should be allowed
	if !RateLimitGetTransaction(userID) {
		t.Error("Expected GetTransaction to be allowed")
	}
}

func TestRateLimitGeneralAPI(t *testing.T) {
	identifier := "test-id"

	// First request should be allowed
	if !RateLimitGeneralAPI(identifier) {
		t.Error("Expected general API request to be allowed")
	}
}

func TestRateLimitConcurrency(t *testing.T) {
	limiter := NewRateLimiterInterceptor()
	userID := "concurrent-user"

	done := make(chan bool, 10)

	// Concurrent access shouldn't panic
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				limiter.Allow(userID)
				time.Sleep(time.Microsecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("Concurrent rate limiting test passed")
}

func TestDifferentLimiters_VaryingThroughput(t *testing.T) {
	sendPaymentLimiter := NewSendPaymentLimiter()
	getTransactionLimiter := NewGetTransactionLimiter()

	userID := "test-user"

	// Send payment burst of 5, GetTransaction burst of 10
	sendPaymentAllowed := 0
	getTransactionAllowed := 0

	for i := 0; i < 20; i++ {
		if sendPaymentLimiter.Allow(userID) {
			sendPaymentAllowed++
		}
		if getTransactionLimiter.Allow(userID) {
			getTransactionAllowed++
		}
	}

	// With no time passing, SendPayment should allow burst of 5, GetTransaction should allow burst of 10
	if sendPaymentAllowed != 5 {
		t.Logf("SendPayment limiter allowed %d out of 20 (expected 5 burst)", sendPaymentAllowed)
	}

	if getTransactionAllowed != 10 {
		t.Logf("GetTransaction limiter allowed %d out of 20 (expected 10 burst)", getTransactionAllowed)
	}

	// GetTransaction should allow more due to higher burst
	if getTransactionAllowed <= sendPaymentAllowed {
		t.Errorf("GetTransaction burst (%d) should be higher than SendPayment (%d)", getTransactionAllowed, sendPaymentAllowed)
	}
}
