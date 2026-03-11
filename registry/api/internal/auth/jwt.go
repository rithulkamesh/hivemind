package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Scopes   []string `json:"scopes,omitempty"`
}

// IssueAccessToken creates a short-lived access token (e.g. 24h).
func IssueAccessToken(secret string, userID, username string, ttl time.Duration) (string, error) {
	return issueToken(secret, userID, username, nil, ttl, "access")
}

// IssueRefreshToken creates a long-lived refresh token (e.g. 30d).
func IssueRefreshToken(secret string, userID, username string, ttl time.Duration) (string, error) {
	return issueToken(secret, userID, username, nil, ttl, "refresh")
}

func issueToken(secret string, userID, username string, scopes []string, ttl time.Duration, subject string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ID:        uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		UserID:   userID,
		Username: username,
		Scopes:   scopes,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// VerifyToken parses and validates a JWT, returns claims or error.
func VerifyToken(secret, tokenString string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := t.Claims.(*Claims); ok && t.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
