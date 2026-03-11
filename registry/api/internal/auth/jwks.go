package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

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
func (v *JWKSVerifier) Verify(tokenString string) (*JWKSClaims, error) {
	fmt.Printf("Verifying token against JWKS. Token len: %d\n", len(tokenString))
	token, err := jwt.ParseWithClaims(tokenString, &JWKSClaims{}, v.jwks.Keyfunc)
	if err != nil {
		fmt.Printf("JWT Parse Error: %v\n", err)
		return nil, err
	}
	claims, ok := token.Claims.(*JWKSClaims)
	if !ok || !token.Valid {
		fmt.Printf("Invalid Token Claims or !token.Valid\n")
		return nil, fmt.Errorf("invalid token claims")
	}
	// Optional: check expiry is already validated by jwt.ParseWithClaims
	_ = time.Now()
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
