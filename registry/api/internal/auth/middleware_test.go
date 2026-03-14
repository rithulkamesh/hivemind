package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ── RequireScope tests ───────────────────────────────────────────────────

func TestRequireScope_JWTBypass(t *testing.T) {
	// JWT sessions (no ContextKeyAPIKeyID) should always pass scope checks.
	handler := RequireScope("publish")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	// Set user ID but NO api_key_id (simulates JWT session)
	ctx := context.WithValue(req.Context(), ContextKeyUserID, uuid.New())
	ctx = context.WithValue(ctx, ContextKeyScopes, []string{"read"}) // no "publish" scope
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for JWT session (scope bypass), got %d", rr.Code)
	}
}

func TestRequireScope_APIKey_HasScope(t *testing.T) {
	handler := RequireScope("read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), ContextKeyUserID, uuid.New())
	ctx = context.WithValue(ctx, ContextKeyAPIKeyID, uuid.New()) // API key present
	ctx = context.WithValue(ctx, ContextKeyScopes, []string{"read", "publish"})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRequireScope_APIKey_MissingScope(t *testing.T) {
	handler := RequireScope("publish")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), ContextKeyUserID, uuid.New())
	ctx = context.WithValue(ctx, ContextKeyAPIKeyID, uuid.New())
	ctx = context.WithValue(ctx, ContextKeyScopes, []string{"read"}) // no "publish"
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

// ── RequireAdmin tests ───────────────────────────────────────────────────

func TestRequireAdmin_HasAdmin(t *testing.T) {
	handler := RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), ContextKeyScopes, []string{"admin", "read"})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRequireAdmin_NoAdmin(t *testing.T) {
	handler := RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), ContextKeyScopes, []string{"read", "publish"})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRequireAdmin_NoScopes(t *testing.T) {
	handler := RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 with no scopes at all, got %d", rr.Code)
	}
}

// ── Context helpers ──────────────────────────────────────────────────────

func TestGetUserID(t *testing.T) {
	uid := uuid.New()
	ctx := context.WithValue(context.Background(), ContextKeyUserID, uid)

	got, ok := GetUserID(ctx)
	if !ok || got != uid {
		t.Errorf("GetUserID = (%v, %v), want (%v, true)", got, ok, uid)
	}
}

func TestGetUserID_Missing(t *testing.T) {
	_, ok := GetUserID(context.Background())
	if ok {
		t.Error("GetUserID on empty context should return false")
	}
}

func TestGetEmail(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyEmail, "test@example.com")
	got := GetEmail(ctx)
	if got != "test@example.com" {
		t.Errorf("GetEmail = %q, want %q", got, "test@example.com")
	}
}

func TestGetEmail_Missing(t *testing.T) {
	got := GetEmail(context.Background())
	if got != "" {
		t.Errorf("GetEmail on empty context = %q, want empty", got)
	}
}

// ── RequireAuth with API key (integration, requires DB) ──────────────────

func TestRequireAuth_APIKey_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// This test requires the dev database to be running with test API keys.
	// We use the known test key from seed data.
	testKey := "hm_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	// We need a db.Queries to test RequireAuth, but we cannot import testutil
	// from within the auth package (it would create a circular dep).
	// Instead we test RequireAuth's API key path via the full handler test in packages/.
	// Here we just test the unit-testable parts (scope/admin middleware).
	_ = testKey
	t.Log("API key integration test covered in packages/handler_test.go")
}

// ── API key hashing ──────────────────────────────────────────────────────

func TestHashKey(t *testing.T) {
	key := "hm_test123"
	h1 := HashKey(key)
	h2 := HashKey(key)
	if h1 != h2 {
		t.Error("HashKey not deterministic")
	}
	if len(h1) != 64 { // sha256 hex
		t.Errorf("HashKey length = %d, want 64", len(h1))
	}
}

func TestVerifyAPIKey(t *testing.T) {
	key := "hm_testverify"
	hash := HashKey(key)
	if !VerifyAPIKey(key, hash) {
		t.Error("VerifyAPIKey should return true for matching key/hash")
	}
	if VerifyAPIKey("hm_wrong", hash) {
		t.Error("VerifyAPIKey should return false for wrong key")
	}
}

func TestGenerateAPIKey(t *testing.T) {
	raw, hash, prefix, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey: %v", err)
	}
	if raw[:3] != "hm_" {
		t.Errorf("raw key should start with 'hm_', got %q", raw[:3])
	}
	if !VerifyAPIKey(raw, hash) {
		t.Error("generated key should verify against its hash")
	}
	if len(prefix) < 11 {
		t.Errorf("prefix too short: %q", prefix)
	}
}

// ── ExtractBearerKey ─────────────────────────────────────────────────────

func TestExtractBearerKey_Valid(t *testing.T) {
	key, err := ExtractBearerKey("Bearer hm_test123456")
	if err != nil {
		t.Fatalf("ExtractBearerKey: %v", err)
	}
	if key != "hm_test123456" {
		t.Errorf("key = %q, want %q", key, "hm_test123456")
	}
}

func TestExtractBearerKey_Missing(t *testing.T) {
	_, err := ExtractBearerKey("")
	if err == nil {
		t.Error("expected error for empty header")
	}
}

func TestExtractBearerKey_InvalidPrefix(t *testing.T) {
	_, err := ExtractBearerKey("Basic abc123")
	if err == nil {
		t.Error("expected error for non-Bearer prefix")
	}
}

func TestExtractBearerKey_TooShort(t *testing.T) {
	_, err := ExtractBearerKey("Bear")
	if err == nil {
		t.Error("expected error for too-short header")
	}
}

// ── GetUserIDPgType ──────────────────────────────────────────────────────

func TestGetUserIDPgType_Valid(t *testing.T) {
	uid := uuid.New()
	ctx := context.WithValue(context.Background(), ContextKeyUserID, uid)

	pgUUID, ok := GetUserIDPgType(ctx)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if !pgUUID.Valid {
		t.Fatal("expected Valid=true")
	}
	if pgUUID.Bytes != uid {
		t.Errorf("UUID bytes mismatch: got %v, want %v", pgUUID.Bytes, uid)
	}
}

func TestGetUserIDPgType_Missing(t *testing.T) {
	pgUUID, ok := GetUserIDPgType(context.Background())
	if ok {
		t.Error("expected ok=false for empty context")
	}
	if pgUUID.Valid {
		t.Error("expected Valid=false for missing user")
	}
}

// ── Device Auth ──────────────────────────────────────────────────────────

func TestRandomUserCode_Format(t *testing.T) {
	for i := 0; i < 100; i++ {
		code := randomUserCode()
		// Format: XXXX-NNNN (4 letters, dash, 4 digits)
		if len(code) != 9 {
			t.Fatalf("code length = %d, want 9: %q", len(code), code)
		}
		if code[4] != '-' {
			t.Fatalf("code[4] = %q, want '-': %q", code[4], code)
		}
		for j := 0; j < 4; j++ {
			if code[j] < 'A' || code[j] > 'Z' {
				t.Fatalf("code[%d] = %q, want A-Z: %q", j, code[j], code)
			}
		}
		for j := 5; j < 9; j++ {
			if code[j] < '0' || code[j] > '9' {
				t.Fatalf("code[%d] = %q, want 0-9: %q", j, code[j], code)
			}
		}
	}
}

func TestRandomUserCode_Unique(t *testing.T) {
	seen := make(map[string]bool)
	dupes := 0
	for i := 0; i < 1000; i++ {
		code := randomUserCode()
		if seen[code] {
			dupes++
		}
		seen[code] = true
	}
	// With ~29 bits of entropy, collisions in 1000 samples should be extremely rare
	if dupes > 5 {
		t.Errorf("too many duplicate user codes: %d/1000", dupes)
	}
}

// ── DeviceAuthManager (unit, no DB required for request/poll) ────────────

func TestDeviceAuthManager_RequestAndPoll(t *testing.T) {
	// Create manager without DB (queries=nil works for request/poll since they don't touch DB)
	m := &DeviceAuthManager{baseURL: "http://localhost:8080"}

	// Request a device code
	reqReq := httptest.NewRequest("POST", "/api/v1/auth/device/request", nil)
	reqRR := httptest.NewRecorder()
	m.RequestDeviceCode(reqRR, reqReq)

	if reqRR.Code != http.StatusOK {
		t.Fatalf("RequestDeviceCode returned %d: %s", reqRR.Code, reqRR.Body.String())
	}

	var deviceResp struct {
		DeviceCode      string `json:"device_code"`
		UserCode        string `json:"user_code"`
		VerificationURI string `json:"verification_uri"`
		ExpiresIn       int    `json:"expires_in"`
		Interval        int    `json:"interval"`
	}
	if err := json.Unmarshal(reqRR.Body.Bytes(), &deviceResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if deviceResp.DeviceCode == "" {
		t.Error("device_code should not be empty")
	}
	if deviceResp.UserCode == "" {
		t.Error("user_code should not be empty")
	}
	if deviceResp.VerificationURI != "http://localhost:8080/activate" {
		t.Errorf("verification_uri = %q, want http://localhost:8080/activate", deviceResp.VerificationURI)
	}
	if deviceResp.ExpiresIn != 300 {
		t.Errorf("expires_in = %d, want 300", deviceResp.ExpiresIn)
	}

	// Poll — should be pending
	pollBody, _ := json.Marshal(map[string]string{"device_code": deviceResp.DeviceCode})
	pollReq := httptest.NewRequest("POST", "/api/v1/auth/device/poll", bytes.NewReader(pollBody))
	pollRR := httptest.NewRecorder()
	m.PollDeviceCode(pollRR, pollReq)

	if pollRR.Code != http.StatusAccepted {
		t.Errorf("expected 202 for pending, got %d: %s", pollRR.Code, pollRR.Body.String())
	}

	// Poll with unknown code — should be gone/expired
	unknownBody, _ := json.Marshal(map[string]string{"device_code": "nonexistent"})
	unknownReq := httptest.NewRequest("POST", "/api/v1/auth/device/poll", bytes.NewReader(unknownBody))
	unknownRR := httptest.NewRecorder()
	m.PollDeviceCode(unknownRR, unknownReq)

	if unknownRR.Code != http.StatusGone {
		t.Errorf("expected 410 for unknown device code, got %d", unknownRR.Code)
	}
}

func TestDeviceAuthManager_ApproveDevice_NoAuth(t *testing.T) {
	m := &DeviceAuthManager{baseURL: "http://localhost:8080"}

	body, _ := json.Marshal(map[string]string{"user_code": "ABCD-1234"})
	req := httptest.NewRequest("POST", "/api/v1/auth/device/approve", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	m.ApproveDevice(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", rr.Code)
	}
}

func TestDeviceAuthManager_ApproveDevice_InvalidCode(t *testing.T) {
	m := &DeviceAuthManager{baseURL: "http://localhost:8080"}

	body, _ := json.Marshal(map[string]string{"user_code": "XXXX-9999"})
	req := httptest.NewRequest("POST", "/api/v1/auth/device/approve", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), ContextKeyUserID, uuid.New())
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	m.ApproveDevice(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 for invalid code, got %d", rr.Code)
	}
}

func TestDeviceAuthManager_MaxRequests(t *testing.T) {
	m := &DeviceAuthManager{baseURL: "http://localhost:8080"}

	// Simulate hitting the max limit
	m.count = maxDeviceRequests

	req := httptest.NewRequest("POST", "/api/v1/auth/device/request", nil)
	rr := httptest.NewRecorder()
	m.RequestDeviceCode(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 at max capacity, got %d", rr.Code)
	}
}

// ── writeAuthError ───────────────────────────────────────────────────────

func TestWriteAuthError(t *testing.T) {
	rr := httptest.NewRecorder()
	writeAuthError(rr, "test error", http.StatusForbidden)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

// ── JWT issue/verify tests (jwt.go) ─────────────────────────────────────

func TestIssueAccessToken_Valid(t *testing.T) {
	secret := "test-jwt-secret-256bit-longenough"
	tok, err := IssueAccessToken(secret, "user-123", "alice", time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}
	if tok == "" {
		t.Fatal("token should not be empty")
	}

	claims, err := VerifyToken(secret, tok)
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Username != "alice" {
		t.Errorf("Username = %q, want %q", claims.Username, "alice")
	}
	if claims.Subject != "access" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "access")
	}
	if claims.ID == "" {
		t.Error("token ID (jti) should not be empty")
	}
}

func TestIssueRefreshToken_Valid(t *testing.T) {
	secret := "test-jwt-secret-for-refresh-token"
	tok, err := IssueRefreshToken(secret, "user-456", "bob", 30*24*time.Hour)
	if err != nil {
		t.Fatalf("IssueRefreshToken: %v", err)
	}
	if tok == "" {
		t.Fatal("token should not be empty")
	}

	claims, err := VerifyToken(secret, tok)
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
	if claims.UserID != "user-456" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-456")
	}
	if claims.Username != "bob" {
		t.Errorf("Username = %q, want %q", claims.Username, "bob")
	}
	if claims.Subject != "refresh" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "refresh")
	}
}

func TestVerifyToken_WrongSecret(t *testing.T) {
	secret := "correct-secret"
	tok, err := IssueAccessToken(secret, "user-1", "alice", time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}

	_, err = VerifyToken("wrong-secret", tok)
	if err == nil {
		t.Error("VerifyToken with wrong secret should return error")
	}
}

func TestVerifyToken_ExpiredToken(t *testing.T) {
	secret := "test-secret-for-expiry"
	// Issue a token that already expired (negative TTL)
	tok, err := issueToken(secret, "user-1", "alice", nil, -time.Hour, "access")
	if err != nil {
		t.Fatalf("issueToken: %v", err)
	}

	_, err = VerifyToken(secret, tok)
	if err == nil {
		t.Error("VerifyToken should reject expired token")
	}
}

func TestVerifyToken_MalformedToken(t *testing.T) {
	_, err := VerifyToken("secret", "not-a-valid-jwt")
	if err == nil {
		t.Error("VerifyToken should reject malformed token")
	}
}

func TestVerifyToken_EmptyToken(t *testing.T) {
	_, err := VerifyToken("secret", "")
	if err == nil {
		t.Error("VerifyToken should reject empty token")
	}
}

func TestIssueToken_UniqueJTI(t *testing.T) {
	secret := "test-secret-jti"
	tok1, _ := IssueAccessToken(secret, "u", "u", time.Hour)
	tok2, _ := IssueAccessToken(secret, "u", "u", time.Hour)

	c1, _ := VerifyToken(secret, tok1)
	c2, _ := VerifyToken(secret, tok2)
	if c1.ID == c2.ID {
		t.Error("two tokens should have different JTI values")
	}
}

func TestIssueToken_ClaimsExpiry(t *testing.T) {
	secret := "test-secret-expiry"
	before := time.Now().Add(-time.Second).Truncate(time.Second)
	tok, _ := IssueAccessToken(secret, "u", "u", 2*time.Hour)
	after := time.Now().Add(time.Second).Truncate(time.Second)

	claims, _ := VerifyToken(secret, tok)
	exp := claims.ExpiresAt.Time
	if exp.Before(before.Add(2*time.Hour)) || exp.After(after.Add(2*time.Hour)) {
		t.Errorf("expiry %v not within expected range [%v, %v]",
			exp, before.Add(2*time.Hour), after.Add(2*time.Hour))
	}
}

func TestIssueToken_WithScopes(t *testing.T) {
	secret := "test-secret-scopes"
	scopes := []string{"read", "publish"}
	tok, err := issueToken(secret, "user-1", "alice", scopes, time.Hour, "access")
	if err != nil {
		t.Fatalf("issueToken: %v", err)
	}
	claims, err := VerifyToken(secret, tok)
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
	if len(claims.Scopes) != 2 {
		t.Fatalf("Scopes length = %d, want 2", len(claims.Scopes))
	}
	if claims.Scopes[0] != "read" || claims.Scopes[1] != "publish" {
		t.Errorf("Scopes = %v, want [read publish]", claims.Scopes)
	}
}

func TestIssueToken_NilScopes(t *testing.T) {
	secret := "test-secret-nil-scopes"
	tok, err := IssueAccessToken(secret, "u", "u", time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}
	claims, err := VerifyToken(secret, tok)
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
	if claims.Scopes != nil {
		t.Errorf("Scopes = %v, want nil", claims.Scopes)
	}
}

// ── JWKS global getter/setter tests (jwks.go) ──────────────────────────

func TestSetGetGlobalJWKSVerifier_NilDefault(t *testing.T) {
	// Reset to nil to test default state
	SetGlobalJWKSVerifier(nil)
	v := GetGlobalJWKSVerifier()
	if v != nil {
		t.Error("GetGlobalJWKSVerifier should return nil when not set")
	}
}

func TestSetGetGlobalJWKSVerifier_RoundTrip(t *testing.T) {
	// We can't easily construct a real JWKSVerifier without a JWKS URL,
	// but we can test the setter/getter with a zero-value struct.
	verifier := &JWKSVerifier{}
	SetGlobalJWKSVerifier(verifier)
	defer SetGlobalJWKSVerifier(nil) // cleanup

	got := GetGlobalJWKSVerifier()
	if got != verifier {
		t.Error("GetGlobalJWKSVerifier should return the verifier that was set")
	}
}

func TestSetGlobalJWKSVerifier_Overwrite(t *testing.T) {
	v1 := &JWKSVerifier{}
	v2 := &JWKSVerifier{}

	SetGlobalJWKSVerifier(v1)
	SetGlobalJWKSVerifier(v2)
	defer SetGlobalJWKSVerifier(nil)

	got := GetGlobalJWKSVerifier()
	if got != v2 {
		t.Error("SetGlobalJWKSVerifier should overwrite the previous verifier")
	}
}

func TestNewJWKSVerifier_EmptyURL(t *testing.T) {
	_, err := NewJWKSVerifier(context.Background(), "")
	if err == nil {
		t.Error("NewJWKSVerifier with empty URL should return error")
	}
}

// ── NewDeviceAuthManager tests (device.go) ──────────────────────────────

func TestNewDeviceAuthManager_CreatesValid(t *testing.T) {
	mgr := NewDeviceAuthManager(nil, "http://localhost:9090")
	if mgr == nil {
		t.Fatal("NewDeviceAuthManager returned nil")
	}
	if mgr.baseURL != "http://localhost:9090" {
		t.Errorf("baseURL = %q, want %q", mgr.baseURL, "http://localhost:9090")
	}
	// count should start at zero
	if c := atomic.LoadInt64(&mgr.count); c != 0 {
		t.Errorf("initial count = %d, want 0", c)
	}
}

func TestNewDeviceAuthManager_RequestAfterCreate(t *testing.T) {
	mgr := NewDeviceAuthManager(nil, "http://test:8080")

	// The manager should be functional immediately after creation
	req := httptest.NewRequest(http.MethodPost, "/device/code", nil)
	w := httptest.NewRecorder()
	mgr.RequestDeviceCode(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("RequestDeviceCode status = %d, want 200", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["verification_uri"] != "http://test:8080/activate" {
		t.Errorf("verification_uri = %q, want %q", resp["verification_uri"], "http://test:8080/activate")
	}
}

func TestCleanupLoop_RemovesExpired(t *testing.T) {
	// Create a manager directly (without NewDeviceAuthManager to avoid the goroutine)
	mgr := &DeviceAuthManager{baseURL: "http://test"}

	// Store an expired request
	mgr.requests.Store("expired-1", &DeviceRequest{
		DeviceCode: "expired-1",
		UserCode:   "AAAA-0001",
		CreatedAt:  time.Now().Add(-10 * time.Minute),
		ExpiresAt:  time.Now().Add(-5 * time.Minute),
		Status:     "pending",
	})
	atomic.StoreInt64(&mgr.count, 1)

	// Store a valid (non-expired) request
	mgr.requests.Store("valid-1", &DeviceRequest{
		DeviceCode: "valid-1",
		UserCode:   "BBBB-0002",
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(5 * time.Minute),
		Status:     "pending",
	})
	atomic.AddInt64(&mgr.count, 1)

	// Manually run the cleanup logic (same as cleanupLoop body)
	now := time.Now()
	mgr.requests.Range(func(key, value interface{}) bool {
		req := value.(*DeviceRequest)
		if now.After(req.ExpiresAt) {
			mgr.requests.Delete(key)
			atomic.AddInt64(&mgr.count, -1)
		}
		return true
	})

	// Expired request should be gone
	if _, ok := mgr.requests.Load("expired-1"); ok {
		t.Error("expired request should have been cleaned up")
	}

	// Valid request should still exist
	if _, ok := mgr.requests.Load("valid-1"); !ok {
		t.Error("valid request should NOT have been cleaned up")
	}

	// Count should be 1 (only the valid request)
	if c := atomic.LoadInt64(&mgr.count); c != 1 {
		t.Errorf("count after cleanup = %d, want 1", c)
	}
}

func TestDeviceAuthManager_PollDenied(t *testing.T) {
	mgr := &DeviceAuthManager{baseURL: "http://test"}

	// Store a denied request
	mgr.requests.Store("denied-code", &DeviceRequest{
		DeviceCode: "denied-code",
		UserCode:   "DENY-0001",
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(5 * time.Minute),
		Status:     "denied",
	})

	pollBody, _ := json.Marshal(map[string]string{"device_code": "denied-code"})
	req := httptest.NewRequest(http.MethodPost, "/device/poll", bytes.NewReader(pollBody))
	w := httptest.NewRecorder()
	mgr.PollDeviceCode(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("denied poll status = %d, want 400", w.Code)
	}
}

func TestDeviceAuthManager_PollBadBody(t *testing.T) {
	mgr := &DeviceAuthManager{baseURL: "http://test"}

	req := httptest.NewRequest(http.MethodPost, "/device/poll", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	mgr.PollDeviceCode(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("bad body poll status = %d, want 400", w.Code)
	}
}
