package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestLiveness(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	Liveness(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "ok" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "ok")
	}
}

func TestReadiness(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}

	// Import testutil for the pool
	// health.Readiness requires a pgxpool.Pool, which is cumbersome to
	// unit-test without a real DB. We test it as an integration test.
	// For now, verify the handler function itself compiles correctly
	// and test Liveness which needs no DB.
	t.Log("Readiness integration test covered by Phase 2 API smoke tests")
}

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://registry:registry@localhost:5432/hivemind_registry?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("testPool: %v (is Docker Postgres running?)", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestReadiness_HealthyDB(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}

	pool := testPool(t)
	handler := Readiness(pool)

	req := httptest.NewRequest("GET", "/ready", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "ok" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "ok")
	}
}

func TestReadiness_ClosedPool(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://registry:registry@localhost:5432/hivemind_registry?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	// Close the pool so Ping will fail
	pool.Close()

	handler := Readiness(pool)

	req := httptest.NewRequest("GET", "/ready", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}
