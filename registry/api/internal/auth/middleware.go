package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/db"
)

type contextKey string

const (
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyUsername contextKey = "username"
	ContextKeyEmail    contextKey = "email"
	ContextKeyScopes   contextKey = "scopes"
	ContextKeyAPIKeyID contextKey = "api_key_id"
)

// AuthConfig provides optional legacy JWT secret (used only when JWKS is not set).
type AuthConfig interface {
	GetJWTSecret() string
}

// RequireAuth extracts Bearer JWT (verified via JWKS or legacy secret) or X-API-Key and sets identity on context.
// On failure responds with 401 JSON {"error": "unauthorized"}.
func RequireAuth(cfg AuthConfig, q *db.Queries) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1) Bearer token (Better Auth JWT via JWKS or legacy JWT)
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString := strings.TrimSpace(authHeader[7:])
				if tokenString == "" {
					writeAuthError(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				if jwks := GetGlobalJWKSVerifier(); jwks != nil {
					claims, err := jwks.Verify(tokenString)
					if err != nil {
						writeAuthError(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					uid, err := uuid.Parse(claims.Subject) // sub = Better Auth user ID
					if err != nil {
						writeAuthError(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					username := claims.Email
					if username == "" {
						username = claims.Subject
					}
					ctx := context.WithValue(r.Context(), ContextKeyUserID, uid)
					ctx = context.WithValue(ctx, ContextKeyUsername, username)
					ctx = context.WithValue(ctx, ContextKeyEmail, claims.Email)
					ctx = context.WithValue(ctx, ContextKeyScopes, claims.Scopes)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				// Legacy JWT (custom signing)
				claims, err := VerifyToken(cfg.GetJWTSecret(), tokenString)
				if err != nil {
					writeAuthError(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				uid, _ := uuid.Parse(claims.UserID)
				ctx := context.WithValue(r.Context(), ContextKeyUserID, uid)
				ctx = context.WithValue(ctx, ContextKeyUsername, claims.Username)
				ctx = context.WithValue(ctx, ContextKeyScopes, claims.Scopes)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 2) API key via X-API-Key header OR Basic Auth (Twine)
			rawKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
			if rawKey == "" {
				if _, pass, ok := r.BasicAuth(); ok {
					rawKey = strings.TrimSpace(pass)
				}
			}

			if rawKey != "" {
				// Early prefix check: reject keys that don't start with "hm_" before
				// wasting a hash+DB round-trip on garbage input.
				if !strings.HasPrefix(rawKey, KeyPrefix) {
					writeAuthError(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				hash := HashKey(rawKey)
				key, err := q.GetAPIKeyByHash(r.Context(), hash)
				if err != nil {
					writeAuthError(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				// C5: Check API key expiry (DB query checks revoked but NOT expires_at).
				if key.ExpiresAt.Valid && time.Now().After(key.ExpiresAt.Time) {
					writeAuthError(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				ctx := context.WithValue(r.Context(), ContextKeyUserID, key.UserID)
				ctx = context.WithValue(ctx, ContextKeyUsername, "")
				ctx = context.WithValue(ctx, ContextKeyScopes, key.Scopes)
				ctx = context.WithValue(ctx, ContextKeyAPIKeyID, key.ID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			writeAuthError(w, "unauthorized", http.StatusUnauthorized)
		})
	}
}

func writeAuthError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// RequireScope ensures the request has the given scope (API key scopes or admin for JWT session).
// JWT sessions (no API key ID in context) are treated as having all scopes.
// On failure responds with 403 JSON {"error": "insufficient_scope"}.
func RequireScope(scope string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// JWT sessions (Better Auth or legacy) bypass scope checks — they have full access.
			if _, hasKey := r.Context().Value(ContextKeyAPIKeyID).(uuid.UUID); !hasKey {
				next.ServeHTTP(w, r)
				return
			}
			scopes, _ := r.Context().Value(ContextKeyScopes).([]string)
			for _, s := range scopes {
				if s == scope {
					next.ServeHTTP(w, r)
					return
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "insufficient_scope"})
		})
	}
}

// RequireAdmin ensures the request has admin scope (or future: user role in DB).
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scopes, _ := r.Context().Value(ContextKeyScopes).([]string)
		for _, s := range scopes {
			if s == "admin" {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "insufficient_scope"})
	})
}

// GetUserID returns the authenticated user ID from context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(ContextKeyUserID)
	if v == nil {
		return uuid.Nil, false
	}
	uid, ok := v.(uuid.UUID)
	return uid, ok
}

// GetUserIDPgType returns pgtype.UUID for DB calls.
func GetUserIDPgType(ctx context.Context) (pgtype.UUID, bool) {
	uid, ok := GetUserID(ctx)
	if !ok {
		return pgtype.UUID{}, false
	}
	var u pgtype.UUID
	u.Bytes = uid
	u.Valid = true
	return u, true
}

// GetEmail returns the email from context (set from Better Auth JWT claims).
func GetEmail(ctx context.Context) string {
	v := ctx.Value(ContextKeyEmail)
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
