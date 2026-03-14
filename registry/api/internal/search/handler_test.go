package search_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rithul/hivemind/registry/api/internal/search"
	"github.com/rithul/hivemind/registry/api/internal/testutil"
)

func TestSearch_EmptyQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := search.NewHandler(q)

	req := httptest.NewRequest("GET", "/api/v1/search?q=", nil)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Search returned %d", rr.Code)
	}
	var resp struct {
		Results []json.RawMessage `json:"results"`
		Page    int               `json:"page"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results for empty query, got %d", len(resp.Results))
	}
	if resp.Page != 1 {
		t.Errorf("page = %d, want 1", resp.Page)
	}
}

func TestSearch_WithQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	seed := testutil.SetupTestData(t, q)
	h := search.NewHandler(q)

	// Search for the seeded package by part of its name
	// The package name is "test-pkg-XXXX" and search_vector includes name + description
	req := httptest.NewRequest("GET", "/api/v1/search?q=test+package", nil)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Search returned %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Results []json.RawMessage `json:"results"`
		Page    int               `json:"page"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Should find at least our seeded package (description contains "A test package")
	found := false
	for _, r := range resp.Results {
		if strings.Contains(string(r), seed.PackageName) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find %q in search results", seed.PackageName)
	}
}

func TestSearch_ReturnsEmptyArrayNotNull(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := search.NewHandler(q)

	req := httptest.NewRequest("GET", "/api/v1/search?q=zzz_nonexistent_pkg_xyz", nil)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Search returned %d", rr.Code)
	}
	body := rr.Body.String()
	if strings.Contains(body, `"results":null`) {
		t.Error("results should be [] not null when no matches")
	}
}

func TestSearch_Pagination(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := search.NewHandler(q)

	req := httptest.NewRequest("GET", "/api/v1/search?q=test&page=2", nil)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Search returned %d", rr.Code)
	}
	var resp struct {
		Page int `json:"page"`
	}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Page != 2 {
		t.Errorf("page = %d, want 2", resp.Page)
	}
}

func TestSearch_SQLInjectionSafe(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := search.NewHandler(q)

	req := httptest.NewRequest("GET", "/api/v1/search?q='+DROP+TABLE+packages%3B--", nil)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Search returned %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["results"]; !ok {
		t.Error("response missing 'results' key")
	}
}

func TestSearch_SpecialCharacters(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := search.NewHandler(q)

	req := httptest.NewRequest("GET", "/api/v1/search?q=%25%26%3C%3E%22", nil)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Search returned %d: %s", rr.Code, rr.Body.String())
	}
}

func TestSearch_NegativePage(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := search.NewHandler(q)

	req := httptest.NewRequest("GET", "/api/v1/search?q=test&page=-1", nil)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Search returned %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Page int `json:"page"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Page != 1 {
		t.Errorf("page = %d, want 1 (should default negative page to 1)", resp.Page)
	}
}

func TestSearch_LargePageNumber(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := search.NewHandler(q)

	req := httptest.NewRequest("GET", "/api/v1/search?q=test&page=999999", nil)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Search returned %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Results []json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results for huge page offset, got %d", len(resp.Results))
	}
}

func TestSearch_Suggestions(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	q := testutil.Queries(t)
	h := search.NewHandler(q)

	// Search for "tset" (typo for "test") — should return 200 regardless of result count
	req := httptest.NewRequest("GET", "/api/v1/search?q=tset", nil)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Search returned %d: %s", rr.Code, rr.Body.String())
	}
}
