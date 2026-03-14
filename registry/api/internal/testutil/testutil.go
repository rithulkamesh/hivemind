// Package testutil provides shared helpers for integration tests.
//
// Tests require a running PostgreSQL (the same one from docker-compose.dev.yml).
// Set DATABASE_URL or it defaults to the dev compose DB.
package testutil

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rithul/hivemind/registry/api/internal/config"
	"github.com/rithul/hivemind/registry/api/internal/db"
)

const defaultDatabaseURL = "postgres://registry:registry@localhost:5432/hivemind_registry?sslmode=disable"

var (
	poolOnce sync.Once
	pool     *pgxpool.Pool
	poolErr  error
)

// Pool returns a shared pgxpool.Pool for the test database.
// Calls t.Fatal if connection fails.
func Pool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	poolOnce.Do(func() {
		url := os.Getenv("DATABASE_URL")
		if url == "" {
			url = defaultDatabaseURL
		}
		pool, poolErr = pgxpool.New(context.Background(), url)
	})
	if poolErr != nil {
		t.Fatalf("testutil.Pool: %v (is Docker Postgres running?)", poolErr)
	}
	return pool
}

// Queries returns *db.Queries backed by the test pool.
func Queries(t *testing.T) *db.Queries {
	t.Helper()
	return db.New(Pool(t))
}

// TestConfig returns a minimal config suitable for handler tests.
func TestConfig() *config.Config {
	return &config.Config{
		JWTSecret:       "test-secret-key-for-tests",
		BaseURL:         "http://localhost:8080",
		Port:            "8080",
		MaxUploadSizeMB: 100,
	}
}

// Seed holds references to test data created by SetupTestData.
type Seed struct {
	UserID      uuid.UUID
	UserEmail   string
	APIKey      string // raw key (hm_...)
	APIKeyID    uuid.UUID
	PackageID   uuid.UUID
	PackageName string
	VersionID   uuid.UUID
}

// SetupTestData inserts a user, API key, and package into the test DB.
// It uses a unique suffix to avoid conflicts between parallel tests.
// Cleanup is registered via t.Cleanup.
func SetupTestData(t *testing.T, q *db.Queries) *Seed {
	t.Helper()
	ctx := context.Background()
	suffix := uuid.New().String()[:8]

	// Create user
	user, err := q.CreateUser(ctx, db.CreateUserParams{
		Email:    fmt.Sprintf("test-%s@example.com", suffix),
		Username: fmt.Sprintf("testuser-%s", suffix),
	})
	if err != nil {
		t.Fatalf("SetupTestData: create user: %v", err)
	}

	// Create API key
	rawKey := fmt.Sprintf("hm_test_%s_%s", suffix, "abcdef0123456789abcdef0123456789")
	keyHash := hashKey(rawKey)
	prefix := rawKey[:11] + "..."
	apiKey, err := q.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		UserID:    user.ID,
		OrgID:     pgtype.UUID{},
		Name:      "test-key",
		KeyHash:   keyHash,
		KeyPrefix: prefix,
		Scopes:    []string{"publish", "read"},
		ExpiresAt: pgtype.Timestamptz{},
	})
	if err != nil {
		t.Fatalf("SetupTestData: create api key: %v", err)
	}

	// Create package
	pkgName := fmt.Sprintf("test-pkg-%s", suffix)
	pkg, err := q.CreatePackage(ctx, db.CreatePackageParams{
		Name:        pkgName,
		Namespace:   pgtype.Text{},
		DisplayName: pkgName,
		Description: pgtype.Text{String: "A test package", Valid: true},
		Homepage:    pgtype.Text{},
		Repository:  pgtype.Text{},
		License:     pgtype.Text{String: "MIT", Valid: true},
		Keywords:    []string{"test"},
		OwnerUserID: pgtype.UUID{Bytes: user.ID, Valid: true},
		OwnerOrgID:  pgtype.UUID{},
	})
	if err != nil {
		t.Fatalf("SetupTestData: create package: %v", err)
	}

	// Create version
	pv, err := q.CreatePackageVersion(ctx, db.CreatePackageVersionParams{
		PackageID:          pkg.ID,
		Version:            "1.0.0",
		RequiresPython:     pgtype.Text{String: ">=3.12", Valid: true},
		RequiresHivemind:   pgtype.Text{},
		UploadedBy:         pgtype.UUID{Bytes: user.ID, Valid: true},
		VerificationStatus: pgtype.Text{String: "passed", Valid: true},
	})
	if err != nil {
		t.Fatalf("SetupTestData: create version: %v", err)
	}

	// Mark published
	_, err = q.UpdatePackageVersionVerification(ctx, db.UpdatePackageVersionVerificationParams{
		ID:                 pv.ID,
		VerificationStatus: pgtype.Text{String: "passed", Valid: true},
		Published:          pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		t.Fatalf("SetupTestData: publish version: %v", err)
	}

	seed := &Seed{
		UserID:      user.ID,
		UserEmail:   user.Email,
		APIKey:      rawKey,
		APIKeyID:    apiKey.ID,
		PackageID:   pkg.ID,
		PackageName: pkgName,
		VersionID:   pv.ID,
	}

	t.Cleanup(func() {
		// Best-effort cleanup
		_ = q.DeletePackageVersion(ctx, pv.ID)
		_ = q.DeletePackage(ctx, pkg.ID)
		_ = q.RevokeAPIKey(ctx, apiKey.ID)
		// Note: no DeleteUser query exists; rows are harmless with unique suffixes.
	})

	return seed
}

func hashKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// ── Mock Storage ──────────────────────────────────────────────────────────

// MockStorage is an in-memory implementation of packages.Storage for tests.
type MockStorage struct {
	mu    sync.Mutex
	Files map[string][]byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{Files: make(map[string][]byte)}
}

func (m *MockStorage) Upload(key string, body []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Files[key] = body
	return nil
}

func (m *MockStorage) PresignedDownloadURL(key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.Files[key]; !ok {
		return "", fmt.Errorf("not found: %s", key)
	}
	return fmt.Sprintf("http://mock-s3/download/%s", key), nil
}

// ── Auth Context Helpers ──────────────────────────────────────────────────

// Note: The auth package uses unexported contextKey type. To inject context values
// in tests, import auth and use auth.ContextKeyUserID etc. directly.
// The helpers below are kept for convenience but callers in _test packages
// should use auth.ContextKey* constants for type safety.
