package auth_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/auth"
	"github.com/rithul/hivemind/registry/api/internal/config"
	"github.com/rithul/hivemind/registry/api/internal/db"
	"github.com/rithul/hivemind/registry/api/internal/testutil"
)

// mockAuthConfig implements auth.AuthConfig for tests.
type mockAuthConfig struct {
	secret string
}

func (m *mockAuthConfig) GetJWTSecret() string { return m.secret }

// okHandler is a simple handler that returns 200.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.GetUserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"user_id": uid.String()})
})

func TestRequireAuth_NoHeader(t *testing.T) {
	cfg := &mockAuthConfig{secret: "test-secret"}
	q := testutil.Queries(t)

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no auth header, got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["error"] != "unauthorized" {
		t.Errorf("error = %q, want %q", resp["error"], "unauthorized")
	}
}

func TestRequireAuth_EmptyBearer(t *testing.T) {
	cfg := &mockAuthConfig{secret: "test-secret"}
	q := testutil.Queries(t)

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer ")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with empty bearer, got %d", rr.Code)
	}
}

func TestRequireAuth_InvalidBearer(t *testing.T) {
	cfg := &mockAuthConfig{secret: "test-secret"}
	q := testutil.Queries(t)

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with invalid bearer, got %d", rr.Code)
	}
}

func TestRequireAuth_LegacyJWT(t *testing.T) {
	secret := "test-jwt-secret-for-auth"
	cfg := &mockAuthConfig{secret: secret}
	q := testutil.Queries(t)

	// Ensure no JWKS verifier is set (legacy path)
	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	// Issue a valid token
	token, err := auth.IssueAccessToken(secret, "00000000-0000-0000-0000-000000000001", "testuser", 1*time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid JWT, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRequireAuth_APIKey_Valid(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	cfg := testutil.TestConfig()

	// Ensure no JWKS verifier is set
	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", seed.APIKey)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid API key, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["user_id"] != seed.UserID.String() {
		t.Errorf("user_id = %q, want %q", resp["user_id"], seed.UserID.String())
	}
}

func TestRequireAuth_APIKey_BasicAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	cfg := testutil.TestConfig()

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	// Twine-style Basic Auth with API key as password
	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("__token__", seed.APIKey)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 with BasicAuth API key, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRequireAuth_APIKey_InvalidHash(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	cfg := testutil.TestConfig()

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "hm_invalid_key_that_does_not_exist_in_db")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with invalid API key, got %d", rr.Code)
	}
}

func TestRequireAuth_APIKey_WrongPrefix(t *testing.T) {
	cfg := &mockAuthConfig{secret: "test-secret"}
	q := testutil.Queries(t)

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	// Key without "hm_" prefix should be rejected early
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "invalid_prefix_key_12345")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for key without hm_ prefix, got %d", rr.Code)
	}
}

// ── Additional RequireAuth branch coverage ──────────────────────────────

func TestRequireAuth_BearerWhitespaceOnly(t *testing.T) {
	// "Bearer    " should be treated as empty token after TrimSpace.
	cfg := &mockAuthConfig{secret: "test-secret"}
	q := testutil.Queries(t)

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer    ")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with whitespace-only bearer, got %d", rr.Code)
	}
}

func TestRequireAuth_LegacyJWT_WrongSecret(t *testing.T) {
	// A valid JWT signed with a different secret should be rejected.
	cfg := &mockAuthConfig{secret: "correct-secret"}
	q := testutil.Queries(t)

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	token, err := auth.IssueAccessToken("wrong-secret", "00000000-0000-0000-0000-000000000001", "testuser", 1*time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with wrong-secret JWT, got %d", rr.Code)
	}
}

func TestRequireAuth_LegacyJWT_ContextValues(t *testing.T) {
	// Verify that the legacy JWT path correctly sets UserID, Username, and Scopes on context.
	secret := "test-jwt-secret-ctx"
	cfg := &mockAuthConfig{secret: secret}
	q := testutil.Queries(t)

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	userID := "00000000-0000-0000-0000-000000000099"
	username := "ctxuser"
	token, err := auth.IssueAccessToken(secret, userID, username, 1*time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}

	// Use a handler that inspects all context values.
	var gotUserID, gotUsername string
	inspectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.GetUserID(r.Context())
		if ok {
			gotUserID = uid.String()
		}
		// Username is stored via ContextKeyUsername; extract it.
		if v := r.Context().Value(auth.ContextKeyUsername); v != nil {
			gotUsername, _ = v.(string)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(inspectHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if gotUserID != userID {
		t.Errorf("context UserID = %q, want %q", gotUserID, userID)
	}
	if gotUsername != username {
		t.Errorf("context Username = %q, want %q", gotUsername, username)
	}
}

func TestRequireAuth_NonBearerAuthHeader(t *testing.T) {
	// An Authorization header that is not "Bearer ..." should fall through
	// to the API key / Basic Auth check, and with no key present → 401.
	cfg := &mockAuthConfig{secret: "test-secret"}
	q := testutil.Queries(t)

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Token some-opaque-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with non-Bearer auth header, got %d", rr.Code)
	}
}

func TestRequireAuth_APIKey_Expired(t *testing.T) {
	// An API key that exists in DB but has an expired expires_at should be rejected.
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	cfg := testutil.TestConfig()
	ctx := context.Background()

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	// Create a user for this test.
	suffix := uuid.New().String()[:8]
	user, err := q.CreateUser(ctx, db.CreateUserParams{
		Email:    fmt.Sprintf("expired-key-%s@example.com", suffix),
		Username: fmt.Sprintf("expireduser-%s", suffix),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Create an API key that expired 1 hour ago.
	rawKey := fmt.Sprintf("hm_expired_%s_%s", suffix, "abcdef0123456789abcdef01234567")
	keyHash := auth.HashKey(rawKey)
	prefix := rawKey[:11] + "..."
	apiKey, err := q.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		UserID:    user.ID,
		OrgID:     pgtype.UUID{},
		Name:      "expired-key",
		KeyHash:   keyHash,
		KeyPrefix: prefix,
		Scopes:    []string{"publish", "read"},
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatalf("create expired api key: %v", err)
	}
	t.Cleanup(func() {
		_ = q.RevokeAPIKey(ctx, apiKey.ID)
	})

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", rawKey)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with expired API key, got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["error"] != "unauthorized" {
		t.Errorf("error = %q, want %q", resp["error"], "unauthorized")
	}
}

func TestRequireAuth_APIKey_Revoked(t *testing.T) {
	// A revoked API key should be rejected (the DB query filters `AND NOT revoked`).
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	cfg := testutil.TestConfig()
	ctx := context.Background()

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	suffix := uuid.New().String()[:8]
	user, err := q.CreateUser(ctx, db.CreateUserParams{
		Email:    fmt.Sprintf("revoked-key-%s@example.com", suffix),
		Username: fmt.Sprintf("revokeduser-%s", suffix),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	rawKey := fmt.Sprintf("hm_revoked_%s_%s", suffix, "abcdef0123456789abcdef0123456")
	keyHash := auth.HashKey(rawKey)
	prefix := rawKey[:11] + "..."
	apiKey, err := q.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		UserID:    user.ID,
		OrgID:     pgtype.UUID{},
		Name:      "revoked-key",
		KeyHash:   keyHash,
		KeyPrefix: prefix,
		Scopes:    []string{"publish", "read"},
		ExpiresAt: pgtype.Timestamptz{},
	})
	if err != nil {
		t.Fatalf("create api key: %v", err)
	}

	// Revoke the key
	if err := q.RevokeAPIKey(ctx, apiKey.ID); err != nil {
		t.Fatalf("revoke api key: %v", err)
	}

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", rawKey)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with revoked API key, got %d", rr.Code)
	}
}

func TestRequireAuth_APIKey_ContextValues(t *testing.T) {
	// Verify that the API key path sets UserID, Scopes, and APIKeyID on context.
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	cfg := testutil.TestConfig()

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	var gotUserID, gotAPIKeyID string
	var gotScopes []string
	var gotUsername string
	inspectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if uid, ok := auth.GetUserID(r.Context()); ok {
			gotUserID = uid.String()
		}
		if v := r.Context().Value(auth.ContextKeyAPIKeyID); v != nil {
			gotAPIKeyID = v.(uuid.UUID).String()
		}
		if v := r.Context().Value(auth.ContextKeyScopes); v != nil {
			gotScopes, _ = v.([]string)
		}
		if v := r.Context().Value(auth.ContextKeyUsername); v != nil {
			gotUsername, _ = v.(string)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(inspectHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", seed.APIKey)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if gotUserID != seed.UserID.String() {
		t.Errorf("context UserID = %q, want %q", gotUserID, seed.UserID.String())
	}
	if gotAPIKeyID != seed.APIKeyID.String() {
		t.Errorf("context APIKeyID = %q, want %q", gotAPIKeyID, seed.APIKeyID.String())
	}
	if gotUsername != "" {
		t.Errorf("context Username = %q, want empty (API key auth sets empty username)", gotUsername)
	}
	if len(gotScopes) != 2 {
		t.Errorf("context Scopes length = %d, want 2", len(gotScopes))
	}
}

func TestRequireAuth_BasicAuth_WrongPrefixPassword(t *testing.T) {
	// Basic Auth with a password that doesn't start with "hm_" → early prefix rejection.
	cfg := &mockAuthConfig{secret: "test-secret"}
	q := testutil.Queries(t)

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("__token__", "not_a_valid_api_key")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with Basic Auth wrong prefix, got %d", rr.Code)
	}
}

func TestRequireAuth_BasicAuth_EmptyPassword(t *testing.T) {
	// Basic Auth with empty password → rawKey is empty → falls through to 401.
	cfg := &mockAuthConfig{secret: "test-secret"}
	q := testutil.Queries(t)

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("__token__", "")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with empty Basic Auth password, got %d", rr.Code)
	}
}

func TestRequireAuth_XAPIKey_EmptyValue(t *testing.T) {
	// X-API-Key header present but empty → falls through to 401.
	cfg := &mockAuthConfig{secret: "test-secret"}
	q := testutil.Queries(t)

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with empty X-API-Key, got %d", rr.Code)
	}
}

func TestRequireAuth_XAPIKey_WhitespaceOnly(t *testing.T) {
	// X-API-Key header with only whitespace → TrimSpace makes it empty → falls through to 401.
	cfg := &mockAuthConfig{secret: "test-secret"}
	q := testutil.Queries(t)

	auth.SetGlobalJWKSVerifier(nil)
	t.Cleanup(func() { auth.SetGlobalJWKSVerifier(nil) })

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "   ")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with whitespace-only X-API-Key, got %d", rr.Code)
	}
}

func TestRequireAuth_ResponseContentType(t *testing.T) {
	// Verify that 401 responses have Content-Type: application/json.
	cfg := &mockAuthConfig{secret: "test-secret"}
	q := testutil.Queries(t)

	middleware := auth.RequireAuth(cfg, q)
	handler := middleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

// ── RequireScope additional branch coverage ─────────────────────────────

func TestRequireScope_APIKey_NilScopes(t *testing.T) {
	// API key context with nil scopes → should be forbidden.
	handler := auth.RequireScope("publish")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), auth.ContextKeyUserID, uuid.New())
	ctx = context.WithValue(ctx, auth.ContextKeyAPIKeyID, uuid.New())
	// No scopes set at all in context
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for API key with nil scopes, got %d", rr.Code)
	}
}

func TestRequireScope_APIKey_EmptyScopes(t *testing.T) {
	// API key context with empty scopes slice → should be forbidden.
	handler := auth.RequireScope("read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), auth.ContextKeyUserID, uuid.New())
	ctx = context.WithValue(ctx, auth.ContextKeyAPIKeyID, uuid.New())
	ctx = context.WithValue(ctx, auth.ContextKeyScopes, []string{}) // empty slice
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for API key with empty scopes, got %d", rr.Code)
	}
}

func TestRequireScope_ResponseBody(t *testing.T) {
	// Verify the 403 response body contains {"error": "insufficient_scope"}.
	handler := auth.RequireScope("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), auth.ContextKeyUserID, uuid.New())
	ctx = context.WithValue(ctx, auth.ContextKeyAPIKeyID, uuid.New())
	ctx = context.WithValue(ctx, auth.ContextKeyScopes, []string{"read"})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["error"] != "insufficient_scope" {
		t.Errorf("error = %q, want %q", resp["error"], "insufficient_scope")
	}
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

// ── RequireAdmin additional branch coverage ─────────────────────────────

func TestRequireAdmin_ResponseBody(t *testing.T) {
	// Verify the 403 response body contains {"error": "insufficient_scope"}.
	handler := auth.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), auth.ContextKeyScopes, []string{"read"})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["error"] != "insufficient_scope" {
		t.Errorf("error = %q, want %q", resp["error"], "insufficient_scope")
	}
}

// ── GetEmail / GetUserIDPgType additional coverage (external package) ───

func TestGetEmail_NonStringValue(t *testing.T) {
	// If someone stores a non-string value for the email key, GetEmail returns "".
	ctx := context.WithValue(context.Background(), auth.ContextKeyEmail, 12345)
	got := auth.GetEmail(ctx)
	if got != "" {
		t.Errorf("GetEmail with non-string value = %q, want empty", got)
	}
}

func TestGetUserID_NonUUIDValue(t *testing.T) {
	// If someone stores a non-UUID value for the user ID key, GetUserID returns false.
	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, "not-a-uuid")
	_, ok := auth.GetUserID(ctx)
	if ok {
		t.Error("GetUserID with non-UUID value should return false")
	}
}

func TestGetUserIDPgType_RoundTrip(t *testing.T) {
	// Verify the pgtype.UUID round-trip from context.
	uid := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, uid)

	pgUUID, ok := auth.GetUserIDPgType(ctx)
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

// Use the real config type to satisfy the interface.
var _ auth.AuthConfig = (*config.Config)(nil)
