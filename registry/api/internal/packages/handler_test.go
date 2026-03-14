package packages_test

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/auth"
	"github.com/rithul/hivemind/registry/api/internal/db"
	"github.com/rithul/hivemind/registry/api/internal/packages"
	"github.com/rithul/hivemind/registry/api/internal/testutil"
)

// makeTestWheel creates a minimal valid .whl (zip) archive for testing.
func makeTestWheel(t *testing.T, name, version string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Wheel must contain a METADATA file in the .dist-info directory
	distInfo := fmt.Sprintf("%s-%s.dist-info/METADATA",
		strings.ReplaceAll(name, "-", "_"), version)

	mw, err := zw.Create(distInfo)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(mw, "Metadata-Version: 2.1\nName: %s\nVersion: %s\n", name, version)

	// Also add WHEEL file (pip expects it)
	ww, err := zw.Create(fmt.Sprintf("%s-%s.dist-info/WHEEL",
		strings.ReplaceAll(name, "-", "_"), version))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprint(ww, "Wheel-Version: 1.0\nGenerator: test\nRoot-Is-Purelib: true\nTag: py3-none-any\n")

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// withChiParam creates an http.Request with chi URL params set.
func withChiParam(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func withChiParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// addAuthContext injects auth context values matching what RequireAuth sets.
func addAuthContext(r *http.Request, userID uuid.UUID, scopes []string) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, auth.ContextKeyUsername, "testuser")
	ctx = context.WithValue(ctx, auth.ContextKeyScopes, scopes)
	return r.WithContext(ctx)
}

func newHandler(t *testing.T) (*packages.Handler, *db.Queries, *testutil.MockStorage) {
	t.Helper()
	q := testutil.Queries(t)
	store := testutil.NewMockStorage()
	cfg := testutil.TestConfig()
	return packages.NewHandler(q, store, nil, nil, cfg), q, store
}

// ── ListPackages ──────────────────────────────────────────────────────────

func TestListPackages(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)
	_ = seed

	req := httptest.NewRequest("GET", "/api/v1/packages", nil)
	rr := httptest.NewRecorder()
	h.ListPackages(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListPackages returned %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Packages []json.RawMessage `json:"packages"`
		Page     int               `json:"page"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Page != 1 {
		t.Errorf("page = %d, want 1", resp.Page)
	}
	// Should return at least the seeded package
	if len(resp.Packages) < 1 {
		t.Error("expected at least 1 package")
	}
}

func TestListPackages_ReturnsEmptyArray(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, _, _ := newHandler(t)

	// Query with a namespace that surely doesn't exist
	req := httptest.NewRequest("GET", "/api/v1/packages?namespace=nonexistent-ns-xyz", nil)
	rr := httptest.NewRecorder()
	h.ListPackages(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListPackages returned %d", rr.Code)
	}
	body := rr.Body.String()
	// Must be [] not null
	if strings.Contains(body, `"packages":null`) {
		t.Error("packages should be [] not null when empty")
	}
}

// ── GetPackage ────────────────────────────────────────────────────────────

func TestGetPackage(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	req := httptest.NewRequest("GET", "/api/v1/packages/"+seed.PackageName, nil)
	req = withChiParam(req, "name", seed.PackageName)
	rr := httptest.NewRecorder()
	h.GetPackage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GetPackage returned %d: %s", rr.Code, rr.Body.String())
	}
	var pkg db.Package
	if err := json.Unmarshal(rr.Body.Bytes(), &pkg); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if pkg.Name != seed.PackageName {
		t.Errorf("name = %q, want %q", pkg.Name, seed.PackageName)
	}
}

func TestGetPackage_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, _, _ := newHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/packages/nonexistent-pkg-xyz", nil)
	req = withChiParam(req, "name", "nonexistent-pkg-xyz")
	rr := httptest.NewRecorder()
	h.GetPackage(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── ListVersions ──────────────────────────────────────────────────────────

func TestListVersions(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	req := httptest.NewRequest("GET", "/api/v1/packages/"+seed.PackageName+"/versions", nil)
	req = withChiParam(req, "name", seed.PackageName)
	rr := httptest.NewRecorder()
	h.ListVersions(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListVersions returned %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Versions []json.RawMessage `json:"versions"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Versions) != 1 {
		t.Errorf("expected 1 version, got %d", len(resp.Versions))
	}
}

func TestListVersions_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, _, _ := newHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/packages/nonexistent-pkg-xyz/versions", nil)
	req = withChiParam(req, "name", "nonexistent-pkg-xyz")
	rr := httptest.NewRecorder()
	h.ListVersions(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── GetVersion ────────────────────────────────────────────────────────────

func TestGetVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	req := httptest.NewRequest("GET", "/api/v1/packages/"+seed.PackageName+"/1.0.0", nil)
	req = withChiParams(req, map[string]string{
		"name":    seed.PackageName,
		"version": "1.0.0",
	})
	rr := httptest.NewRecorder()
	h.GetVersion(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GetVersion returned %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGetVersion_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	req := httptest.NewRequest("GET", "/", nil)
	req = withChiParams(req, map[string]string{
		"name":    seed.PackageName,
		"version": "99.99.99",
	})
	rr := httptest.NewRecorder()
	h.GetVersion(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── CreatePackage ─────────────────────────────────────────────────────────

func TestCreatePackage(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	pkgName := fmt.Sprintf("new-pkg-%s", uuid.New().String()[:8])
	body, _ := json.Marshal(map[string]string{
		"Name":        pkgName,
		"Description": "A brand new package",
	})

	req := httptest.NewRequest("POST", "/api/v1/packages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, seed.UserID, []string{"publish", "read"})
	rr := httptest.NewRecorder()
	h.CreatePackage(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("CreatePackage returned %d: %s", rr.Code, rr.Body.String())
	}

	// Cleanup
	t.Cleanup(func() {
		pkg, err := q.GetPackageByNamespaceName(context.Background(), db.GetPackageByNamespaceNameParams{
			Namespace: pgtype.Text{},
			Name:      pkgName,
		})
		if err == nil {
			_ = q.DeletePackage(context.Background(), pkg.ID)
		}
	})
}

func TestCreatePackage_DuplicateName(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	body, _ := json.Marshal(map[string]string{
		"Name":        seed.PackageName,
		"Description": "duplicate",
	})

	req := httptest.NewRequest("POST", "/api/v1/packages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, seed.UserID, []string{"publish"})
	rr := httptest.NewRecorder()
	h.CreatePackage(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreatePackage_NoAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, _, _ := newHandler(t)

	body, _ := json.Marshal(map[string]string{"Name": "no-auth-pkg"})
	req := httptest.NewRequest("POST", "/api/v1/packages", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.CreatePackage(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth context, got %d", rr.Code)
	}
}

// ── Upload ────────────────────────────────────────────────────────────────

func TestUpload(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, store := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	// Build multipart form with a valid .whl (zip archive)
	version := fmt.Sprintf("2.0.%s", uuid.New().String()[:4])
	wheelBytes := makeTestWheel(t, seed.PackageName, version)
	wheelHash := fmt.Sprintf("%x", sha256.Sum256(wheelBytes))

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	w.WriteField("name", seed.PackageName)
	w.WriteField("version", version)
	w.WriteField("sha256_digest", wheelHash)
	fw, _ := w.CreateFormFile("file", seed.PackageName+"-"+version+"-py3-none-any.whl")
	fw.Write(wheelBytes)
	w.Close()

	req := httptest.NewRequest("POST", "/api/v1/packages/"+seed.PackageName+"/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req = withChiParam(req, "name", seed.PackageName)
	req = addAuthContext(req, seed.UserID, []string{"publish"})

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Upload returned %d: %s", rr.Code, rr.Body.String())
	}

	// Verify file was stored in mock storage
	if len(store.Files) == 0 {
		t.Error("expected file to be stored in mock storage")
	}

	// Cleanup
	t.Cleanup(func() {
		ctx := context.Background()
		pv, err := q.GetPackageVersion(ctx, db.GetPackageVersionParams{
			Namespace: pgtype.Text{},
			Name:      seed.PackageName,
			Version:   version,
		})
		if err == nil {
			_ = q.DeletePackageVersion(ctx, pv.ID)
		}
	})
}

func TestUpload_NoStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	cfg := testutil.TestConfig()
	// Handler with nil storage
	h := packages.NewHandler(q, nil, nil, nil, cfg)
	seed := testutil.SetupTestData(t, q)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	w.WriteField("name", seed.PackageName)
	w.WriteField("version", "3.0.0")
	fw, _ := w.CreateFormFile("file", "test.whl")
	fw.Write([]byte("data"))
	w.Close()

	req := httptest.NewRequest("POST", "/api/v1/packages/"+seed.PackageName+"/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req = withChiParam(req, "name", seed.PackageName)
	req = addAuthContext(req, seed.UserID, []string{"publish"})
	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

// ── Yank ──────────────────────────────────────────────────────────────────

func TestYank(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	body, _ := json.Marshal(map[string]string{"Reason": "testing yank"})
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req = withChiParams(req, map[string]string{
		"name":    seed.PackageName,
		"version": "1.0.0",
	})
	req = addAuthContext(req, seed.UserID, []string{"publish"})
	rr := httptest.NewRecorder()
	h.Yank(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Yank returned %d: %s", rr.Code, rr.Body.String())
	}

	// Verify it's yanked
	pv, err := q.GetPackageVersion(context.Background(), db.GetPackageVersionParams{
		Namespace: pgtype.Text{},
		Name:      seed.PackageName,
		Version:   "1.0.0",
	})
	if err != nil {
		t.Fatalf("get version after yank: %v", err)
	}
	if !pv.Yanked.Bool {
		t.Error("version should be yanked")
	}
}

func TestYank_NotOwner(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	testutil.SetupTestData(t, q)

	// Use a random user (not the owner)
	otherUser := uuid.New()
	body, _ := json.Marshal(map[string]string{"Reason": "unauthorized"})
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	seed := testutil.SetupTestData(t, q) // create another fresh one for this test
	req = withChiParams(req, map[string]string{
		"name":    seed.PackageName,
		"version": "1.0.0",
	})
	req = addAuthContext(req, otherUser, []string{"publish"})
	rr := httptest.NewRecorder()
	h.Yank(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-owner, got %d", rr.Code)
	}
}

// ── Stats ─────────────────────────────────────────────────────────────────

func TestStats(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, _, _ := newHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/stats", nil)
	rr := httptest.NewRecorder()
	h.Stats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Stats returned %d: %s", rr.Code, rr.Body.String())
	}
}

// ── RequireAuth + handler integration ─────────────────────────────────────

func TestRequireAuth_APIKey_ListPackages(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	cfg := testutil.TestConfig()
	store := testutil.NewMockStorage()
	h := packages.NewHandler(q, store, nil, nil, cfg)

	// Build a chi router with RequireAuth middleware
	r := chi.NewRouter()
	r.Use(auth.RequireAuth(cfg, q))
	r.Get("/api/v1/packages", h.ListPackages)

	req := httptest.NewRequest("GET", "/api/v1/packages", nil)
	req.Header.Set("X-API-Key", seed.APIKey)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRequireAuth_InvalidAPIKey(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	cfg := testutil.TestConfig()
	store := testutil.NewMockStorage()
	h := packages.NewHandler(q, store, nil, nil, cfg)

	r := chi.NewRouter()
	r.Use(auth.RequireAuth(cfg, q))
	r.Get("/api/v1/packages", h.ListPackages)

	req := httptest.NewRequest("GET", "/api/v1/packages", nil)
	req.Header.Set("X-API-Key", "hm_invalid_key_that_does_not_exist")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRequireAuth_NoCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	cfg := testutil.TestConfig()

	r := chi.NewRouter()
	r.Use(auth.RequireAuth(cfg, q))
	r.Get("/api/v1/me", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/v1/me", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// ── BasicAuth (Twine compat) ──────────────────────────────────────────────

func TestRequireAuth_BasicAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	cfg := testutil.TestConfig()
	h := packages.NewHandler(q, nil, nil, nil, cfg)

	r := chi.NewRouter()
	r.Use(auth.RequireAuth(cfg, q))
	r.Get("/test", h.ListPackages)

	req := httptest.NewRequest("GET", "/test", nil)
	req.SetBasicAuth("__token__", seed.APIKey) // Twine-style
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("BasicAuth should work, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── Additional Upload & CreatePackage edge-case tests ─────────────────────

func TestUpload_TarGz(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, store := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	version := fmt.Sprintf("3.0.%s", uuid.New().String()[:4])
	tarContent := []byte("fake tar.gz content - not a real archive but extension is enough")
	tarHash := fmt.Sprintf("%x", sha256.Sum256(tarContent))

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	w.WriteField("name", seed.PackageName)
	w.WriteField("version", version)
	w.WriteField("sha256_digest", tarHash)
	fw, _ := w.CreateFormFile("file", "pkgname-1.0.0.tar.gz")
	fw.Write(tarContent)
	w.Close()

	req := httptest.NewRequest("POST", "/api/v1/packages/"+seed.PackageName+"/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req = withChiParam(req, "name", seed.PackageName)
	req = addAuthContext(req, seed.UserID, []string{"publish"})

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Upload .tar.gz returned %d: %s", rr.Code, rr.Body.String())
	}
	if len(store.Files) == 0 {
		t.Error("expected file to be stored in mock storage")
	}

	t.Cleanup(func() {
		ctx := context.Background()
		pv, err := q.GetPackageVersion(ctx, db.GetPackageVersionParams{
			Namespace: pgtype.Text{},
			Name:      seed.PackageName,
			Version:   version,
		})
		if err == nil {
			_ = q.DeletePackageVersion(ctx, pv.ID)
		}
	})
}

func TestUpload_ExeRejected(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	version := fmt.Sprintf("1.0.%s", uuid.New().String()[:4])

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	w.WriteField("name", seed.PackageName)
	w.WriteField("version", version)
	fw, _ := w.CreateFormFile("file", "malware.exe")
	fw.Write([]byte("evil bytes"))
	w.Close()

	req := httptest.NewRequest("POST", "/api/v1/packages/"+seed.PackageName+"/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req = withChiParam(req, "name", seed.PackageName)
	req = addAuthContext(req, seed.UserID, []string{"publish"})

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for .exe, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "invalid file extension") {
		t.Errorf("expected 'invalid file extension' in body, got: %s", rr.Body.String())
	}
}

func TestUpload_EmptyFile(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	w.WriteField("name", seed.PackageName)
	w.WriteField("version", "1.0.0")
	// No "file" or "content" field
	w.Close()

	req := httptest.NewRequest("POST", "/api/v1/packages/"+seed.PackageName+"/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req = withChiParam(req, "name", seed.PackageName)
	req = addAuthContext(req, seed.UserID, []string{"publish"})

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing file, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpload_InvalidVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	w.WriteField("name", seed.PackageName)
	w.WriteField("version", "../../../etc/passwd")
	fw, _ := w.CreateFormFile("file", "test-1.0.0.whl")
	fw.Write(makeTestWheel(t, seed.PackageName, "1.0.0"))
	w.Close()

	req := httptest.NewRequest("POST", "/api/v1/packages/"+seed.PackageName+"/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req = withChiParam(req, "name", seed.PackageName)
	req = addAuthContext(req, seed.UserID, []string{"publish"})

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid version, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreatePackage_InvalidName(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	body, _ := json.Marshal(map[string]string{
		"Name":        "' OR 1=1--",
		"Description": "sql injection attempt",
	})

	req := httptest.NewRequest("POST", "/api/v1/packages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, seed.UserID, []string{"publish"})
	rr := httptest.NewRecorder()
	h.CreatePackage(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid name, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "invalid name") {
		t.Errorf("expected 'invalid name' in body, got: %s", rr.Body.String())
	}
}

func TestCreatePackage_TooLongName(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	longName := strings.Repeat("a", 200)
	body, _ := json.Marshal(map[string]string{
		"Name":        longName,
		"Description": "too long name",
	})

	req := httptest.NewRequest("POST", "/api/v1/packages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, seed.UserID, []string{"publish"})
	rr := httptest.NewRecorder()
	h.CreatePackage(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for too-long name, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpload_PathTraversalFilename(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	version := fmt.Sprintf("1.0.%s", uuid.New().String()[:4])

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	w.WriteField("name", seed.PackageName)
	w.WriteField("version", version)
	// Go's multipart reader sanitizes directory traversals via filepath.Base,
	// so use a filename that contains ".." without path separators to trigger
	// the handler's guard (strings.Contains(filename, "..")).
	fw, _ := w.CreateFormFile("file", "..evil..pkg.whl")
	fw.Write(makeTestWheel(t, seed.PackageName, "1.0.0"))
	w.Close()

	req := httptest.NewRequest("POST", "/api/v1/packages/"+seed.PackageName+"/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req = withChiParam(req, "name", seed.PackageName)
	req = addAuthContext(req, seed.UserID, []string{"publish"})

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for path traversal filename, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpload_MissingVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	w.WriteField("name", seed.PackageName)
	w.WriteField("version", "")
	fw, _ := w.CreateFormFile("file", "test-1.0.0.whl")
	fw.Write(makeTestWheel(t, seed.PackageName, "1.0.0"))
	w.Close()

	req := httptest.NewRequest("POST", "/api/v1/packages/"+seed.PackageName+"/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req = withChiParam(req, "name", seed.PackageName)
	req = addAuthContext(req, seed.UserID, []string{"publish"})

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing version, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── UpdatePackage ─────────────────────────────────────────────────────────

func TestUpdatePackage_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	body, _ := json.Marshal(map[string]interface{}{
		"display_name": "Updated Display Name",
		"description":  "Updated description text",
		"homepage":     "https://example.com",
		"repository":   "https://github.com/example/repo",
		"license":      "Apache-2.0",
		"keywords":     []string{"updated", "test"},
	})

	req := httptest.NewRequest("PUT", "/api/v1/packages/"+seed.PackageName, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withChiParam(req, "name", seed.PackageName)
	req = addAuthContext(req, seed.UserID, []string{"publish"})

	rr := httptest.NewRecorder()
	h.UpdatePackage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("UpdatePackage returned %d: %s", rr.Code, rr.Body.String())
	}

	var updated map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got, ok := updated["display_name"].(string); !ok || got != "Updated Display Name" {
		t.Errorf("display_name = %q, want %q", got, "Updated Display Name")
	}
	if got, ok := updated["description"].(string); !ok || got != "Updated description text" {
		t.Errorf("description = %q, want %q", got, "Updated description text")
	}
}

func TestUpdatePackage_NotOwner(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	otherUser := uuid.New()
	body, _ := json.Marshal(map[string]interface{}{
		"display_name": "Hacked",
		"description":  "Should not work",
	})

	req := httptest.NewRequest("PUT", "/api/v1/packages/"+seed.PackageName, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withChiParam(req, "name", seed.PackageName)
	req = addAuthContext(req, otherUser, []string{"publish"})

	rr := httptest.NewRecorder()
	h.UpdatePackage(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-owner, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── DeletePackage ─────────────────────────────────────────────────────────

func TestDeletePackage_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	// Create a fresh package specifically for deletion so we don't interfere
	// with other tests that rely on the seed package.
	pkgName := fmt.Sprintf("del-pkg-%s", uuid.New().String()[:8])
	createBody, _ := json.Marshal(map[string]string{
		"Name":        pkgName,
		"Description": "Package to delete",
	})

	createReq := httptest.NewRequest("POST", "/api/v1/packages", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = addAuthContext(createReq, seed.UserID, []string{"publish"})
	createRR := httptest.NewRecorder()
	h.CreatePackage(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("setup: CreatePackage returned %d: %s", createRR.Code, createRR.Body.String())
	}

	// Now delete it
	delReq := httptest.NewRequest("DELETE", "/api/v1/packages/"+pkgName, nil)
	delReq = withChiParam(delReq, "name", pkgName)
	delReq = addAuthContext(delReq, seed.UserID, []string{"publish"})

	delRR := httptest.NewRecorder()
	h.DeletePackage(delRR, delReq)

	if delRR.Code != http.StatusNoContent {
		t.Fatalf("DeletePackage returned %d: %s", delRR.Code, delRR.Body.String())
	}

	// Verify it's gone
	getReq := httptest.NewRequest("GET", "/api/v1/packages/"+pkgName, nil)
	getReq = withChiParam(getReq, "name", pkgName)
	getRR := httptest.NewRecorder()
	h.GetPackage(getRR, getReq)

	if getRR.Code != http.StatusNotFound {
		t.Errorf("expected 404 after deletion, got %d", getRR.Code)
	}
}

func TestDeletePackage_NotOwner(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	otherUser := uuid.New()
	req := httptest.NewRequest("DELETE", "/api/v1/packages/"+seed.PackageName, nil)
	req = withChiParam(req, "name", seed.PackageName)
	req = addAuthContext(req, otherUser, []string{"publish"})

	rr := httptest.NewRecorder()
	h.DeletePackage(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-owner, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── DeleteVersion ─────────────────────────────────────────────────────────

func TestDeleteVersion_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	req := httptest.NewRequest("DELETE", "/api/v1/packages/"+seed.PackageName+"/1.0.0", nil)
	req = withChiParams(req, map[string]string{
		"name":    seed.PackageName,
		"version": "1.0.0",
	})
	req = addAuthContext(req, seed.UserID, []string{"publish"})

	rr := httptest.NewRecorder()
	h.DeleteVersion(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("DeleteVersion returned %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeleteVersion_NotOwner(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	otherUser := uuid.New()
	req := httptest.NewRequest("DELETE", "/api/v1/packages/"+seed.PackageName+"/1.0.0", nil)
	req = withChiParams(req, map[string]string{
		"name":    seed.PackageName,
		"version": "1.0.0",
	})
	req = addAuthContext(req, otherUser, []string{"publish"})

	rr := httptest.NewRecorder()
	h.DeleteVersion(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-owner, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── GetVersionStatus ──────────────────────────────────────────────────────

func TestGetVersionStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	req := httptest.NewRequest("GET", "/api/v1/packages/"+seed.PackageName+"/1.0.0/status", nil)
	req = withChiParams(req, map[string]string{
		"name":    seed.PackageName,
		"version": "1.0.0",
	})

	rr := httptest.NewRecorder()
	h.GetVersionStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GetVersionStatus returned %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := resp["verification_status"]; !ok {
		t.Error("response missing verification_status field")
	}
	if status, ok := resp["verification_status"].(string); !ok || status != "passed" {
		t.Errorf("verification_status = %q, want %q", resp["verification_status"], "passed")
	}
	if _, ok := resp["published"]; !ok {
		t.Error("response missing published field")
	}
}

// ── Publish ───────────────────────────────────────────────────────────────

func TestPublish_NotImplemented(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, _, _ := newHandler(t)

	req := httptest.NewRequest("POST", "/api/v1/packages/some-pkg/publish", nil)
	rr := httptest.NewRecorder()
	h.Publish(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── SimpleIndex (PEP 503) ────────────────────────────────────────────────

func TestSimpleIndex_HTML(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	req := httptest.NewRequest("GET", "/simple/", nil)
	rr := httptest.NewRecorder()
	h.SimpleIndex(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("SimpleIndex returned %d: %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	body := rr.Body.String()
	if !strings.Contains(body, seed.PackageName) {
		t.Errorf("SimpleIndex should contain package %q, got:\n%s", seed.PackageName, body)
	}
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("SimpleIndex should contain DOCTYPE")
	}
}

func TestSimpleIndex_JSON(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	req := httptest.NewRequest("GET", "/simple/", nil)
	req.Header.Set("Accept", "application/vnd.pypi.simple.v1+json")
	rr := httptest.NewRecorder()
	h.SimpleIndex(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("SimpleIndex JSON returned %d: %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/vnd.pypi.simple.v1+json") {
		t.Errorf("Content-Type = %q, want pypi JSON", ct)
	}
	var resp struct {
		Projects []struct {
			Name string `json:"name"`
		} `json:"projects"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	found := false
	for _, p := range resp.Projects {
		if p.Name == seed.PackageName {
			found = true
		}
	}
	if !found {
		t.Errorf("expected package %q in JSON index", seed.PackageName)
	}
}

// ── SimplePackageIndex ───────────────────────────────────────────────────

func TestSimplePackageIndex_HTML(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	req := httptest.NewRequest("GET", "/simple/"+seed.PackageName+"/", nil)
	req = withChiParam(req, "name", seed.PackageName)
	rr := httptest.NewRecorder()
	h.SimplePackageIndex(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("SimplePackageIndex returned %d: %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(rr.Body.String(), "pypi:repository-version") {
		t.Error("missing pypi:repository-version meta tag")
	}
}

func TestSimplePackageIndex_JSON(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	req := httptest.NewRequest("GET", "/simple/"+seed.PackageName+"/", nil)
	req.Header.Set("Accept", "application/vnd.pypi.simple.v1+json")
	req = withChiParam(req, "name", seed.PackageName)
	rr := httptest.NewRecorder()
	h.SimplePackageIndex(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("SimplePackageIndex JSON returned %d: %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/vnd.pypi.simple.v1+json") {
		t.Errorf("Content-Type = %q, want pypi JSON", ct)
	}
	var resp struct {
		Name  string        `json:"name"`
		Files []interface{} `json:"files"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Name != seed.PackageName {
		t.Errorf("name = %q, want %q", resp.Name, seed.PackageName)
	}
}

func TestSimplePackageIndex_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, _, _ := newHandler(t)

	req := httptest.NewRequest("GET", "/simple/nonexistent-package/", nil)
	req = withChiParam(req, "name", "nonexistent-package")
	rr := httptest.NewRecorder()
	h.SimplePackageIndex(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── SimpleFileRedirect ───────────────────────────────────────────────────

func TestSimpleFileRedirect_NoStore(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	cfg := testutil.TestConfig()
	// nil store
	h := packages.NewHandler(q, nil, nil, nil, cfg)

	req := httptest.NewRequest("GET", "/simple/pkg/file.whl", nil)
	req = withChiParams(req, map[string]string{"name": "pkg", "filename": "file.whl"})
	rr := httptest.NewRecorder()
	h.SimpleFileRedirect(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 with nil store, got %d", rr.Code)
	}
}

func TestSimpleFileRedirect_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, _, _ := newHandler(t)

	req := httptest.NewRequest("GET", "/simple/nonexistent/file.whl", nil)
	req = withChiParams(req, map[string]string{"name": "nonexistent", "filename": "file.whl"})
	rr := httptest.NewRecorder()
	h.SimpleFileRedirect(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── TrustPackage ─────────────────────────────────────────────────────────

func TestTrustPackage_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	req := httptest.NewRequest("POST", "/api/v1/admin/packages/"+seed.PackageName+"/trust", nil)
	req = withChiParam(req, "name", seed.PackageName)
	rr := httptest.NewRecorder()
	h.TrustPackage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("TrustPackage returned %d: %s", rr.Code, rr.Body.String())
	}
}

func TestTrustPackage_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, _, _ := newHandler(t)

	req := httptest.NewRequest("POST", "/api/v1/admin/packages/nonexistent/trust", nil)
	req = withChiParam(req, "name", "nonexistent")
	rr := httptest.NewRecorder()
	h.TrustPackage(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── RemovePackage (admin) ────────────────────────────────────────────────

func TestRemovePackage_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	seed := testutil.SetupTestData(t, q)

	// Create a throwaway package for removal
	pkgName := fmt.Sprintf("rm-pkg-%s", uuid.New().String()[:8])
	createBody, _ := json.Marshal(map[string]string{"Name": pkgName, "Description": "to remove"})
	createReq := httptest.NewRequest("POST", "/api/v1/packages", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = addAuthContext(createReq, seed.UserID, []string{"publish"})
	createRR := httptest.NewRecorder()
	h.CreatePackage(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("setup: CreatePackage returned %d", createRR.Code)
	}

	req := httptest.NewRequest("DELETE", "/api/v1/admin/packages/"+pkgName, nil)
	req = withChiParam(req, "name", pkgName)
	rr := httptest.NewRecorder()
	h.RemovePackage(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("RemovePackage returned %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRemovePackage_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, _, _ := newHandler(t)

	req := httptest.NewRequest("DELETE", "/api/v1/admin/packages/nonexistent", nil)
	req = withChiParam(req, "name", "nonexistent")
	rr := httptest.NewRecorder()
	h.RemovePackage(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── VerifyPackage stub ───────────────────────────────────────────────────

func TestVerifyPackage_NotImplemented(t *testing.T) {
	h, _, _ := newHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/admin/packages/x/verify", nil)
	rr := httptest.NewRecorder()
	h.VerifyPackage(rr, req)
	if rr.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

// ── MyDownloads / PackageDownloads stubs ─────────────────────────────────

func TestMyDownloads_NotImplemented(t *testing.T) {
	h, _, _ := newHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/me/downloads", nil)
	rr := httptest.NewRecorder()
	h.MyDownloads(rr, req)
	if rr.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

func TestPackageDownloads_NotImplemented(t *testing.T) {
	h, _, _ := newHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/packages/x/downloads", nil)
	rr := httptest.NewRecorder()
	h.PackageDownloads(rr, req)
	if rr.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

// ── VerificationQueue ────────────────────────────────────────────────────

func TestVerificationQueue(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, _, _ := newHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/admin/verification-queue", nil)
	rr := httptest.NewRecorder()
	h.VerificationQueue(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("VerificationQueue returned %d: %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

// ── GetPresignedURLForFile ───────────────────────────────────────────────

func TestGetPresignedURLForFile_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	store := testutil.NewMockStorage()

	info, err := packages.GetPresignedURLForFile(context.Background(), q, "nonexistent-pkg", "file.whl", store)
	if err == nil && info != nil {
		t.Error("expected nil info for nonexistent package")
	}
}

func TestGetPresignedURLForFile_NoMatchingFile(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	store := testutil.NewMockStorage()
	seed := testutil.SetupTestData(t, q)

	info, err := packages.GetPresignedURLForFile(context.Background(), q, seed.PackageName, "nonexistent-file.whl", store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info != nil {
		t.Error("expected nil info for nonexistent file")
	}
}

// ── Service layer ────────────────────────────────────────────────────────

func TestService_CreateVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	store := testutil.NewMockStorage()
	verifier := packages.NewVerifier(q, 0)
	svc := packages.NewService(q, store, verifier)

	ver := fmt.Sprintf("9.0.%s", uuid.New().String()[:4])
	pv, err := svc.CreateVersion(context.Background(), seed.PackageID, ver, ">=3.12", "", seed.UserID)
	if err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}
	if pv.Version != ver {
		t.Errorf("version = %q, want %q", pv.Version, ver)
	}
	if pv.RequiresPython.String != ">=3.12" {
		t.Errorf("requires_python = %q, want %q", pv.RequiresPython.String, ">=3.12")
	}
}

func TestService_AddFile(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	store := testutil.NewMockStorage()
	verifier := packages.NewVerifier(q, 0)
	svc := packages.NewService(q, store, verifier)

	// Create a version first
	ver := fmt.Sprintf("9.1.%s", uuid.New().String()[:4])
	pv, err := svc.CreateVersion(context.Background(), seed.PackageID, ver, "", "", seed.UserID)
	if err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}

	hash := sha256.Sum256([]byte("test content"))
	hashHex := fmt.Sprintf("%x", hash)
	f, err := svc.AddFile(context.Background(), pv.ID, "test-pkg-1.0.0.whl", "bdist_wheel", "cp312", "cp312", "any", 1024, hashHex, "", "packages/test.whl")
	if err != nil {
		t.Fatalf("AddFile: %v", err)
	}
	if f.Filename != "test-pkg-1.0.0.whl" {
		t.Errorf("filename = %q, want %q", f.Filename, "test-pkg-1.0.0.whl")
	}
}

// ── Verifier ─────────────────────────────────────────────────────────────

func TestVerifier_RunVerification(t *testing.T) {
	q := testutil.Queries(t)
	v := packages.NewVerifier(q, 2)
	// Stub should return nil (no-op)
	if err := v.RunVerification(context.Background(), uuid.New()); err != nil {
		t.Errorf("RunVerification stub should return nil, got: %v", err)
	}
}

// ── Signer ───────────────────────────────────────────────────────────────

func TestSigner_SignAndStore(t *testing.T) {
	q := testutil.Queries(t)
	s := packages.NewSigner(q)
	bundle, err := s.SignAndStore(context.Background(), uuid.New(), "/tmp/artifact.whl")
	if err != nil {
		t.Errorf("SignAndStore stub should return nil, got: %v", err)
	}
	if bundle != nil {
		t.Errorf("SignAndStore stub should return nil bundle, got: %v", bundle)
	}
}

func TestSigner_StoreBundle(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	s := packages.NewSigner(q)

	err := s.StoreBundle(context.Background(), seed.VersionID, []byte(`{"log":{}}`))
	if err != nil {
		t.Errorf("StoreBundle: %v", err)
	}
}

// ── NormalizeName edge cases ─────────────────────────────────────────────

func TestNormalizeName_Exhaustive(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"My_Package", "my-package"},
		{"MY.PACKAGE", "my-package"},
		{"a__b--c..d", "a--b--c--d"},
		{"---abc---", "abc"},
		{"", ""},
		{"A", "a"},
		{"a-b.c_d", "a-b-c-d"},
	}
	for _, tc := range cases {
		got := packages.NormalizeName(tc.in)
		if got != tc.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// ── ListOrgPackages ──────────────────────────────────────────────────────

func TestListOrgPackages(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	ctx := context.Background()
	suffix := uuid.New().String()[:8]

	// Create an org
	orgName := fmt.Sprintf("testorg-%s", suffix)
	org, err := q.CreateOrg(ctx, db.CreateOrgParams{
		Name:         orgName,
		DisplayName:  orgName,
		BillingEmail: pgtype.Text{},
	})
	if err != nil {
		t.Fatalf("create org: %v", err)
	}
	t.Cleanup(func() {
		pool := testutil.Pool(t)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", org.ID)
	})

	// Create a user for ownership
	user, err := q.CreateUser(ctx, db.CreateUserParams{
		Email:    fmt.Sprintf("orgtest-%s@example.com", suffix),
		Username: fmt.Sprintf("orguser-%s", suffix),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Create a package under this org namespace
	pkgName := fmt.Sprintf("orgpkg-%s", suffix)
	pkg, err := q.CreatePackage(ctx, db.CreatePackageParams{
		Name:        pkgName,
		Namespace:   pgtype.Text{String: orgName, Valid: true},
		DisplayName: pkgName,
		Description: pgtype.Text{String: "org pkg", Valid: true},
		Homepage:    pgtype.Text{},
		Repository:  pgtype.Text{},
		License:     pgtype.Text{String: "MIT", Valid: true},
		Keywords:    []string{"test"},
		OwnerUserID: pgtype.UUID{Bytes: user.ID, Valid: true},
		OwnerOrgID:  pgtype.UUID{Bytes: org.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("create package: %v", err)
	}
	t.Cleanup(func() {
		_ = q.DeletePackage(ctx, pkg.ID)
	})

	req := httptest.NewRequest("GET", "/api/v1/orgs/"+orgName+"/packages", nil)
	req = withChiParam(req, "slug", orgName)
	rr := httptest.NewRecorder()
	h.ListOrgPackages(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListOrgPackages returned %d: %s", rr.Code, rr.Body.String())
	}

	var pkgs []json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &pkgs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(pkgs) < 1 {
		t.Error("expected at least 1 package in org listing")
	}
}

func TestListOrgPackages_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, _, _ := newHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/orgs/nonexistent-org-xyz/packages", nil)
	req = withChiParam(req, "slug", "nonexistent-org-xyz")
	rr := httptest.NewRecorder()
	h.ListOrgPackages(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestListOrgPackages_EmptyOrg(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	h, q, _ := newHandler(t)
	ctx := context.Background()
	suffix := uuid.New().String()[:8]

	// Create an org with no packages
	orgName := fmt.Sprintf("emptyorg-%s", suffix)
	org, err := q.CreateOrg(ctx, db.CreateOrgParams{
		Name:         orgName,
		DisplayName:  orgName,
		BillingEmail: pgtype.Text{},
	})
	if err != nil {
		t.Fatalf("create org: %v", err)
	}
	t.Cleanup(func() {
		pool := testutil.Pool(t)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", org.ID)
	})

	req := httptest.NewRequest("GET", "/api/v1/orgs/"+orgName+"/packages", nil)
	req = withChiParam(req, "slug", orgName)
	rr := httptest.NewRecorder()
	h.ListOrgPackages(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var pkgs []json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &pkgs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d", len(pkgs))
	}
}
