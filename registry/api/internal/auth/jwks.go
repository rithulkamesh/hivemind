package auth

import (
	"context"
	"fmt"
	"sync"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// JWKSClaims holds Better Auth JWT claims (sub = user ID, email, session).
type JWKSClaims struct {
	Subject string   `json:"sub"`
	Email   string   `json:"email"`
	Name    string   `json:"name"`
	Scopes  []string `json:"scopes"`
	jwt.RegisteredClaims
}

// JWKSVerifier verifies JWTs using a remote JWKS endpoint (e.g. Better Auth).
type JWKSVerifier struct {
	jwks keyfunc.Keyfunc
}

// NewJWKSVerifier creates a verifier that fetches and caches JWKS from jwksURL, refreshing periodically.
func NewJWKSVerifier(ctx context.Context, jwksURL string) (*JWKSVerifier, error) {
	if jwksURL == "" {
		return nil, fmt.Errorf("JWKS URL is required")
	}
	// NewDefaultCtx starts a background refresh; refresh interval is handled by jwkset (e.g. 15 min).
	k, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("create JWKS keyfunc: %w", err)
	}
	return &JWKSVerifier{jwks: k}, nil
}

// Verify parses and verifies the token string using JWKS and returns claims.
// Algorithm is pinned to RS256 and ES256 to prevent algorithm confusion attacks.
func (v *JWKSVerifier) Verify(tokenString string) (*JWKSClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWKSClaims{}, v.jwks.Keyfunc,
		jwt.WithValidMethods([]string{"RS256", "ES256"}),
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWKSClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// Ensure we only create one JWKS verifier per URL (e.g. in main).
var (
	verifierInstance *JWKSVerifier
	verifierMu       sync.Mutex
)

// SetGlobalJWKSVerifier sets the global verifier used by RequireAuth when JWKS is configured.
func SetGlobalJWKSVerifier(v *JWKSVerifier) {
	verifierMu.Lock()
	defer verifierMu.Unlock()
	verifierInstance = v
}

// GetGlobalJWKSVerifier returns the global JWKS verifier if set.
func GetGlobalJWKSVerifier() *JWKSVerifier {
	verifierMu.Lock()
	defer verifierMu.Unlock()
	return verifierInstance
}
