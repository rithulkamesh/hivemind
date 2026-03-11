package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type contextKey string

const (
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyUsername contextKey = "username"
	ContextKeyScopes   contextKey = "scopes"
	ContextKeyAPIKeyID contextKey = "api_key_id"
)

// Config for auth middleware.
type AuthConfig interface {
	GetJWTSecret() string
}

// RequireAuth extracts JWT or API key and sets user identity on context.
func RequireAuth(cfg AuthConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString := authHeader[7:]
				claims, err := VerifyToken(cfg.GetJWTSecret(), tokenString)
				if err != nil {
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}
				uid, _ := uuid.Parse(claims.UserID)
				ctx := context.WithValue(r.Context(), ContextKeyUserID, uid)
				ctx = context.WithValue(ctx, ContextKeyUsername, claims.Username)
				ctx = context.WithValue(ctx, ContextKeyScopes, claims.Scopes)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			// API key auth would be validated by a separate layer that looks up key_hash in DB
			// and sets user_id + scopes on context. For now we only support Bearer JWT.
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		})
	}
}

// RequireAdmin ensures the request has admin scope or user is admin (e.g. check scope or role).
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scopes, _ := r.Context().Value(ContextKeyScopes).([]string)
		for _, s := range scopes {
			if s == "admin" {
				next.ServeHTTP(w, r)
				return
			}
		}
		// TODO: check user role in DB if no scopes (session auth)
		http.Error(w, "forbidden", http.StatusForbidden)
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
