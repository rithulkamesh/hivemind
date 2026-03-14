package users_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/rithul/hivemind/registry/api/internal/auth"
	"github.com/rithul/hivemind/registry/api/internal/testutil"
	"github.com/rithul/hivemind/registry/api/internal/users"
)

// addAuthContext injects auth context values matching what RequireAuth sets.
func addAuthContext(r *http.Request, userID uuid.UUID, scopes []string) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, auth.ContextKeyUsername, "testuser")
	ctx = context.WithValue(ctx, auth.ContextKeyScopes, scopes)
	return r.WithContext(ctx)
}

// ── GetMe ─────────────────────────────────────────────────────────────────

func TestGetMe(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := users.NewHandler(q)

	req := httptest.NewRequest("GET", "/api/v1/me", nil)
	req = addAuthContext(req, seed.UserID, []string{"read"})
	rr := httptest.NewRecorder()
	h.GetMe(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GetMe returned %d: %s", rr.Code, rr.Body.String())
	}
	var me struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &me); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if me.ID != seed.UserID.String() {
		t.Errorf("id = %q, want %q", me.ID, seed.UserID.String())
	}
	if me.Email != seed.UserEmail {
		t.Errorf("email = %q, want %q", me.Email, seed.UserEmail)
	}
}

func TestGetMe_NoAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := users.NewHandler(q)

	req := httptest.NewRequest("GET", "/api/v1/me", nil)
	rr := httptest.NewRecorder()
	h.GetMe(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// ── ListAPIKeys ───────────────────────────────────────────────────────────

func TestListAPIKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := users.NewHandler(q)

	req := httptest.NewRequest("GET", "/api/v1/me/api-keys", nil)
	req = addAuthContext(req, seed.UserID, []string{"read"})
	rr := httptest.NewRecorder()
	h.ListAPIKeys(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListAPIKeys returned %d: %s", rr.Code, rr.Body.String())
	}
	var keys []json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &keys); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Should find at least our test key
	if len(keys) < 1 {
		t.Error("expected at least 1 API key")
	}
}

func TestListAPIKeys_EmptyForNewUser(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := users.NewHandler(q)

	// Random user with no keys
	req := httptest.NewRequest("GET", "/api/v1/me/api-keys", nil)
	req = addAuthContext(req, uuid.New(), []string{"read"})
	rr := httptest.NewRecorder()
	h.ListAPIKeys(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListAPIKeys returned %d: %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if body != "[]\n" {
		t.Errorf("expected empty array, got %q", body)
	}
}

// ── CreateAPIKey ──────────────────────────────────────────────────────────

func TestCreateAPIKey(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := users.NewHandler(q)

	body, _ := json.Marshal(map[string]interface{}{
		"name":   "ci-key",
		"scopes": []string{"read"},
	})
	req := httptest.NewRequest("POST", "/api/v1/me/api-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, seed.UserID, []string{"read"})
	rr := httptest.NewRecorder()
	h.CreateAPIKey(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("CreateAPIKey returned %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Key    string `json:"key"`
		Prefix string `json:"prefix"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Key == "" || resp.Key[:3] != "hm_" {
		t.Errorf("key should start with hm_, got %q", resp.Key)
	}
}

// ── RevokeAPIKey ──────────────────────────────────────────────────────────

func TestRevokeAPIKey(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := users.NewHandler(q)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", seed.APIKeyID.String())

	req := httptest.NewRequest("DELETE", "/api/v1/me/api-keys/"+seed.APIKeyID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = addAuthContext(req, seed.UserID, []string{"read"})
	rr := httptest.NewRecorder()
	h.RevokeAPIKey(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("RevokeAPIKey returned %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRevokeAPIKey_InvalidID(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := users.NewHandler(q)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "not-a-uuid")

	req := httptest.NewRequest("DELETE", "/api/v1/me/api-keys/not-a-uuid", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = addAuthContext(req, seed.UserID, []string{"read"})
	rr := httptest.NewRecorder()
	h.RevokeAPIKey(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// ── Admin endpoints ───────────────────────────────────────────────────────

func TestAdminListUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	testutil.SetupTestData(t, q)
	h := users.NewHandler(q)

	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	rr := httptest.NewRecorder()
	h.AdminListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("AdminListUsers returned %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAdminBanUser(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := users.NewHandler(q)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", seed.UserID.String())

	req := httptest.NewRequest("POST", "/api/v1/admin/users/"+seed.UserID.String()+"/ban", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()
	h.AdminBanUser(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("AdminBanUser returned %d: %s", rr.Code, rr.Body.String())
	}
}

// ── UpdateMe ──────────────────────────────────────────────────────────────

func TestUpdateMe_LegacyUser(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := users.NewHandler(q)

	body, _ := json.Marshal(map[string]string{
		"username": fmt.Sprintf("updated-%s", uuid.New().String()[:8]),
	})
	req := httptest.NewRequest("PUT", "/api/v1/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, seed.UserID, []string{"read"})
	rr := httptest.NewRecorder()
	h.UpdateMe(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("UpdateMe returned %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdateMe_NoAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := users.NewHandler(q)

	body, _ := json.Marshal(map[string]string{"username": "x"})
	req := httptest.NewRequest("PUT", "/api/v1/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateMe(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestUpdateMe_BadJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := users.NewHandler(q)

	req := httptest.NewRequest("PUT", "/api/v1/me", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, seed.UserID, []string{"read"})
	rr := httptest.NewRecorder()
	h.UpdateMe(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// ── GetMe with Better Auth profile ──────────────────────────────────────

func TestGetMe_CreatesProfileFromEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := users.NewHandler(q)

	// Use a fresh UUID that has no profile or user record,
	// but set email context (simulating Better Auth JWT).
	suffix := uuid.New().String()[:8]
	freshUID := uuid.New()
	email := fmt.Sprintf("newuser-%s@example.com", suffix)
	req := httptest.NewRequest("GET", "/api/v1/me", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, freshUID)
	ctx = context.WithValue(ctx, auth.ContextKeyUsername, fmt.Sprintf("newuser-%s", suffix))
	ctx = context.WithValue(ctx, auth.ContextKeyScopes, []string{"read"})
	ctx = context.WithValue(ctx, auth.ContextKeyEmail, email)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.GetMe(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GetMe returned %d: %s", rr.Code, rr.Body.String())
	}
	var me struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &me); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if me.Email != email {
		t.Errorf("email = %q, want %q", me.Email, email)
	}
}

// ── UpdateMe with Better Auth profile ───────────────────────────────────

func TestUpdateMe_BetterAuthProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := users.NewHandler(q)

	suffix := uuid.New().String()[:8]

	// First create a profile via GetMe with email
	freshUID := uuid.New()
	getReq := httptest.NewRequest("GET", "/api/v1/me", nil)
	ctx := getReq.Context()
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, freshUID)
	ctx = context.WithValue(ctx, auth.ContextKeyUsername, fmt.Sprintf("baprofile-%s", suffix))
	ctx = context.WithValue(ctx, auth.ContextKeyScopes, []string{"read"})
	ctx = context.WithValue(ctx, auth.ContextKeyEmail, fmt.Sprintf("baprofile-%s@example.com", suffix))
	getReq = getReq.WithContext(ctx)
	getRR := httptest.NewRecorder()
	h.GetMe(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("setup GetMe: %d: %s", getRR.Code, getRR.Body.String())
	}

	// Now update via the profile path
	body, _ := json.Marshal(map[string]string{
		"username": fmt.Sprintf("updated-ba-%s", suffix),
		"bio":      "test bio",
		"website":  "https://example.com",
	})
	updateReq := httptest.NewRequest("PUT", "/api/v1/me", bytes.NewReader(body))
	updateReq.Header.Set("Content-Type", "application/json")
	uctx := updateReq.Context()
	uctx = context.WithValue(uctx, auth.ContextKeyUserID, freshUID)
	uctx = context.WithValue(uctx, auth.ContextKeyUsername, fmt.Sprintf("baprofile-%s", suffix))
	uctx = context.WithValue(uctx, auth.ContextKeyScopes, []string{"read"})
	updateReq = updateReq.WithContext(uctx)
	updateRR := httptest.NewRecorder()
	h.UpdateMe(updateRR, updateReq)

	if updateRR.Code != http.StatusOK {
		t.Fatalf("UpdateMe profile returned %d: %s", updateRR.Code, updateRR.Body.String())
	}
}

// ── 2FA stubs ────────────────────────────────────────────────────────────

func TestSetup2FA_NotImplemented(t *testing.T) {
	q := testutil.Queries(t)
	h := users.NewHandler(q)
	req := httptest.NewRequest("POST", "/api/v1/me/2fa/setup", nil)
	rr := httptest.NewRecorder()
	h.Setup2FA(rr, req)
	if rr.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

func TestVerify2FA_NotImplemented(t *testing.T) {
	q := testutil.Queries(t)
	h := users.NewHandler(q)
	req := httptest.NewRequest("POST", "/api/v1/me/2fa/verify", nil)
	rr := httptest.NewRecorder()
	h.Verify2FA(rr, req)
	if rr.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

// ── AdminBanUser edge cases ─────────────────────────────────────────────

func TestAdminBanUser_InvalidID(t *testing.T) {
	q := testutil.Queries(t)
	h := users.NewHandler(q)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "not-a-uuid")

	req := httptest.NewRequest("POST", "/api/v1/admin/users/not-a-uuid/ban", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()
	h.AdminBanUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// ── CreateAPIKey edge cases ─────────────────────────────────────────────

func TestCreateAPIKey_NoAuth(t *testing.T) {
	q := testutil.Queries(t)
	h := users.NewHandler(q)

	body, _ := json.Marshal(map[string]interface{}{"name": "key", "scopes": []string{"read"}})
	req := httptest.NewRequest("POST", "/api/v1/me/api-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateAPIKey(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestCreateAPIKey_BadJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := users.NewHandler(q)

	req := httptest.NewRequest("POST", "/api/v1/me/api-keys", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, seed.UserID, []string{"read"})
	rr := httptest.NewRecorder()
	h.CreateAPIKey(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// ── ListAPIKeys_NoAuth ──────────────────────────────────────────────────

func TestListAPIKeys_NoAuth(t *testing.T) {
	q := testutil.Queries(t)
	h := users.NewHandler(q)

	req := httptest.NewRequest("GET", "/api/v1/me/api-keys", nil)
	rr := httptest.NewRecorder()
	h.ListAPIKeys(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// ── RevokeAPIKey_NoAuth ─────────────────────────────────────────────────

func TestRevokeAPIKey_NoAuth(t *testing.T) {
	q := testutil.Queries(t)
	h := users.NewHandler(q)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", uuid.New().String())
	req := httptest.NewRequest("DELETE", "/api/v1/me/api-keys/x", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()
	h.RevokeAPIKey(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
