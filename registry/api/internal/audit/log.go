package audit

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/db"
)

// Log records an audit event (append-only).
func Log(ctx context.Context, q *db.Queries, actorUserID, actorAPIKeyID *uuid.UUID, action, resourceType, resourceID, ipHash string, metadata interface{}) (int64, error) {
	var userID, keyID pgtype.UUID
	if actorUserID != nil {
		userID.Bytes = *actorUserID
		userID.Valid = true
	}
	if actorAPIKeyID != nil {
		keyID.Bytes = *actorAPIKeyID
		keyID.Valid = true
	}
	var meta []byte
	if metadata != nil {
		meta, _ = json.Marshal(metadata)
	}
	return q.InsertAuditLog(ctx, db.InsertAuditLogParams{
		ActorUserID:   userID,
		ActorApiKeyID: keyID,
		Action:        action,
		ResourceType:  pgtype.Text{String: resourceType, Valid: resourceType != ""},
		ResourceID:    pgtype.Text{String: resourceID, Valid: resourceID != ""},
		Metadata:      meta,
		IpHash:        pgtype.Text{String: ipHash, Valid: ipHash != ""},
	})
}
