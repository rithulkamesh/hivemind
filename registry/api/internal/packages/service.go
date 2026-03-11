package packages

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/db"
)

// Service holds business logic for package upload, verify, publish, yank.
type Service struct {
	q      *db.Queries
	store  Storage
	verify *Verifier
}

// NewService creates a package service.
func NewService(q *db.Queries, store Storage, verify *Verifier) *Service {
	return &Service{q: q, store: store, verify: verify}
}

// CreateVersion creates a package version record (pending verification).
func (s *Service) CreateVersion(ctx context.Context, packageID uuid.UUID, version, requiresPython, requiresHivemind string, uploadedBy uuid.UUID) (*db.PackageVersion, error) {
	var up pgtype.UUID
	up.Bytes = uploadedBy
	up.Valid = true
	pv, err := s.q.CreatePackageVersion(ctx, db.CreatePackageVersionParams{
		PackageID:          packageID,
		Version:             version,
		RequiresPython:      pgtype.Text{String: requiresPython, Valid: requiresPython != ""},
		RequiresHivemind:    pgtype.Text{String: requiresHivemind, Valid: requiresHivemind != ""},
		UploadedBy:          up,
		VerificationStatus:  pgtype.Text{String: "pending", Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &pv, nil
}

// AddFile records a package file and enqueues verification.
func (s *Service) AddFile(ctx context.Context, versionID uuid.UUID, filename, filetype, pythonVersion, abi, platform string, sizeBytes int64, sha256, md5, s3Key string) (*db.PackageFile, error) {
	f, err := s.q.CreatePackageFile(ctx, db.CreatePackageFileParams{
		VersionID:     versionID,
		Filename:      filename,
		Filetype:      filetype,
		PythonVersion: pgtype.Text{String: pythonVersion, Valid: pythonVersion != ""},
		Abi:           pgtype.Text{String: abi, Valid: abi != ""},
		Platform:      pgtype.Text{String: platform, Valid: platform != ""},
		SizeBytes:     sizeBytes,
		Sha256:        sha256,
		Md5:           md5,
		S3Key:         s3Key,
	})
	if err != nil {
		return nil, err
	}
	go s.verify.RunVerification(context.Background(), versionID)
	return &f, nil
}
