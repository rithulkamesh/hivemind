package audit_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/rithul/hivemind/registry/api/internal/audit"
	"github.com/rithul/hivemind/registry/api/internal/testutil"
)

func TestAuditLog_WritesEntry(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)

	userID := seed.UserID
	id, err := audit.Log(context.Background(), q, &userID, nil, "publish", "package", seed.PackageName, "abc123", map[string]string{"version": "1.0.0"})
	if err != nil {
		t.Fatalf("audit.Log: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero audit log ID")
	}
}

func TestAuditLog_WithAPIKey(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)

	userID := seed.UserID
	keyID := seed.APIKeyID
	id, err := audit.Log(context.Background(), q, &userID, &keyID, "login", "api_key", keyID.String(), "", nil)
	if err != nil {
		t.Fatalf("audit.Log: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero audit log ID")
	}
}

func TestAuditLog_NilActors(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)

	// Both actor user and API key can be nil (system-initiated events)
	id, err := audit.Log(context.Background(), q, nil, nil, "system.cleanup", "", "", "", nil)
	if err != nil {
		t.Fatalf("audit.Log with nil actors: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero audit log ID")
	}
}

func TestAuditLog_NilMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	uid := uuid.New()

	// Create a user first so the FK isn't violated — or use nil user
	id, err := audit.Log(context.Background(), q, nil, nil, "test.nil-meta", "test", uid.String(), "hash", nil)
	if err != nil {
		t.Fatalf("audit.Log with nil metadata: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero audit log ID")
	}
}
