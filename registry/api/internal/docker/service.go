package docker

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/db"
)

// Service manages Docker image metadata (ECR sync, tag listing).
type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service {
	return &Service{q: q}
}

// RegisterImage records or updates a Docker image for a package.
func (s *Service) RegisterImage(ctx context.Context, packageID uuid.UUID, tag, digest, ecrURI string, platform []string, sizeBytes *int64) error {
	var sz pgtype.Int8
	if sizeBytes != nil {
		sz.Int64 = *sizeBytes
		sz.Valid = true
	}
	_, err := s.q.CreateDockerImage(ctx, db.CreateDockerImageParams{
		PackageID: packageID,
		Tag:       tag,
		Digest:    digest,
		EcrUri:    ecrURI,
		Platform:  platform,
		SizeBytes:  sz,
	})
	return err
}
