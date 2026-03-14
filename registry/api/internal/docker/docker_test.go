package docker_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/db"
	"github.com/rithul/hivemind/registry/api/internal/docker"
	"github.com/rithul/hivemind/registry/api/internal/testutil"
)

func TestListImages_EmptyReturnsArray(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := docker.NewHandler(q)

	// Use chi router for proper URL param injection
	r := chi.NewRouter()
	r.Get("/api/v1/packages/{name}/docker", h.ListImages)

	req := httptest.NewRequest("GET", "/api/v1/packages/"+seed.PackageName+"/docker", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListImages returned %d: %s", rr.Code, rr.Body.String())
	}

	var images []json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &images); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Should be empty array, not null
	if rr.Body.String() == "null\n" {
		t.Error("expected empty array, got null")
	}
}

func TestListImages_PackageNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := docker.NewHandler(q)

	r := chi.NewRouter()
	r.Get("/api/v1/packages/{name}/docker", h.ListImages)

	req := httptest.NewRequest("GET", "/api/v1/packages/nonexistent-pkg-xyz/docker", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestRegisterImage_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	svc := docker.NewService(q)

	size := int64(12345)
	err := svc.RegisterImage(
		context.Background(),
		seed.PackageID,
		"latest",
		"sha256:abc123",
		"123456789.dkr.ecr.us-east-1.amazonaws.com/test:latest",
		[]string{"linux/amd64"},
		&size,
	)
	if err != nil {
		t.Fatalf("RegisterImage: %v", err)
	}
}

func TestNewService(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	svc := docker.NewService(q)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestNewHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := docker.NewHandler(q)
	if h == nil {
		t.Fatal("NewHandler returned nil")
	}
}

func TestRegisterImage_And_ListImages(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)

	// Register a docker image directly via DB
	size := int64(5000)
	var sz pgtype.Int8
	sz.Int64 = size
	sz.Valid = true
	img, err := q.CreateDockerImage(context.Background(), db.CreateDockerImageParams{
		PackageID: seed.PackageID,
		Tag:       "v1.0.0",
		Digest:    "sha256:deadbeef",
		EcrUri:    "123456789.dkr.ecr.us-east-1.amazonaws.com/test:v1.0.0",
		Platform:  []string{"linux/amd64"},
		SizeBytes: sz,
	})
	if err != nil {
		t.Fatalf("CreateDockerImage: %v", err)
	}
	_ = img

	// Now list via handler
	h := docker.NewHandler(q)
	r := chi.NewRouter()
	r.Get("/api/v1/packages/{name}/docker", h.ListImages)

	req := httptest.NewRequest("GET", "/api/v1/packages/"+seed.PackageName+"/docker", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListImages returned %d: %s", rr.Code, rr.Body.String())
	}

	var images []json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &images); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(images) < 1 {
		t.Errorf("expected at least 1 image, got %d", len(images))
	}
}
