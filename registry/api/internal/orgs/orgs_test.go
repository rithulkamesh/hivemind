package orgs_test

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
	"github.com/rithul/hivemind/registry/api/internal/db"
	"github.com/rithul/hivemind/registry/api/internal/orgs"
	"github.com/rithul/hivemind/registry/api/internal/testutil"
)

// withUserID returns a context with the authenticated user ID set.
func withUserID(ctx context.Context, uid uuid.UUID) context.Context {
	return context.WithValue(ctx, auth.ContextKeyUserID, uid)
}

// orgRouter sets up a chi router matching the real registration pattern.
func orgRouter(h *orgs.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Route("/api/v1/orgs", func(r chi.Router) {
		r.Get("/", h.ListOrgs)
		r.Post("/", h.CreateOrg)
		r.Route("/{slug}", func(r chi.Router) {
			r.Get("/", h.GetOrg)
			r.Patch("/", h.UpdateOrg)
			r.Get("/members", h.ListMembers)
			r.Post("/members/invite", h.InviteMember)
			r.Delete("/members/{userID}", h.RemoveMember)
			r.Post("/sso", h.ConfigureSSO)
		})
	})
	return r
}

func TestCreateOrg_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	suffix := uuid.New().String()[:8]
	orgName := fmt.Sprintf("test-org-%s", suffix)
	body, _ := json.Marshal(map[string]string{
		"name":         orgName,
		"display_name": "Test Org " + suffix,
	})

	req := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader(body))
	req = req.WithContext(withUserID(req.Context(), seed.UserID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("CreateOrg: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var org db.Organization
	if err := json.Unmarshal(rr.Body.Bytes(), &org); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if org.Name != orgName {
		t.Errorf("org name = %q, want %q", org.Name, orgName)
	}

	// Cleanup
	t.Cleanup(func() {
		ctx := context.Background()
		_ = q.RemoveOrgMember(ctx, db.RemoveOrgMemberParams{OrgID: org.ID, UserID: seed.UserID})
		// No DeleteOrg query, use pool directly
		pool := testutil.Pool(t)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", org.ID)
	})
}

func TestCreateOrg_NoAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	body, _ := json.Marshal(map[string]string{
		"name":         "no-auth-org",
		"display_name": "No Auth",
	})

	req := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader(body))
	// No user ID in context
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", rr.Code)
	}
}

func TestCreateOrg_BadJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	req := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader([]byte("not json")))
	req = req.WithContext(withUserID(req.Context(), seed.UserID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad JSON, got %d", rr.Code)
	}
}

func TestCreateOrg_DuplicateName(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	suffix := uuid.New().String()[:8]
	orgName := fmt.Sprintf("dup-org-%s", suffix)
	body, _ := json.Marshal(map[string]string{
		"name":         orgName,
		"display_name": "Dup Org",
	})

	// First create should succeed
	req := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader(body))
	req = req.WithContext(withUserID(req.Context(), seed.UserID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var org db.Organization
	_ = json.Unmarshal(rr.Body.Bytes(), &org)
	t.Cleanup(func() {
		ctx := context.Background()
		_ = q.RemoveOrgMember(ctx, db.RemoveOrgMemberParams{OrgID: org.ID, UserID: seed.UserID})
		pool := testutil.Pool(t)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", org.ID)
	})

	// Second create with same name should fail
	body2, _ := json.Marshal(map[string]string{
		"name":         orgName,
		"display_name": "Dup Org Again",
	})
	req2 := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader(body2))
	req2 = req2.WithContext(withUserID(req2.Context(), seed.UserID))
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusConflict {
		t.Errorf("duplicate create: expected 409, got %d", rr2.Code)
	}
}

func TestGetOrg_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	// Create an org first
	suffix := uuid.New().String()[:8]
	orgName := fmt.Sprintf("get-org-%s", suffix)
	body, _ := json.Marshal(map[string]string{
		"name":         orgName,
		"display_name": "Get Org " + suffix,
	})
	req := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader(body))
	req = req.WithContext(withUserID(req.Context(), seed.UserID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rr.Code)
	}
	var created db.Organization
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	t.Cleanup(func() {
		ctx := context.Background()
		_ = q.RemoveOrgMember(ctx, db.RemoveOrgMemberParams{OrgID: created.ID, UserID: seed.UserID})
		pool := testutil.Pool(t)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", created.ID)
	})

	// GET the org by slug
	req2 := httptest.NewRequest("GET", "/api/v1/orgs/"+orgName, nil)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("GetOrg: expected 200, got %d: %s", rr2.Code, rr2.Body.String())
	}
	var org db.Organization
	if err := json.Unmarshal(rr2.Body.Bytes(), &org); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if org.Name != orgName {
		t.Errorf("org name = %q, want %q", org.Name, orgName)
	}
}

func TestGetOrg_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	req := httptest.NewRequest("GET", "/api/v1/orgs/nonexistent-org-xyz-12345", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestListOrgs_NoAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	req := httptest.NewRequest("GET", "/api/v1/orgs", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", rr.Code)
	}
}

func TestListOrgs_ReturnsCreatedOrg(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	// Create an org
	suffix := uuid.New().String()[:8]
	orgName := fmt.Sprintf("list-org-%s", suffix)
	body, _ := json.Marshal(map[string]string{
		"name":         orgName,
		"display_name": "List Org",
	})
	req := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader(body))
	req = req.WithContext(withUserID(req.Context(), seed.UserID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rr.Code)
	}
	var created db.Organization
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	t.Cleanup(func() {
		ctx := context.Background()
		_ = q.RemoveOrgMember(ctx, db.RemoveOrgMemberParams{OrgID: created.ID, UserID: seed.UserID})
		pool := testutil.Pool(t)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", created.ID)
	})

	// List orgs for user
	req2 := httptest.NewRequest("GET", "/api/v1/orgs", nil)
	req2 = req2.WithContext(withUserID(req2.Context(), seed.UserID))
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("ListOrgs: expected 200, got %d: %s", rr2.Code, rr2.Body.String())
	}

	var orgList []db.Organization
	if err := json.Unmarshal(rr2.Body.Bytes(), &orgList); err != nil {
		t.Fatalf("decode: %v", err)
	}

	found := false
	for _, o := range orgList {
		if o.Name == orgName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created org %q not found in list", orgName)
	}
}

func TestListMembers_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	// Create an org (owner auto-added)
	suffix := uuid.New().String()[:8]
	orgName := fmt.Sprintf("members-org-%s", suffix)
	body, _ := json.Marshal(map[string]string{
		"name":         orgName,
		"display_name": "Members Org",
	})
	req := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader(body))
	req = req.WithContext(withUserID(req.Context(), seed.UserID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rr.Code)
	}
	var created db.Organization
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	t.Cleanup(func() {
		ctx := context.Background()
		_ = q.RemoveOrgMember(ctx, db.RemoveOrgMemberParams{OrgID: created.ID, UserID: seed.UserID})
		pool := testutil.Pool(t)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", created.ID)
	})

	// List members
	req2 := httptest.NewRequest("GET", "/api/v1/orgs/"+orgName+"/members", nil)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("ListMembers: expected 200, got %d: %s", rr2.Code, rr2.Body.String())
	}

	var members []db.ListOrgMembersRow
	if err := json.Unmarshal(rr2.Body.Bytes(), &members); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 member (owner), got %d", len(members))
	}
	if members[0].Role != "owner" {
		t.Errorf("member role = %q, want %q", members[0].Role, "owner")
	}
}

func TestRemoveMember_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	// Create an org
	suffix := uuid.New().String()[:8]
	orgName := fmt.Sprintf("remove-org-%s", suffix)
	body, _ := json.Marshal(map[string]string{
		"name":         orgName,
		"display_name": "Remove Org",
	})
	req := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader(body))
	req = req.WithContext(withUserID(req.Context(), seed.UserID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rr.Code)
	}
	var created db.Organization
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	t.Cleanup(func() {
		ctx := context.Background()
		pool := testutil.Pool(t)
		_, _ = pool.Exec(ctx, "DELETE FROM org_members WHERE org_id = $1", created.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", created.ID)
	})

	// Remove the owner
	req2 := httptest.NewRequest("DELETE", "/api/v1/orgs/"+orgName+"/members/"+seed.UserID.String(), nil)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusNoContent {
		t.Errorf("RemoveMember: expected 204, got %d: %s", rr2.Code, rr2.Body.String())
	}

	// Verify member list is now empty
	req3 := httptest.NewRequest("GET", "/api/v1/orgs/"+orgName+"/members", nil)
	rr3 := httptest.NewRecorder()
	router.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("ListMembers after remove: expected 200, got %d", rr3.Code)
	}
	var members []json.RawMessage
	_ = json.Unmarshal(rr3.Body.Bytes(), &members)
	if len(members) != 0 {
		t.Errorf("expected 0 members after removal, got %d", len(members))
	}
}

func TestRemoveMember_InvalidUserID(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	// Create an org
	suffix := uuid.New().String()[:8]
	orgName := fmt.Sprintf("inv-uid-org-%s", suffix)
	body, _ := json.Marshal(map[string]string{
		"name":         orgName,
		"display_name": "Inv UID Org",
	})
	req := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader(body))
	req = req.WithContext(withUserID(req.Context(), seed.UserID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rr.Code)
	}
	var created db.Organization
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	t.Cleanup(func() {
		ctx := context.Background()
		_ = q.RemoveOrgMember(ctx, db.RemoveOrgMemberParams{OrgID: created.ID, UserID: seed.UserID})
		pool := testutil.Pool(t)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", created.ID)
	})

	req2 := httptest.NewRequest("DELETE", "/api/v1/orgs/"+orgName+"/members/not-a-uuid", nil)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID, got %d", rr2.Code)
	}
}

func TestUpdateOrg_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	// Create an org
	suffix := uuid.New().String()[:8]
	orgName := fmt.Sprintf("update-org-%s", suffix)
	body, _ := json.Marshal(map[string]string{
		"name":         orgName,
		"display_name": "Before Update",
	})
	req := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader(body))
	req = req.WithContext(withUserID(req.Context(), seed.UserID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rr.Code)
	}
	var created db.Organization
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	t.Cleanup(func() {
		ctx := context.Background()
		_ = q.RemoveOrgMember(ctx, db.RemoveOrgMemberParams{OrgID: created.ID, UserID: seed.UserID})
		pool := testutil.Pool(t)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", created.ID)
	})

	// Update display name
	newDisplayName := "After Update " + suffix
	updateBody, _ := json.Marshal(map[string]string{
		"display_name": newDisplayName,
	})
	req2 := httptest.NewRequest("PATCH", "/api/v1/orgs/"+orgName, bytes.NewReader(updateBody))
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("UpdateOrg: expected 200, got %d: %s", rr2.Code, rr2.Body.String())
	}

	// Verify the update persisted
	req3 := httptest.NewRequest("GET", "/api/v1/orgs/"+orgName, nil)
	rr3 := httptest.NewRecorder()
	router.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("GetOrg after update: expected 200, got %d", rr3.Code)
	}
	var updated db.Organization
	_ = json.Unmarshal(rr3.Body.Bytes(), &updated)
	if updated.DisplayName != newDisplayName {
		t.Errorf("display_name = %q, want %q", updated.DisplayName, newDisplayName)
	}
}

func TestUpdateOrg_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	body, _ := json.Marshal(map[string]string{
		"display_name": "ghost",
	})
	req := httptest.NewRequest("PATCH", "/api/v1/orgs/nonexistent-org-xyz-99999", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateOrg_BadJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	// Create an org
	suffix := uuid.New().String()[:8]
	orgName := fmt.Sprintf("badjson-org-%s", suffix)
	body, _ := json.Marshal(map[string]string{
		"name":         orgName,
		"display_name": "Bad JSON Org",
	})
	req := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader(body))
	req = req.WithContext(withUserID(req.Context(), seed.UserID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rr.Code)
	}
	var created db.Organization
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	t.Cleanup(func() {
		ctx := context.Background()
		_ = q.RemoveOrgMember(ctx, db.RemoveOrgMemberParams{OrgID: created.ID, UserID: seed.UserID})
		pool := testutil.Pool(t)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", created.ID)
	})

	req2 := httptest.NewRequest("PATCH", "/api/v1/orgs/"+orgName, bytes.NewReader([]byte("not json")))
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad JSON, got %d", rr2.Code)
	}
}

func TestInviteMember_NotImplemented(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	// Create an org
	suffix := uuid.New().String()[:8]
	orgName := fmt.Sprintf("invite-org-%s", suffix)
	body, _ := json.Marshal(map[string]string{
		"name":         orgName,
		"display_name": "Invite Org",
	})
	req := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader(body))
	req = req.WithContext(withUserID(req.Context(), seed.UserID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rr.Code)
	}
	var created db.Organization
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	t.Cleanup(func() {
		ctx := context.Background()
		_ = q.RemoveOrgMember(ctx, db.RemoveOrgMemberParams{OrgID: created.ID, UserID: seed.UserID})
		pool := testutil.Pool(t)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", created.ID)
	})

	req2 := httptest.NewRequest("POST", "/api/v1/orgs/"+orgName+"/members/invite", nil)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusNotImplemented {
		t.Errorf("InviteMember: expected 501, got %d", rr2.Code)
	}
}

func TestConfigureSSO_NotImplemented(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := orgs.NewHandler(q)
	router := orgRouter(h)

	// Create an org
	suffix := uuid.New().String()[:8]
	orgName := fmt.Sprintf("sso-org-%s", suffix)
	body, _ := json.Marshal(map[string]string{
		"name":         orgName,
		"display_name": "SSO Org",
	})
	req := httptest.NewRequest("POST", "/api/v1/orgs", bytes.NewReader(body))
	req = req.WithContext(withUserID(req.Context(), seed.UserID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rr.Code)
	}
	var created db.Organization
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	t.Cleanup(func() {
		ctx := context.Background()
		_ = q.RemoveOrgMember(ctx, db.RemoveOrgMemberParams{OrgID: created.ID, UserID: seed.UserID})
		pool := testutil.Pool(t)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", created.ID)
	})

	req2 := httptest.NewRequest("POST", "/api/v1/orgs/"+orgName+"/sso", nil)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusNotImplemented {
		t.Errorf("ConfigureSSO: expected 501, got %d", rr2.Code)
	}
}
