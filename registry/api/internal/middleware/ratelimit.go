package middleware

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// tokenBucket implements a simple token bucket per key.
type tokenBucket struct {
	tokens float64
	last   time.Time
	rate   float64
	burst  int
}

// RateLimiter limits requests per key (e.g. IP or API key).
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	rate    float64 // tokens per second
	burst   int
	now     func() time.Time
}

// NewRateLimiter creates a limiter with rps tokens per second and burst capacity.
func NewRateLimiter(rps int, burst int) *RateLimiter {
	if burst <= 0 {
		burst = rps * 2
	}
	return &RateLimiter{
		buckets: make(map[string]*tokenBucket),
		rate:    float64(rps),
		burst:   burst,
		now:     time.Now,
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := rl.now()
	b, ok := rl.buckets[key]
	if !ok {
		b = &tokenBucket{tokens: float64(rl.burst), rate: rl.rate, burst: rl.burst}
		rl.buckets[key] = b
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * b.rate
	if b.tokens > float64(b.burst) {
		b.tokens = float64(b.burst)
	}
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// RateLimit returns middleware that limits by request RemoteAddr (or X-Forwarded-For
// if TRUSTED_PROXY is set and matches the direct peer).
func RateLimit(rps int) func(next http.Handler) http.Handler {
	limiter := NewRateLimiter(rps, rps*2)
	trustedProxy := os.Getenv("TRUSTED_PROXY")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.RemoteAddr
			// M8: Only trust X-Forwarded-For when behind a known trusted proxy.
			if trustedProxy != "" && strings.HasPrefix(r.RemoteAddr, trustedProxy) {
				if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
					// Use the first (leftmost) IP in the chain.
					if idx := strings.IndexByte(xff, ','); idx > 0 {
						key = strings.TrimSpace(xff[:idx])
					} else {
						key = strings.TrimSpace(xff)
					}
				}
			}
			if !limiter.allow(key) {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
