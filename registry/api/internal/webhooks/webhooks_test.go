package webhooks

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitHub_AcceptsPost(t *testing.T) {
	h := NewHandler()

	req := httptest.NewRequest("POST", "/webhooks/github", strings.NewReader(`{"action":"published"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.GitHub(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rr.Code)
	}
}

func TestGitHub_RejectsGet(t *testing.T) {
	h := NewHandler()

	req := httptest.NewRequest("GET", "/webhooks/github", nil)
	rr := httptest.NewRecorder()
	h.GitHub(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestGitHub_RejectsPut(t *testing.T) {
	h := NewHandler()

	req := httptest.NewRequest("PUT", "/webhooks/github", nil)
	rr := httptest.NewRecorder()
	h.GitHub(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestNewHandler(t *testing.T) {
	h := NewHandler()
	if h == nil {
		t.Fatal("NewHandler() returned nil")
	}
}
