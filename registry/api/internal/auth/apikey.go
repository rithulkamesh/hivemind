package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const KeyPrefix = "hm_"

// GenerateAPIKey returns a new raw key (to show once) and its hash and prefix for storage.
func GenerateAPIKey() (raw, hash, prefix string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", err
	}
	raw = KeyPrefix + hex.EncodeToString(b)
	hash = HashKey(raw)
	prefix = raw
	if len(prefix) > len(KeyPrefix)+8 {
		prefix = prefix[:len(KeyPrefix)+8] + "..."
	}
	return raw, hash, prefix, nil
}

// HashKey returns sha256 hex of the raw key.
func HashKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// VerifyAPIKey returns true if rawKey hashes to storedHash.
func VerifyAPIKey(rawKey, storedHash string) bool {
	return HashKey(rawKey) == storedHash
}

// ExtractBearerKey returns the key from "Bearer hm_xxx" or empty.
func ExtractBearerKey(authHeader string) (string, error) {
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", fmt.Errorf("missing or invalid authorization header")
	}
	return authHeader[7:], nil
}
