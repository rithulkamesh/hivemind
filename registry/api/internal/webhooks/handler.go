package webhooks

import (
	"net/http"
)

// Handler handles webhook callbacks (e.g. GitHub Actions publish events).
type Handler struct{}

// NewHandler creates a webhook handler.
func NewHandler() *Handler {
	return &Handler{}
}

// GitHub handles POST /webhooks/github (e.g. workflow_dispatch or release publish).
func (h *Handler) GitHub(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// TODO: verify GitHub signature, parse event, trigger package publish or Docker image sync
	w.WriteHeader(http.StatusAccepted)
}
