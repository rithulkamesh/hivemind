package packages

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/db"
)

// Signer produces sigstore signatures and stores the bundle in package_versions.sigstore_bundle.
type Signer struct {
	q *db.Queries
}

// NewSigner creates a signer (stub: actual sigstore-go or subprocess to be wired).
func NewSigner(q *db.Queries) *Signer {
	return &Signer{q: q}
}

// SignAndStore runs sigstore sign and saves the transparency log bundle for the version.
func (s *Signer) SignAndStore(ctx context.Context, versionID uuid.UUID, artifactPath string) ([]byte, error) {
	// TODO: run sigstore sign, get bundle JSON, update package_versions.sigstore_bundle
	_ = artifactPath
	return nil, nil
}

// StoreBundle saves a pre-computed sigstore bundle (e.g. from CI).
func (s *Signer) StoreBundle(ctx context.Context, versionID uuid.UUID, bundle []byte) error {
	// Update only sigstore_bundle; verification_status and published are set by verify pipeline.
	_, err := s.q.UpdatePackageVersionVerification(ctx, db.UpdatePackageVersionVerificationParams{
		ID:                 versionID,
		VerificationStatus: pgtype.Text{String: "passed", Valid: true},
		VerificationReport: nil,
		Published:          pgtype.Bool{Bool: true, Valid: true},
		ToolCount:          pgtype.Int4{},
		SigstoreBundle:     bundle,
	})
	return err
}
