package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

// ── CORS ──────────────────────────────────────────────────────────────────

func TestCORSWithAllowlist_AllowedOrigin(t *testing.T) {
	handler := CORSWithAllowlist("http://localhost:3000")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS, got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("Allow-Origin = %q, want %q", got, "http://localhost:3000")
	}
	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Allow-Credentials = %q, want %q", got, "true")
	}
}

func TestCORSWithAllowlist_DisallowedOrigin(t *testing.T) {
	handler := CORSWithAllowlist("http://localhost:3000")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin should be absent for disallowed origin, got %q", got)
	}
}

func TestCORSWithAllowlist_WildcardBlocked(t *testing.T) {
	handler := CORSWithAllowlist("http://localhost:3000")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://anything.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got == "*" {
		t.Error("CORS should never set Access-Control-Allow-Origin: *")
	}
}

func TestCORSWithAllowlist_MultipleOrigins(t *testing.T) {
	handler := CORSWithAllowlist("http://localhost:3000", "https://registry.hivemind.rithul.dev")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		origin  string
		allowed bool
	}{
		{"http://localhost:3000", true},
		{"https://registry.hivemind.rithul.dev", true},
		{"http://evil.com", false},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", tt.origin)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		got := rr.Header().Get("Access-Control-Allow-Origin")
		if tt.allowed && got != tt.origin {
			t.Errorf("origin %q should be allowed, got Allow-Origin=%q", tt.origin, got)
		}
		if !tt.allowed && got != "" {
			t.Errorf("origin %q should be rejected, got Allow-Origin=%q", tt.origin, got)
		}
	}
}

func TestCORSWithAllowlist_NormalizesTrailingSlash(t *testing.T) {
	// Origin with trailing slash should be normalized
	handler := CORSWithAllowlist("http://localhost:3000/")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("Allow-Origin = %q, want %q (trailing slash should be normalized)", got, "http://localhost:3000")
	}
}

func TestCORSWithAllowlist_NoOriginHeader(t *testing.T) {
	handler := CORSWithAllowlist("http://localhost:3000")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	// No Origin header
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin should be absent when no Origin sent, got %q", got)
	}
	// Should still set methods/headers
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("Allow-Methods should always be set")
	}
}

func TestCORSWithAllowlist_PassesThroughNonOptions(t *testing.T) {
	called := false
	handler := CORSWithAllowlist("http://localhost:3000")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("next handler should be called for non-OPTIONS requests")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// ── Rate Limiter ──────────────────────────────────────────────────────────

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(10, 5)
	for i := 0; i < 5; i++ {
		if !rl.allow("test-ip") {
			t.Fatalf("request %d should be allowed (under burst)", i)
		}
	}
}

func TestRateLimiter_BlocksWhenExceeded(t *testing.T) {
	rl := NewRateLimiter(1, 1)
	// First request consumes the one burst token
	if !rl.allow("test-ip") {
		t.Fatal("first request should be allowed")
	}
	// Second request should be blocked (no tokens left, rate too slow to refill immediately)
	if rl.allow("test-ip") {
		t.Fatal("second request should be blocked")
	}
}

func TestRateLimiter_SeparateIPsIndependent(t *testing.T) {
	rl := NewRateLimiter(1, 1)
	if !rl.allow("1.2.3.4") {
		t.Fatal("first IP should be allowed")
	}
	if !rl.allow("5.6.7.8") {
		t.Fatal("second IP should be allowed (different bucket)")
	}
}

func TestRateLimiter_RefillsOverTime(t *testing.T) {
	// Use a controllable time source
	now := time.Now()
	rl := NewRateLimiter(10, 1)
	rl.now = func() time.Time { return now }

	if !rl.allow("test-ip") {
		t.Fatal("first request should be allowed")
	}
	if rl.allow("test-ip") {
		t.Fatal("second request should be blocked (no tokens)")
	}

	// Advance time by 200ms → should refill 2 tokens at 10/s
	now = now.Add(200 * time.Millisecond)
	if !rl.allow("test-ip") {
		t.Fatal("request after time advance should be allowed")
	}
}

func TestEndpointRateLimiter_Middleware_AllowsUnderLimit(t *testing.T) {
	erl := NewEndpointRateLimiter(100, 10)
	handler := erl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, rr.Code)
		}
	}
}

func TestEndpointRateLimiter_Middleware_Returns429(t *testing.T) {
	erl := NewEndpointRateLimiter(1, 1)
	handler := erl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First: allowed
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", rr.Code)
	}

	// Second: rate limited
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", rr.Code)
	}
}

func TestClientIP_WithoutTrustedProxy(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")

	// No trusted proxy → should use RemoteAddr, not XFF
	ip := clientIP(req, "")
	if ip != "127.0.0.1:12345" {
		t.Errorf("clientIP without trusted proxy = %q, want %q", ip, "127.0.0.1:12345")
	}
}

func TestClientIP_WithTrustedProxy(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.50")

	ip := clientIP(req, "10.0.0.1")
	if ip != "203.0.113.50" {
		t.Errorf("clientIP with trusted proxy = %q, want %q", ip, "203.0.113.50")
	}
}

func TestClientIP_WithTrustedProxy_MultipleXFF(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 10.0.0.2, 10.0.0.1")

	ip := clientIP(req, "10.0.0.1")
	if ip != "203.0.113.50" {
		t.Errorf("clientIP with multiple XFF = %q, want first IP %q", ip, "203.0.113.50")
	}
}

func TestClientIP_UntrustedProxyIgnoresXFF(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")

	// Trusted proxy is "10.0.0.1" but RemoteAddr is "192.168.1.1" → don't trust XFF
	ip := clientIP(req, "10.0.0.1")
	if ip != "192.168.1.1:12345" {
		t.Errorf("clientIP with untrusted peer = %q, want RemoteAddr", ip)
	}
}

// ── RateLimit middleware (global) ────────────────────────────────────────

func TestRateLimit_Middleware_PassesThrough(t *testing.T) {
	handler := RateLimit(100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.1.1.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// ── Recover ──────────────────────────────────────────────────────────────

func TestRecover_CatchesPanic(t *testing.T) {
	handler := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	// This should NOT panic the test
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
	if body := rr.Body.String(); body == "" {
		t.Error("expected error body, got empty")
	}
}

func TestRecover_PassesThroughNormal(t *testing.T) {
	handler := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// ── Logger ───────────────────────────────────────────────────────────────

func TestLogger_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	log := zerolog.New(&buf)

	handler := Logger(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test-path", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	output := buf.String()
	if output == "" {
		t.Fatal("expected log output, got empty")
	}
	for _, want := range []string{"GET", "/test-path", "status", "duration_ms"} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("log output missing %q, got: %s", want, output)
		}
	}
}

func TestLogger_LogsStatus(t *testing.T) {
	var buf bytes.Buffer
	log := zerolog.New(&buf)

	handler := Logger(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest("GET", "/missing", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
	if !bytes.Contains(buf.Bytes(), []byte("404")) {
		t.Errorf("log should contain status 404, got: %s", buf.String())
	}
}

// ── Pre-built rate limiters ─────────────────────────────────────────────

func TestPreBuiltRateLimiters_Exist(t *testing.T) {
	if DeviceRateLimit == nil {
		t.Error("DeviceRateLimit should not be nil")
	}
	if UploadRateLimit == nil {
		t.Error("UploadRateLimit should not be nil")
	}
	if SearchRateLimit == nil {
		t.Error("SearchRateLimit should not be nil")
	}
}
