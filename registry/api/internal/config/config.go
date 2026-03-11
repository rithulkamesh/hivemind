package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// AuthConfig is satisfied by Config for JWT secret.
type AuthConfig interface {
	GetJWTSecret() string
}

// Config holds all configuration loaded from environment variables.
type Config struct {
	// Database
	DatabaseURL string

	// S3 / storage
	S3Bucket            string
	S3Region            string
	S3CloudFrontDomain  string

	// JWT
	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	// OAuth
	GitHubClientID     string
	GitHubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string

	// SES
	SESRegion       string
	SESFromAddress  string
	SESReplyTo      string

	// App
	BaseURL string
	Port    string

	// ECR (Docker images)
	ECRRegistry string

	// Admin bootstrap
	AdminSecret string

	// Rate limit
	RateLimitRPS int

	// Upload / verification
	MaxUploadSizeMB   int
	VerificationWorkers int
}

// Load reads configuration from environment. No config file in production.
func Load() (*Config, error) {
	c := &Config{
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		S3Bucket:            getEnv("S3_BUCKET", "hivemind-registry-packages"),
		S3Region:            getEnv("S3_REGION", "us-east-1"),
		S3CloudFrontDomain:  getEnv("S3_CLOUDFRONT_DOMAIN", "packages.hivemind.rithul.dev"),
		JWTSecret:           getEnv("JWT_SECRET", ""),
		JWTAccessTTL:        parseDuration(getEnv("JWT_ACCESS_TTL", "24h"), 24*time.Hour),
		JWTRefreshTTL:       parseDuration(getEnv("JWT_REFRESH_TTL", "720h"), 720*time.Hour),
		GitHubClientID:      getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret:  getEnv("GITHUB_CLIENT_SECRET", ""),
		GoogleClientID:      getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:  getEnv("GOOGLE_CLIENT_SECRET", ""),
		SESRegion:           getEnv("SES_REGION", "us-east-1"),
		SESFromAddress:      getEnv("SES_FROM_ADDRESS", "noreply@hivemind.rithul.dev"),
		SESReplyTo:          getEnv("SES_REPLY_TO", "support@hivemind.rithul.dev"),
		BaseURL:             getEnv("BASE_URL", "https://registry.hivemind.rithul.dev"),
		Port:                getEnv("PORT", "8080"),
		ECRRegistry:         getEnv("ECR_REGISTRY", ""),
		AdminSecret:         getEnv("ADMIN_SECRET", ""),
		RateLimitRPS:        getEnvInt("RATE_LIMIT_RPS", 10),
		MaxUploadSizeMB:     getEnvInt("MAX_UPLOAD_SIZE_MB", 100),
		VerificationWorkers: getEnvInt("VERIFICATION_WORKERS", 4),
	}

	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return c, nil
}

// GetJWTSecret returns the JWT signing secret for auth middleware.
func (c *Config) GetJWTSecret() string {
	return c.JWTSecret
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func parseDuration(s string, def time.Duration) time.Duration {
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	return d
}
