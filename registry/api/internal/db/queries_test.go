package db_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/db"
	"github.com/rithul/hivemind/registry/api/internal/testutil"
)

// helper to create a unique org and return it alongside cleanup.
func createTestOrg(t *testing.T, q *db.Queries) db.Organization {
	t.Helper()
	ctx := context.Background()
	suffix := uuid.New().String()[:8]
	org, err := q.CreateOrg(ctx, db.CreateOrgParams{
		Name:         fmt.Sprintf("org-%s", suffix),
		DisplayName:  fmt.Sprintf("Org %s", suffix),
		BillingEmail: pgtype.Text{String: fmt.Sprintf("billing-%s@example.com", suffix), Valid: true},
	})
	if err != nil {
		t.Fatalf("createTestOrg: %v", err)
	}
	return org
}

// helper to create a package file for the seed version.
func createTestFile(t *testing.T, q *db.Queries, versionID uuid.UUID) db.PackageFile {
	t.Helper()
	ctx := context.Background()
	suffix := uuid.New().String()[:8]
	h := sha256.Sum256([]byte("content-" + suffix))
	shaHex := hex.EncodeToString(h[:])
	pf, err := q.CreatePackageFile(ctx, db.CreatePackageFileParams{
		VersionID:     versionID,
		Filename:      fmt.Sprintf("pkg-%s-1.0.0.tar.gz", suffix),
		Filetype:      "sdist",
		PythonVersion: pgtype.Text{String: "source", Valid: true},
		Abi:           pgtype.Text{},
		Platform:      pgtype.Text{},
		SizeBytes:     1024,
		Sha256:        shaHex,
		Md5:           shaHex[:32],
		S3Key:         fmt.Sprintf("packages/test/%s/pkg.tar.gz", suffix),
	})
	if err != nil {
		t.Fatalf("createTestFile: %v", err)
	}
	return pf
}

// ── 1. CreateOAuthIdentity ───────────────────────────────────────────────

func TestCreateOAuthIdentity(t *testing.T) {
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	ctx := context.Background()
	suffix := uuid.New().String()[:8]

	oi, err := q.CreateOAuthIdentity(ctx, db.CreateOAuthIdentityParams{
		UserID:         seed.UserID,
		Provider:       "github",
		ProviderUserID: fmt.Sprintf("gh-%s", suffix),
		ProviderEmail:  pgtype.Text{String: fmt.Sprintf("oauth-%s@github.com", suffix), Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateOAuthIdentity: %v", err)
	}
	if oi.Provider != "github" {
		t.Fatalf("expected provider github, got %s", oi.Provider)
	}
	if oi.UserID != seed.UserID {
		t.Fatalf("expected user_id %v, got %v", seed.UserID, oi.UserID)
	}
}

// ── 2. GetOAuthIdentity ──────────────────────────────────────────────────

func TestGetOAuthIdentity(t *testing.T) {
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	ctx := context.Background()
	suffix := uuid.New().String()[:8]
	providerUID := fmt.Sprintf("gh-%s", suffix)

	// Create first
	_, err := q.CreateOAuthIdentity(ctx, db.CreateOAuthIdentityParams{
		UserID:         seed.UserID,
		Provider:       "github",
		ProviderUserID: providerUID,
		ProviderEmail:  pgtype.Text{String: "x@gh.com", Valid: true},
	})
	if err != nil {
		t.Fatalf("setup CreateOAuthIdentity: %v", err)
	}

	// Found case
	oi, err := q.GetOAuthIdentity(ctx, db.GetOAuthIdentityParams{
		Provider:       "github",
		ProviderUserID: providerUID,
	})
	if err != nil {
		t.Fatalf("GetOAuthIdentity (found): %v", err)
	}
	if oi.ProviderUserID != providerUID {
		t.Fatalf("expected provider_user_id %s, got %s", providerUID, oi.ProviderUserID)
	}

	// Not-found case
	_, err = q.GetOAuthIdentity(ctx, db.GetOAuthIdentityParams{
		Provider:       "github",
		ProviderUserID: "nonexistent-user-id",
	})
	if err == nil {
		t.Fatal("GetOAuthIdentity: expected error for not-found, got nil")
	}
}

// ── 3. GetOrgByID ────────────────────────────────────────────────────────

func TestGetOrgByID(t *testing.T) {
	q := testutil.Queries(t)
	ctx := context.Background()

	org := createTestOrg(t, q)

	// Found case
	got, err := q.GetOrgByID(ctx, org.ID)
	if err != nil {
		t.Fatalf("GetOrgByID (found): %v", err)
	}
	if got.Name != org.Name {
		t.Fatalf("expected name %s, got %s", org.Name, got.Name)
	}

	// Not-found case
	_, err = q.GetOrgByID(ctx, uuid.New())
	if err == nil {
		t.Fatal("GetOrgByID: expected error for not-found, got nil")
	}
}

// ── 4. GetOrgMember ──────────────────────────────────────────────────────

func TestGetOrgMember(t *testing.T) {
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	ctx := context.Background()

	org := createTestOrg(t, q)

	// Add member
	_, err := q.AddOrgMember(ctx, db.AddOrgMemberParams{
		OrgID:  org.ID,
		UserID: seed.UserID,
		Role:   "member",
	})
	if err != nil {
		t.Fatalf("AddOrgMember: %v", err)
	}

	// Found case
	m, err := q.GetOrgMember(ctx, db.GetOrgMemberParams{
		OrgID:  org.ID,
		UserID: seed.UserID,
	})
	if err != nil {
		t.Fatalf("GetOrgMember (found): %v", err)
	}
	if m.Role != "member" {
		t.Fatalf("expected role member, got %s", m.Role)
	}

	// Not-found case
	_, err = q.GetOrgMember(ctx, db.GetOrgMemberParams{
		OrgID:  org.ID,
		UserID: uuid.New(),
	})
	if err == nil {
		t.Fatal("GetOrgMember: expected error for not-found, got nil")
	}
}

// ── 5. GetPackageByID ────────────────────────────────────────────────────

func TestGetPackageByID(t *testing.T) {
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	ctx := context.Background()

	// Found case
	pkg, err := q.GetPackageByID(ctx, seed.PackageID)
	if err != nil {
		t.Fatalf("GetPackageByID (found): %v", err)
	}
	if pkg.Name != seed.PackageName {
		t.Fatalf("expected name %s, got %s", seed.PackageName, pkg.Name)
	}

	// Not-found case
	_, err = q.GetPackageByID(ctx, uuid.New())
	if err == nil {
		t.Fatal("GetPackageByID: expected error for not-found, got nil")
	}
}

// ── 6. GetPackageFileByID ────────────────────────────────────────────────

func TestGetPackageFileByID(t *testing.T) {
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	ctx := context.Background()

	pf := createTestFile(t, q, seed.VersionID)

	// Found case
	got, err := q.GetPackageFileByID(ctx, pf.ID)
	if err != nil {
		t.Fatalf("GetPackageFileByID (found): %v", err)
	}
	if got.Filename != pf.Filename {
		t.Fatalf("expected filename %s, got %s", pf.Filename, got.Filename)
	}

	// Not-found case
	_, err = q.GetPackageFileByID(ctx, uuid.New())
	if err == nil {
		t.Fatal("GetPackageFileByID: expected error for not-found, got nil")
	}
}

// ── 7. GetUserByEmail ────────────────────────────────────────────────────

func TestGetUserByEmail(t *testing.T) {
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	ctx := context.Background()

	// Found case
	u, err := q.GetUserByEmail(ctx, seed.UserEmail)
	if err != nil {
		t.Fatalf("GetUserByEmail (found): %v", err)
	}
	if u.ID != seed.UserID {
		t.Fatalf("expected user ID %v, got %v", seed.UserID, u.ID)
	}

	// Not-found case
	_, err = q.GetUserByEmail(ctx, "nonexistent@example.com")
	if err == nil {
		t.Fatal("GetUserByEmail: expected error for not-found, got nil")
	}
}

// ── 8. GetUserByUsername ─────────────────────────────────────────────────

func TestGetUserByUsername(t *testing.T) {
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	ctx := context.Background()

	// Get user to find the username
	u, err := q.GetUserByEmail(ctx, seed.UserEmail)
	if err != nil {
		t.Fatalf("setup GetUserByEmail: %v", err)
	}

	// Found case
	got, err := q.GetUserByUsername(ctx, u.Username)
	if err != nil {
		t.Fatalf("GetUserByUsername (found): %v", err)
	}
	if got.ID != seed.UserID {
		t.Fatalf("expected user ID %v, got %v", seed.UserID, got.ID)
	}

	// Not-found case
	_, err = q.GetUserByUsername(ctx, "nonexistent-user-xyz")
	if err == nil {
		t.Fatal("GetUserByUsername: expected error for not-found, got nil")
	}
}

// ── 9. GetVersionByID ────────────────────────────────────────────────────

func TestGetVersionByID(t *testing.T) {
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	ctx := context.Background()

	// Found case
	v, err := q.GetVersionByID(ctx, seed.VersionID)
	if err != nil {
		t.Fatalf("GetVersionByID (found): %v", err)
	}
	if v.Version != "1.0.0" {
		t.Fatalf("expected version 1.0.0, got %s", v.Version)
	}
	if v.PackageID != seed.PackageID {
		t.Fatalf("expected package ID %v, got %v", seed.PackageID, v.PackageID)
	}

	// Not-found case
	_, err = q.GetVersionByID(ctx, uuid.New())
	if err == nil {
		t.Fatal("GetVersionByID: expected error for not-found, got nil")
	}
}

// ── 10. IncrementFileDownloadCount ───────────────────────────────────────

func TestIncrementFileDownloadCount(t *testing.T) {
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	ctx := context.Background()

	pf := createTestFile(t, q, seed.VersionID)

	// Get initial count
	before, err := q.GetPackageFileByID(ctx, pf.ID)
	if err != nil {
		t.Fatalf("GetPackageFileByID (before): %v", err)
	}

	// Increment
	err = q.IncrementFileDownloadCount(ctx, pf.ID)
	if err != nil {
		t.Fatalf("IncrementFileDownloadCount: %v", err)
	}

	// Verify increment
	after, err := q.GetPackageFileByID(ctx, pf.ID)
	if err != nil {
		t.Fatalf("GetPackageFileByID (after): %v", err)
	}
	if after.DownloadCount.Int64 != before.DownloadCount.Int64+1 {
		t.Fatalf("expected download_count %d, got %d", before.DownloadCount.Int64+1, after.DownloadCount.Int64)
	}
}

// ── 11. IncrementPackageDownloadCount ────────────────────────────────────

func TestIncrementPackageDownloadCount(t *testing.T) {
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	ctx := context.Background()

	// Get initial count
	before, err := q.GetPackageByID(ctx, seed.PackageID)
	if err != nil {
		t.Fatalf("GetPackageByID (before): %v", err)
	}

	// Increment
	err = q.IncrementPackageDownloadCount(ctx, seed.PackageID)
	if err != nil {
		t.Fatalf("IncrementPackageDownloadCount: %v", err)
	}

	// Verify increment
	after, err := q.GetPackageByID(ctx, seed.PackageID)
	if err != nil {
		t.Fatalf("GetPackageByID (after): %v", err)
	}
	if after.TotalDownloads.Int64 != before.TotalDownloads.Int64+1 {
		t.Fatalf("expected total_downloads %d, got %d", before.TotalDownloads.Int64+1, after.TotalDownloads.Int64)
	}
}

// ── 12. RecordDownloadEvent ──────────────────────────────────────────────

func TestRecordDownloadEvent(t *testing.T) {
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	ctx := context.Background()

	pf := createTestFile(t, q, seed.VersionID)

	id, err := q.RecordDownloadEvent(ctx, db.RecordDownloadEventParams{
		FileID:      pf.ID,
		IpHash:      pgtype.Text{String: "abc123hash", Valid: true},
		UserAgent:   pgtype.Text{String: "pip/23.0", Valid: true},
		CountryCode: pgtype.Text{String: "US", Valid: true},
		Installer:   pgtype.Text{String: "pip", Valid: true},
	})
	if err != nil {
		t.Fatalf("RecordDownloadEvent: %v", err)
	}
	if !id.Valid || id.Int64 <= 0 {
		t.Fatalf("expected valid positive id, got %+v", id)
	}
}

// ── 13. UpdateAPIKeyLastUsed ─────────────────────────────────────────────

func TestUpdateAPIKeyLastUsed(t *testing.T) {
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	ctx := context.Background()

	err := q.UpdateAPIKeyLastUsed(ctx, seed.APIKeyID)
	if err != nil {
		t.Fatalf("UpdateAPIKeyLastUsed: %v", err)
	}

	// Verify by fetching the key — use the raw key hash to look it up.
	h := sha256.Sum256([]byte(seed.APIKey))
	keyHash := hex.EncodeToString(h[:])
	key, err := q.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		t.Fatalf("GetAPIKeyByHash after update: %v", err)
	}
	if !key.LastUsedAt.Valid {
		t.Fatal("expected last_used_at to be set after UpdateAPIKeyLastUsed")
	}
}

// ── 14. UpdateOrgMemberRole ──────────────────────────────────────────────

func TestUpdateOrgMemberRole(t *testing.T) {
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	ctx := context.Background()

	org := createTestOrg(t, q)

	// Add member with role "member"
	_, err := q.AddOrgMember(ctx, db.AddOrgMemberParams{
		OrgID:  org.ID,
		UserID: seed.UserID,
		Role:   "member",
	})
	if err != nil {
		t.Fatalf("AddOrgMember: %v", err)
	}

	// Update role to "admin"
	updated, err := q.UpdateOrgMemberRole(ctx, db.UpdateOrgMemberRoleParams{
		OrgID:  org.ID,
		UserID: seed.UserID,
		Role:   "admin",
	})
	if err != nil {
		t.Fatalf("UpdateOrgMemberRole: %v", err)
	}
	if updated.Role != "admin" {
		t.Fatalf("expected role admin, got %s", updated.Role)
	}

	// Verify via GetOrgMember
	m, err := q.GetOrgMember(ctx, db.GetOrgMemberParams{
		OrgID:  org.ID,
		UserID: seed.UserID,
	})
	if err != nil {
		t.Fatalf("GetOrgMember after update: %v", err)
	}
	if m.Role != "admin" {
		t.Fatalf("expected role admin after update, got %s", m.Role)
	}
}
