package email

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

const internalSecretHeader = "X-Internal-Secret"

// Sender sends transactional email (SES or SMTP e.g. Mailhog).
type Sender interface {
	Send(ctx context.Context, to, subject, bodyText, bodyHTML string) error
}

// InternalHandler handles internal email endpoints (called by Better Auth service).
type InternalHandler struct {
	Secret string
	Send   Sender // nil means no-op (no email sent)
}

// VerifyRequest is the body for POST /internal/email/verify.
type VerifyRequest struct {
	Email string `json:"email"`
	URL   string `json:"url"`
}

// SendRequest is the body for POST /internal/email/send.
type SendRequest struct {
	To       string `json:"to"`
	Subject  string `json:"subject"`
	BodyText string `json:"body_text"`
	BodyHTML string `json:"body_html,omitempty"`
}

// RequireInternalSecret returns a middleware that checks X-Internal-Secret header.
func (h *InternalHandler) RequireInternalSecret(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := strings.TrimSpace(r.Header.Get(internalSecretHeader))
		if h.Secret == "" || secret != h.Secret {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ServeVerify sends a verification email (link) via SES. Called by Better Auth.
func (h *InternalHandler) ServeVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}
	if req.Email == "" || req.URL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "email and url required"})
		return
	}
	if h.Send == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	subject := "Verify your email – Hivemind Registry"
	body := "Verify your email by clicking the link below (expires in 24 hours):\n\n" + req.URL + "\n\n— Hivemind Registry"
	if err := h.Send.Send(r.Context(), req.Email, subject, body, ""); err != nil {
		http.Error(w, "failed to send email", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ServeSend sends a generic email via SES.
func (h *InternalHandler) ServeSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}
	if req.To == "" || req.Subject == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "to and subject required"})
		return
	}
	if h.Send == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := h.Send.Send(r.Context(), req.To, req.Subject, req.BodyText, req.BodyHTML); err != nil {
		http.Error(w, "failed to send email", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
