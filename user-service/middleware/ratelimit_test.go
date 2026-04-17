package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetClientIP_Direct(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.100:54321"

	ip := GetClientIP(req)
	if ip != "192.168.1.100" {
		t.Errorf("Expected IP 192.168.1.100, got %s", ip)
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:54321"
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 70.41.3.18, 150.172.238.178")

	ip := GetClientIP(req)
	if ip != "203.0.113.5" {
		t.Errorf("Expected IP 203.0.113.5 from X-Forwarded-For, got %s", ip)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter()

	ip := "192.168.1.100"

	// First request should be allowed (initial burst)
	if !limiter.Allow(ip) {
		t.Error("Expected first request to be allowed")
	}

	// Should still have burst available
	if !limiter.Allow(ip) {
		t.Error("Expected second request (burst) to be allowed")
	}

	// After burst is exhausted, requests should be rate limited
	// Current limit is 10/sec, so after burst of 2, it should block
	allowed := 0
	for i := 0; i < 100; i++ {
		if limiter.Allow(ip) {
			allowed++
		}
	}

	// Should allow approximately 10 more requests per second
	// With a fast loop, we might get 0-2 more before hitting the limit
	if allowed > 10 {
		t.Errorf("Expected ~10 requests/sec, but allowed %d out of 100 rapid requests", allowed)
	}
}

func TestRateLimiter_PerIP(t *testing.T) {
	limiter := NewRateLimiter()

	ip1 := "192.168.1.100"
	ip2 := "192.168.1.101"

	// Both IPs should be able to make requests independently
	if !limiter.Allow(ip1) {
		t.Error("Expected request from IP1 to be allowed")
	}
	if !limiter.Allow(ip2) {
		t.Error("Expected request from IP2 to be allowed")
	}

	// Each IP should have its own rate limit
	limiter1 := limiter.getLimiter(ip1)
	limiter2 := limiter.getLimiter(ip2)

	if limiter1 == limiter2 {
		t.Error("Expected separate rate limiters for different IPs")
	}
}

func TestAuthLimiter_StrictLimit(t *testing.T) {
	limiter := NewAuthLimiter()
	ip := "192.168.1.100"

	// Auth limiter: 5 requests per minute
	// Should allow 1 immediately (burst)
	if !limiter.Allow(ip) {
		t.Error("Expected first auth request to be allowed (burst)")
	}

	// Subsequent requests within burst should work
	for i := 0; i < 10; i++ {
		if limiter.Allow(ip) {
			// We should hit the rate limit after the burst
			if i == 0 {
				// First additional request might still work due to precision
				continue
			}
			if i > 2 {
				t.Errorf("Auth limiter should be more restrictive, but allowed %d requests", i+2)
				break
			}
		}
	}
}

func TestApiLimiter_HigherLimit(t *testing.T) {
	limiter := NewApiLimiter()
	ip := "192.168.1.100"

	// API limiter: 100 requests per minute ≈ 1.67/sec
	allowed := 0
	for i := 0; i < 100; i++ {
		if limiter.Allow(ip) {
			allowed++
		}
	}

	// Should allow at least 5 (burst)
	if allowed < 5 {
		t.Errorf("API limiter burst should be at least 5, got %d", allowed)
	}

	// But shouldn't allow all 100 in a tight loop
	if allowed > 90 {
		t.Errorf("API limiter should still enforce rate limit, allowed %d out of 100", allowed)
	}
}

func TestLimitAuth_RateLimitExceeded(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	limitedHandler := LimitAuth(handler)

	ip := "192.168.1.100"
	req := httptest.NewRequest("POST", "/register", nil)
	req.RemoteAddr = ip + ":54321"

	// First request should succeed
	rec := httptest.NewRecorder()
	limitedHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected first request to return 200, got %d", rec.Code)
	}

	// Hammer with many requests to exceed rate limit
	blockedCount := 0
	var lastRec *httptest.ResponseRecorder
	for i := 0; i < 20; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/register", nil)
		req.RemoteAddr = ip + ":54321"
		limitedHandler(rec, req)
		lastRec = rec

		if rec.Code == http.StatusTooManyRequests {
			blockedCount++
		}
	}

	if blockedCount == 0 {
		t.Error("Expected some requests to be rate limited")
	}

	// Check the last response had the Retry-After header
	if lastRec != nil && lastRec.Code == http.StatusTooManyRequests {
		if lastRec.Header().Get("Retry-After") != "60" {
			t.Logf("Warning: Retry-After header is '%s', expected '60'", lastRec.Header().Get("Retry-After"))
		}
	}
}

func TestLimitApi_HigherLimit(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	limitedHandler := LimitApi(handler)
	ip := "192.168.1.200" // Use different IP to avoid interference with other tests

	// Should allow requests
	successCount := 0
	for i := 0; i < 10; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/profile", nil)
		req.RemoteAddr = ip + ":54321"
		limitedHandler(rec, req)

		if rec.Code == http.StatusOK {
			successCount++
		}
	}

	// API limiter should allow at least some requests
	if successCount == 0 {
		t.Errorf("API limiter should allow requests, allowed %d out of 10", successCount)
	}
}

func TestRateLimit_ConcurrentRequests(t *testing.T) {
	limiter := NewRateLimiter()
	ip := "192.168.1.100"

	// Test concurrent access doesn't panic or cause race conditions
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				limiter.Allow(ip)
				time.Sleep(time.Microsecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without a race condition, test passed
	t.Log("Concurrent rate limiting test passed")
}

func TestRateLimit_DifferentEndpoints(t *testing.T) {
	// Auth limiter should be stricter
	authLimiter := NewAuthLimiter()
	apiLimiter := NewApiLimiter()

	ip := "192.168.1.100"

	// Rapid fire requests
	authAllowed := 0
	apiAllowed := 0

	for i := 0; i < 50; i++ {
		if authLimiter.Allow(ip) {
			authAllowed++
		}
		if apiLimiter.Allow(ip) {
			apiAllowed++
		}
	}

	// API limiter should allow more requests
	if authAllowed >= apiAllowed {
		t.Errorf("Expected API limiter (%d) to allow more than auth (%d)", apiAllowed, authAllowed)
	}
}
