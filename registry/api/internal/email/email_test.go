package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockSender records calls to Send for verification.
type mockSender struct {
	calls []sendCall
	err   error
}

type sendCall struct {
	to, subject, bodyText, bodyHTML string
}

func (m *mockSender) Send(_ context.Context, to, subject, bodyText, bodyHTML string) error {
	m.calls = append(m.calls, sendCall{to, subject, bodyText, bodyHTML})
	return m.err
}

// ── RequireInternalSecret ────────────────────────────────────────────────

func TestRequireInternalSecret_ValidSecret(t *testing.T) {
	h := &InternalHandler{Secret: "my-secret-32-chars-long-for-test"}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := h.RequireInternalSecret(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Internal-Secret", "my-secret-32-chars-long-for-test")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 with valid secret, got %d", rr.Code)
	}
}

func TestRequireInternalSecret_InvalidSecret(t *testing.T) {
	h := &InternalHandler{Secret: "correct-secret"}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	handler := h.RequireInternalSecret(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Internal-Secret", "wrong-secret")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with wrong secret, got %d", rr.Code)
	}
}

func TestRequireInternalSecret_MissingSecret(t *testing.T) {
	h := &InternalHandler{Secret: "correct-secret"}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	handler := h.RequireInternalSecret(inner)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with missing secret, got %d", rr.Code)
	}
}

func TestRequireInternalSecret_EmptyConfigSecret(t *testing.T) {
	h := &InternalHandler{Secret: ""} // Server misconfiguration
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	handler := h.RequireInternalSecret(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Internal-Secret", "anything")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 when server secret is empty, got %d", rr.Code)
	}
}

// ── ServeVerify ──────────────────────────────────────────────────────────

func TestServeVerify_Success(t *testing.T) {
	sender := &mockSender{}
	h := &InternalHandler{Secret: "test", Send: sender}

	body, _ := json.Marshal(VerifyRequest{
		Email: "user@test.com",
		URL:   "https://example.com/verify?token=abc",
	})
	req := httptest.NewRequest("POST", "/internal/email/verify", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeVerify(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
	if len(sender.calls) != 1 {
		t.Fatalf("expected 1 send call, got %d", len(sender.calls))
	}
	if sender.calls[0].to != "user@test.com" {
		t.Errorf("to = %q, want %q", sender.calls[0].to, "user@test.com")
	}
	if sender.calls[0].subject == "" {
		t.Error("subject should not be empty")
	}
}

func TestServeVerify_MissingFields(t *testing.T) {
	h := &InternalHandler{Secret: "test", Send: &mockSender{}}

	tests := []struct {
		name string
		body VerifyRequest
	}{
		{"missing email", VerifyRequest{URL: "https://example.com/verify"}},
		{"missing url", VerifyRequest{Email: "user@test.com"}},
		{"both empty", VerifyRequest{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/internal/email/verify", bytes.NewReader(body))
			rr := httptest.NewRecorder()
			h.ServeVerify(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}
		})
	}
}

func TestServeVerify_MethodNotAllowed(t *testing.T) {
	h := &InternalHandler{Secret: "test"}

	req := httptest.NewRequest("GET", "/internal/email/verify", nil)
	rr := httptest.NewRecorder()
	h.ServeVerify(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestServeVerify_NilSender_Returns204(t *testing.T) {
	h := &InternalHandler{Secret: "test", Send: nil}

	body, _ := json.Marshal(VerifyRequest{Email: "user@test.com", URL: "https://example.com"})
	req := httptest.NewRequest("POST", "/internal/email/verify", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeVerify(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 with nil sender, got %d", rr.Code)
	}
}

func TestServeVerify_SenderError(t *testing.T) {
	sender := &mockSender{err: fmt.Errorf("SMTP connection refused")}
	h := &InternalHandler{Secret: "test", Send: sender}

	body, _ := json.Marshal(VerifyRequest{Email: "user@test.com", URL: "https://example.com"})
	req := httptest.NewRequest("POST", "/internal/email/verify", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeVerify(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on sender error, got %d", rr.Code)
	}
}

// ── ServeSend ────────────────────────────────────────────────────────────

func TestServeSend_Success(t *testing.T) {
	sender := &mockSender{}
	h := &InternalHandler{Secret: "test", Send: sender}

	body, _ := json.Marshal(SendRequest{
		To:       "user@test.com",
		Subject:  "Test Subject",
		BodyText: "Hello!",
	})
	req := httptest.NewRequest("POST", "/internal/email/send", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeSend(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
	if len(sender.calls) != 1 {
		t.Fatalf("expected 1 send call, got %d", len(sender.calls))
	}
	if sender.calls[0].to != "user@test.com" {
		t.Errorf("to = %q, want %q", sender.calls[0].to, "user@test.com")
	}
	if sender.calls[0].subject != "Test Subject" {
		t.Errorf("subject = %q, want %q", sender.calls[0].subject, "Test Subject")
	}
}

func TestServeSend_WithHTML(t *testing.T) {
	sender := &mockSender{}
	h := &InternalHandler{Secret: "test", Send: sender}

	body, _ := json.Marshal(SendRequest{
		To:       "user@test.com",
		Subject:  "HTML Email",
		BodyText: "plain text",
		BodyHTML: "<h1>Hello</h1>",
	})
	req := httptest.NewRequest("POST", "/internal/email/send", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeSend(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
	if sender.calls[0].bodyHTML != "<h1>Hello</h1>" {
		t.Errorf("bodyHTML = %q, want %q", sender.calls[0].bodyHTML, "<h1>Hello</h1>")
	}
}

func TestServeSend_MissingFields(t *testing.T) {
	h := &InternalHandler{Secret: "test", Send: &mockSender{}}

	tests := []struct {
		name string
		body SendRequest
	}{
		{"missing to", SendRequest{Subject: "Test"}},
		{"missing subject", SendRequest{To: "user@test.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/internal/email/send", bytes.NewReader(body))
			rr := httptest.NewRecorder()
			h.ServeSend(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}
		})
	}
}

func TestServeSend_MethodNotAllowed(t *testing.T) {
	h := &InternalHandler{Secret: "test"}

	req := httptest.NewRequest("GET", "/internal/email/send", nil)
	rr := httptest.NewRecorder()
	h.ServeSend(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestServeSend_NilSender(t *testing.T) {
	h := &InternalHandler{Secret: "test", Send: nil}

	body, _ := json.Marshal(SendRequest{To: "user@test.com", Subject: "Test"})
	req := httptest.NewRequest("POST", "/internal/email/send", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeSend(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 with nil sender, got %d", rr.Code)
	}
}

func TestServeSend_SenderError(t *testing.T) {
	sender := &mockSender{err: fmt.Errorf("send failed")}
	h := &InternalHandler{Secret: "test", Send: sender}

	body, _ := json.Marshal(SendRequest{To: "user@test.com", Subject: "Test", BodyText: "body"})
	req := httptest.NewRequest("POST", "/internal/email/send", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeSend(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on sender error, got %d", rr.Code)
	}
}

func TestServeSend_InvalidJSON(t *testing.T) {
	h := &InternalHandler{Secret: "test", Send: &mockSender{}}

	req := httptest.NewRequest("POST", "/internal/email/send", bytes.NewReader([]byte("not json")))
	rr := httptest.NewRecorder()
	h.ServeSend(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rr.Code)
	}
}

func TestServeVerify_InvalidJSON(t *testing.T) {
	h := &InternalHandler{Secret: "test", Send: &mockSender{}}

	req := httptest.NewRequest("POST", "/internal/email/verify", bytes.NewReader([]byte("{bad")))
	rr := httptest.NewRecorder()
	h.ServeVerify(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rr.Code)
	}
}
