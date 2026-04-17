package middleware

import (
	"context"
	"log"
	"sync"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RateLimiterInterceptor stores per-user rate limiters for gRPC
type RateLimiterInterceptor struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

// NewRateLimiterInterceptor creates a new gRPC rate limiter
func NewRateLimiterInterceptor() *RateLimiterInterceptor {
	return &RateLimiterInterceptor{
		limiters: make(map[string]*rate.Limiter),
	}
}

// getLimiter gets or creates a rate limiter for a user ID
// Each user gets their own rate limiter (per-user rate limiting)
func (rli *RateLimiterInterceptor) getLimiter(userID string) *rate.Limiter {
	rli.mu.Lock()
	defer rli.mu.Unlock()

	limiter, exists := rli.limiters[userID]
	if !exists {
		// Create new limiter: 100 requests per second with burst of 5
		// For SendPayment, this is about 360,000 payments per hour per user
		limiter = rate.NewLimiter(rate.Limit(100), 5)
		rli.limiters[userID] = limiter
	}
	return limiter
}

// Allow checks if a request from the given user ID is allowed
func (rli *RateLimiterInterceptor) Allow(userID string) bool {
	limiter := rli.getLimiter(userID)
	return limiter.Allow()
}

// ============================================================================
//  gRPC Unary Interceptor for Rate Limiting
// ============================================================================

// UnaryServerInterceptor returns a grpc.UnaryServerInterceptor that enforces rate limits
// Extract user ID from request context and apply per-user rate limiting
func UnaryServerInterceptor(limiter *RateLimiterInterceptor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Rate limit based on method name (you could also extract user ID from request)
		// For now, we use method name as the identifier
		method := info.FullMethod

		// Extract user ID from context if available (set by auth middleware)
		// Otherwise use the method name as identifier
		identifier := method
		if userID := extractUserID(ctx); userID != "" {
			identifier = userID
		}

		// Check rate limit
		if !limiter.Allow(identifier) {
			log.Printf("⚠️  gRPC rate limit exceeded for %s on method %s", identifier, method)
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded: too many requests")
		}

		// Call the handler
		return handler(ctx, req)
	}
}

// extractUserID extracts the user ID from the request context
// This assumes the user ID is set by an auth interceptor
func extractUserID(ctx context.Context) string {
	// Check for user ID in context metadata
	userID, ok := ctx.Value("user_id").(string)
	if ok && userID != "" {
		return userID
	}
	return ""
}

// ============================================================================
//  Specialized Rate Limiters
// ============================================================================

// SendPaymentLimiter: 100 requests per second per user
type SendPaymentLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

func NewSendPaymentLimiter() *SendPaymentLimiter {
	return &SendPaymentLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (spl *SendPaymentLimiter) Allow(userID string) bool {
	spl.mu.Lock()
	defer spl.mu.Unlock()

	limiter, exists := spl.limiters[userID]
	if !exists {
		// 100 requests per second, burst of 5
		limiter = rate.NewLimiter(100, 5)
		spl.limiters[userID] = limiter
	}
	return limiter.Allow()
}

// GetTransactionLimiter: 1000 requests per minute per user
type GetTransactionLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

func NewGetTransactionLimiter() *GetTransactionLimiter {
	return &GetTransactionLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (gtl *GetTransactionLimiter) Allow(userID string) bool {
	gtl.mu.Lock()
	defer gtl.mu.Unlock()

	limiter, exists := gtl.limiters[userID]
	if !exists {
		// 1000 requests per minute = 16.67 per second
		// Burst of 10
		limiter = rate.NewLimiter(rate.Limit(1000.0/60), 10)
		gtl.limiters[userID] = limiter
	}
	return limiter.Allow()
}

// Global instances
var (
	sendPaymentLimiter    = NewSendPaymentLimiter()
	getTransactionLimiter = NewGetTransactionLimiter()
	generalRateLimiter    = NewRateLimiterInterceptor()
)

// RateLimitSendPayment checks if user can send a payment
func RateLimitSendPayment(userID string) bool {
	return sendPaymentLimiter.Allow(userID)
}

// RateLimitGetTransaction checks if user can query transactions
func RateLimitGetTransaction(userID string) bool {
	return getTransactionLimiter.Allow(userID)
}

// RateLimitGeneralAPI checks if user/request can proceed
func RateLimitGeneralAPI(identifier string) bool {
	return generalRateLimiter.Allow(identifier)
}
