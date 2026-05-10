package middleware

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter stores per-IP rate limiters
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

// getLimiter gets or creates a rate limiter for an IP address
// Each IP gets its own rate limiter (per-IP rate limiting)
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		// Create new limiter: 10 requests per second with burst of 2
		// Tokens per second = 10, Burst = 2 (can send 2 tokens at once)
		limiter = rate.NewLimiter(rate.Limit(10), 2)
		rl.limiters[ip] = limiter
	}
	return limiter
}

// Allow checks if a request from the given IP is allowed
// Returns true if request is allowed, false if rate limit exceeded
func (rl *RateLimiter) Allow(ip string) bool {
	limiter := rl.getLimiter(ip)
	return limiter.Allow()
}

// GetClientIP extracts the client IP from the request
// Handles X-Forwarded-For header (for requests through proxy/load balancer)
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for requests through reverse proxy)
	// This is set by nginx, CloudFlare, etc.
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs: client, proxy1, proxy2
		// We want the first one (original client)
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Fall back to direct connection IP
	// RemoteAddr is in format "ip:port"
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, return the whole RemoteAddr
		return r.RemoteAddr
	}
	return ip
}

// Global rate limiters for different endpoints
var (
	// AuthRateLimiter limits /register and /login (5 requests/min per IP)
	AuthRateLimiter = NewRateLimiter()

	// ApiRateLimiter limits general API endpoints (100 requests/min per IP)
	ApiRateLimiter = NewRateLimiter()
)

// RateLimitMiddleware returns a middleware that enforces rate limits
// Use this to wrap sensitive endpoints like /register and /login
func RateLimitMiddleware(limiter *RateLimiter, limit int, burst int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := GetClientIP(r)

			// Create a custom limiter for this specific request
			// This allows different endpoints to have different limits
			if !limiter.Allow(clientIP) {
				log.Printf("⚠️  Rate limit exceeded for IP %s on %s", clientIP, r.URL.Path)
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitHandler wraps an HTTP handler with rate limiting
// Simpler version: just wraps a single handler
func RateLimitHandler(limiter *RateLimiter, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := GetClientIP(r)

		if !limiter.Allow(clientIP) {
			log.Printf("⚠️  Rate limit exceeded for IP %s on %s", clientIP, r.URL.Path)
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		handler(w, r)
	}
}

// ============================================================================
//  Specialized Rate Limiters for Different Endpoints
// ============================================================================

// AuthLimiter: 5 requests per minute per IP (for /register and /login)
type AuthLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

func NewAuthLimiter() *AuthLimiter {
	return &AuthLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (al *AuthLimiter) Allow(ip string) bool {
	al.mu.Lock()
	defer al.mu.Unlock()

	limiter, exists := al.limiters[ip]
	if !exists {
		// 5 requests per minute = 5/60 = 0.083 per second
		// Burst of 1 (can send 1 token at once)
		limiter = rate.NewLimiter(rate.Limit(5.0/60), 1)
		al.limiters[ip] = limiter
	}
	return limiter.Allow()
}

// ApiLimiter: 100 requests per minute per IP (for /profile, /wallet)
type ApiLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

func NewApiLimiter() *ApiLimiter {
	return &ApiLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (al *ApiLimiter) Allow(ip string) bool {
	al.mu.Lock()
	defer al.mu.Unlock()

	limiter, exists := al.limiters[ip]
	if !exists {
		// 100 requests per minute = 100/60 = 1.67 per second
		// Burst of 5 (can send 5 tokens at once)
		limiter = rate.NewLimiter(rate.Limit(100.0/60), 5)
		al.limiters[ip] = limiter
	}
	return limiter.Allow()
}

// Global instances
var (
	authLimiter = NewAuthLimiter()
	apiLimiter  = NewApiLimiter()
)

// LimitAuth returns a handler that rate limits at 5 requests/minute
func LimitAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := GetClientIP(r)
		if !authLimiter.Allow(clientIP) {
			log.Printf("⚠️  Auth rate limit exceeded for IP %s", clientIP)
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Too many authentication attempts. Please try again in 1 minute.", http.StatusTooManyRequests)
			return
		}
		handler(w, r)
	}
}

// LimitApi returns a handler that rate limits at 100 requests/minute
func LimitApi(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := GetClientIP(r)
		if !apiLimiter.Allow(clientIP) {
			log.Printf("⚠️  API rate limit exceeded for IP %s", clientIP)
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}
		handler(w, r)
	}
}
