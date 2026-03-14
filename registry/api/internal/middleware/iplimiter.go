package middleware

import (
	"net/http"
	"os"
	"strings"
)

// EndpointRateLimiter applies per-endpoint rate limits.
// Each endpoint group gets its own independent rate limiter.
type EndpointRateLimiter struct {
	limiter      *RateLimiter
	trustedProxy string
}

// NewEndpointRateLimiter creates a rate limiter with the given requests-per-second
// and burst capacity, scoped to a specific endpoint group.
func NewEndpointRateLimiter(rps, burst int) *EndpointRateLimiter {
	return &EndpointRateLimiter{
		limiter:      NewRateLimiter(rps, burst),
		trustedProxy: os.Getenv("TRUSTED_PROXY"),
	}
}

// Middleware returns an http middleware that enforces this rate limit.
func (e *EndpointRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r, e.trustedProxy)
		if !e.limiter.allow(key) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the client IP, only trusting X-Forwarded-For when behind
// a known trusted proxy.
func clientIP(r *http.Request, trustedProxy string) string {
	key := r.RemoteAddr
	if trustedProxy != "" && strings.HasPrefix(r.RemoteAddr, trustedProxy) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if idx := strings.IndexByte(xff, ','); idx > 0 {
				key = strings.TrimSpace(xff[:idx])
			} else {
				key = strings.TrimSpace(xff)
			}
		}
	}
	return key
}

// Common per-endpoint rate limiters.
var (
	// DeviceRateLimit: 5 requests per minute per IP for device auth flow.
	DeviceRateLimit = NewEndpointRateLimiter(5, 10)

	// UploadRateLimit: 10 uploads per minute per IP.
	UploadRateLimit = NewEndpointRateLimiter(10, 20)

	// SearchRateLimit: 60 requests per minute per IP.
	SearchRateLimit = NewEndpointRateLimiter(60, 120)
)
