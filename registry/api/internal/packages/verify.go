package packages

import (
	"context"

	"github.com/google/uuid"
	"github.com/rithul/hivemind/registry/api/internal/db"
)

// Verifier runs the verification pipeline (safety, pip-audit, bandit, etc.) in the background.
type Verifier struct {
	q      *db.Queries
	workers int
}

// NewVerifier creates a verifier that enqueues jobs (stub: no workers yet).
func NewVerifier(q *db.Queries, workers int) *Verifier {
	return &Verifier{q: q, workers: workers}
}

// RunVerification runs the pipeline for a package version (stub).
func (v *Verifier) RunVerification(ctx context.Context, versionID uuid.UUID) error {
	// TODO: extract artifact, run safety/pip-audit/bandit, update verification_status and verification_report
	return nil
}
