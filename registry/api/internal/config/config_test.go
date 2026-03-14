package config

import (
	"strings"
	"testing"
	"time"
)

func TestConfig_ValidateProduction_RejectsHTTPJWKS(t *testing.T) {
	cfg := Config{
		Env:            "production",
		JWKSURL:        "http://insecure.example.com/jwks",
		InternalSecret: strings.Repeat("x", 32),
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for HTTP JWKS URL in production")
	}
	if !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("error should mention HTTPS, got: %v", err)
	}
}

func TestConfig_ValidateProduction_AcceptsHTTPS(t *testing.T) {
	cfg := Config{
		Env:            "production",
		JWKSURL:        "https://example.com/jwks",
		InternalSecret: strings.Repeat("x", 32),
	}
	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestConfig_ValidateDevelopment_AllowsHTTP(t *testing.T) {
	cfg := Config{
		Env:     "development",
		JWKSURL: "http://localhost:3000/auth/jwks",
	}
	err := cfg.Validate()
	if err != nil {
		t.Errorf("dev mode should allow HTTP JWKS, got: %v", err)
	}
}

func TestConfig_ValidateProduction_RejectsShortSecret(t *testing.T) {
	cfg := Config{
		Env:            "production",
		JWKSURL:        "https://example.com/jwks",
		InternalSecret: "short",
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for short INTERNAL_SECRET in production")
	}
	if !strings.Contains(err.Error(), "32") {
		t.Errorf("error should mention 32 characters, got: %v", err)
	}
}

func TestConfig_ValidateProduction_EmptyJWKSAllowed(t *testing.T) {
	// Production with no JWKS_URL is fine (uses legacy JWT_SECRET)
	cfg := Config{
		Env:            "production",
		InternalSecret: strings.Repeat("x", 32),
	}
	err := cfg.Validate()
	if err != nil {
		t.Errorf("empty JWKS_URL should not cause validation error, got: %v", err)
	}
}

func TestConfig_Validate_EmptyEnvDefaultsDev(t *testing.T) {
	cfg := Config{
		Env: "", // empty = not production
	}
	// Empty Env is not "production" so relaxed rules apply
	err := cfg.Validate()
	if err != nil {
		t.Errorf("empty Env should not trigger production checks, got: %v", err)
	}
}

func TestConfig_GetJWTSecret(t *testing.T) {
	cfg := Config{JWTSecret: "my-secret"}
	if got := cfg.GetJWTSecret(); got != "my-secret" {
		t.Errorf("GetJWTSecret() = %q, want %q", got, "my-secret")
	}
}

func TestConfig_Defaults_Port(t *testing.T) {
	// Verify the default port value when constructing manually
	cfg := Config{Port: "8080"}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
}

func TestConfig_Defaults_MaxUploadSizeMB(t *testing.T) {
	cfg := Config{MaxUploadSizeMB: 100}
	if cfg.MaxUploadSizeMB <= 0 {
		t.Error("MaxUploadSizeMB should be > 0")
	}
}

func TestConfig_ValidateProduction_InvalidJWKSURL(t *testing.T) {
	cfg := Config{
		Env:            "production",
		JWKSURL:        "://bad-url",
		InternalSecret: strings.Repeat("x", 32),
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for invalid JWKS URL")
	}
}

// clearConfigEnv unsets all env vars that Load() reads so defaults take effect.
// Each call uses t.Setenv, which restores the original value after the test.
func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"DATABASE_URL", "S3_BUCKET", "S3_REGION", "S3_CLOUDFRONT_DOMAIN",
		"S3_ENDPOINT", "S3_PUBLIC_ENDPOINT",
		"JWT_SECRET", "JWT_ACCESS_TTL", "JWT_REFRESH_TTL", "JWKS_URL",
		"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
		"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
		"SES_REGION", "SES_FROM_ADDRESS", "SES_REPLY_TO",
		"SMTP_HOST", "SMTP_PORT", "SMTP_FROM", "SMTP_USER", "SMTP_PASSWORD",
		"BASE_URL", "FRONTEND_URL", "PORT",
		"ECR_REGISTRY", "ADMIN_SECRET", "INTERNAL_SECRET",
		"RATE_LIMIT_RPS", "MAX_UPLOAD_SIZE_MB", "VERIFICATION_WORKERS",
		"ENV",
	} {
		t.Setenv(key, "")
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearConfigEnv(t)

	// Load requires DATABASE_URL and (JWT_SECRET or JWKS_URL)
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/testdb")
	t.Setenv("JWT_SECRET", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.MaxUploadSizeMB != 100 {
		t.Errorf("MaxUploadSizeMB = %d, want 100", cfg.MaxUploadSizeMB)
	}
	if cfg.RateLimitRPS != 10 {
		t.Errorf("RateLimitRPS = %d, want 10", cfg.RateLimitRPS)
	}
	if cfg.VerificationWorkers != 4 {
		t.Errorf("VerificationWorkers = %d, want 4", cfg.VerificationWorkers)
	}
	if cfg.Env != "development" {
		t.Errorf("Env = %q, want %q", cfg.Env, "development")
	}
	if cfg.BaseURL != "https://registry.hivemind.rithul.dev" {
		t.Errorf("BaseURL = %q, want default", cfg.BaseURL)
	}
	if cfg.S3Bucket != "hivemind-registry-packages" {
		t.Errorf("S3Bucket = %q, want default", cfg.S3Bucket)
	}
	if cfg.S3Region != "us-east-1" {
		t.Errorf("S3Region = %q, want %q", cfg.S3Region, "us-east-1")
	}
	if cfg.JWTAccessTTL != 24*time.Hour {
		t.Errorf("JWTAccessTTL = %v, want 24h", cfg.JWTAccessTTL)
	}
	if cfg.JWTRefreshTTL != 720*time.Hour {
		t.Errorf("JWTRefreshTTL = %v, want 720h", cfg.JWTRefreshTTL)
	}
}

func TestLoad_CustomEnv(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("DATABASE_URL", "postgres://custom:custom@db:5432/mydb")
	t.Setenv("JWT_SECRET", "custom-secret")
	t.Setenv("PORT", "9090")
	t.Setenv("BASE_URL", "https://custom.example.com")
	t.Setenv("MAX_UPLOAD_SIZE_MB", "500")
	t.Setenv("RATE_LIMIT_RPS", "50")
	t.Setenv("ENV", "production")
	t.Setenv("JWKS_URL", "https://auth.example.com/.well-known/jwks.json")
	t.Setenv("INTERNAL_SECRET", strings.Repeat("s", 32))
	t.Setenv("S3_BUCKET", "my-bucket")
	t.Setenv("JWT_ACCESS_TTL", "1h")
	t.Setenv("VERIFICATION_WORKERS", "8")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.DatabaseURL != "postgres://custom:custom@db:5432/mydb" {
		t.Errorf("DatabaseURL = %q, want custom value", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "custom-secret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "custom-secret")
	}
	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9090")
	}
	if cfg.BaseURL != "https://custom.example.com" {
		t.Errorf("BaseURL = %q, want custom value", cfg.BaseURL)
	}
	if cfg.MaxUploadSizeMB != 500 {
		t.Errorf("MaxUploadSizeMB = %d, want 500", cfg.MaxUploadSizeMB)
	}
	if cfg.RateLimitRPS != 50 {
		t.Errorf("RateLimitRPS = %d, want 50", cfg.RateLimitRPS)
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q, want %q", cfg.Env, "production")
	}
	if cfg.JWKSURL != "https://auth.example.com/.well-known/jwks.json" {
		t.Errorf("JWKSURL = %q, want custom value", cfg.JWKSURL)
	}
	if cfg.S3Bucket != "my-bucket" {
		t.Errorf("S3Bucket = %q, want %q", cfg.S3Bucket, "my-bucket")
	}
	if cfg.JWTAccessTTL != 1*time.Hour {
		t.Errorf("JWTAccessTTL = %v, want 1h", cfg.JWTAccessTTL)
	}
	if cfg.VerificationWorkers != 8 {
		t.Errorf("VerificationWorkers = %d, want 8", cfg.VerificationWorkers)
	}
}

func TestGetEnv_Fallback(t *testing.T) {
	got := getEnv("NONEXISTENT_KEY_XYZ", "fallback")
	if got != "fallback" {
		t.Errorf("getEnv() = %q, want %q", got, "fallback")
	}
}

func TestGetEnvInt_Fallback(t *testing.T) {
	got := getEnvInt("NONEXISTENT_KEY_XYZ", 42)
	if got != 42 {
		t.Errorf("getEnvInt() = %d, want 42", got)
	}
}

func TestGetEnvInt_ParsesInt(t *testing.T) {
	t.Setenv("TEST_INT_KEY", "99")
	got := getEnvInt("TEST_INT_KEY", 0)
	if got != 99 {
		t.Errorf("getEnvInt() = %d, want 99", got)
	}
}

func TestParseDuration_Valid(t *testing.T) {
	got := parseDuration("2h", 1*time.Hour)
	if got != 2*time.Hour {
		t.Errorf("parseDuration(\"2h\") = %v, want 2h", got)
	}
}

func TestParseDuration_Invalid(t *testing.T) {
	fallback := 1 * time.Hour
	got := parseDuration("not-a-duration", fallback)
	if got != fallback {
		t.Errorf("parseDuration(\"not-a-duration\") = %v, want %v", got, fallback)
	}
}

func TestParseDuration_Empty(t *testing.T) {
	fallback := 30 * time.Minute
	got := parseDuration("", fallback)
	if got != fallback {
		t.Errorf("parseDuration(\"\") = %v, want %v", got, fallback)
	}
}

// ── Load error cases ─────────────────────────────────────────────────────

func TestLoad_MissingDatabaseURL(t *testing.T) {
	clearConfigEnv(t)
	// Don't set DATABASE_URL; set JWT_SECRET so that's not the error
	t.Setenv("JWT_SECRET", "test-secret")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("error should mention DATABASE_URL, got: %v", err)
	}
}

func TestLoad_MissingJWTSecretAndJWKS(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/testdb")
	// Don't set JWT_SECRET or JWKS_URL

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when both JWT_SECRET and JWKS_URL are missing")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") || !strings.Contains(err.Error(), "JWKS_URL") {
		t.Errorf("error should mention JWT_SECRET and JWKS_URL, got: %v", err)
	}
}

func TestLoad_JWKSOnly(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/testdb")
	t.Setenv("JWKS_URL", "http://localhost:3001/auth/.well-known/jwks.json")
	// No JWT_SECRET — should succeed in dev mode

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with JWKS_URL only should succeed in dev, got: %v", err)
	}
	if cfg.JWKSURL != "http://localhost:3001/auth/.well-known/jwks.json" {
		t.Errorf("JWKSURL = %q, want custom value", cfg.JWKSURL)
	}
	if cfg.JWTSecret != "" {
		t.Errorf("JWTSecret should be empty, got %q", cfg.JWTSecret)
	}
}

func TestLoad_ProductionValidationFails(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/testdb")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ENV", "production")
	t.Setenv("JWKS_URL", "http://insecure.example.com/jwks")
	t.Setenv("INTERNAL_SECRET", strings.Repeat("x", 32))

	// Production with HTTP JWKS URL should fail validation
	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for HTTP JWKS URL in production")
	}
	if !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("error should mention HTTPS, got: %v", err)
	}
}

func TestLoad_ProductionShortSecret(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/testdb")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ENV", "production")
	t.Setenv("INTERNAL_SECRET", "short")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for short INTERNAL_SECRET in production")
	}
	if !strings.Contains(err.Error(), "32") {
		t.Errorf("error should mention 32 characters, got: %v", err)
	}
}

func TestLoad_SMTPEnvVars(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/testdb")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("SMTP_HOST", "mailhog")
	t.Setenv("SMTP_PORT", "1025")
	t.Setenv("SMTP_FROM", "noreply@test.local")
	t.Setenv("SMTP_USER", "smtpuser")
	t.Setenv("SMTP_PASSWORD", "smtppass")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.SMTPHost != "mailhog" {
		t.Errorf("SMTPHost = %q, want %q", cfg.SMTPHost, "mailhog")
	}
	if cfg.SMTPPort != "1025" {
		t.Errorf("SMTPPort = %q, want %q", cfg.SMTPPort, "1025")
	}
	if cfg.SMTPFrom != "noreply@test.local" {
		t.Errorf("SMTPFrom = %q, want %q", cfg.SMTPFrom, "noreply@test.local")
	}
	if cfg.SMTPUser != "smtpuser" {
		t.Errorf("SMTPUser = %q, want %q", cfg.SMTPUser, "smtpuser")
	}
	if cfg.SMTPPassword != "smtppass" {
		t.Errorf("SMTPPassword = %q, want %q", cfg.SMTPPassword, "smtppass")
	}
}

func TestLoad_OAuthEnvVars(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/testdb")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("GITHUB_CLIENT_ID", "gh-id")
	t.Setenv("GITHUB_CLIENT_SECRET", "gh-secret")
	t.Setenv("GOOGLE_CLIENT_ID", "g-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "g-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.GitHubClientID != "gh-id" {
		t.Errorf("GitHubClientID = %q, want %q", cfg.GitHubClientID, "gh-id")
	}
	if cfg.GitHubClientSecret != "gh-secret" {
		t.Errorf("GitHubClientSecret = %q, want %q", cfg.GitHubClientSecret, "gh-secret")
	}
	if cfg.GoogleClientID != "g-id" {
		t.Errorf("GoogleClientID = %q, want %q", cfg.GoogleClientID, "g-id")
	}
	if cfg.GoogleClientSecret != "g-secret" {
		t.Errorf("GoogleClientSecret = %q, want %q", cfg.GoogleClientSecret, "g-secret")
	}
}

func TestLoad_S3EnvVars(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/testdb")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("S3_ENDPOINT", "http://localhost:4566")
	t.Setenv("S3_PUBLIC_ENDPOINT", "http://localhost:4566")
	t.Setenv("S3_CLOUDFRONT_DOMAIN", "cdn.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.S3Endpoint != "http://localhost:4566" {
		t.Errorf("S3Endpoint = %q, want %q", cfg.S3Endpoint, "http://localhost:4566")
	}
	if cfg.S3PublicEndpoint != "http://localhost:4566" {
		t.Errorf("S3PublicEndpoint = %q, want %q", cfg.S3PublicEndpoint, "http://localhost:4566")
	}
	if cfg.S3CloudFrontDomain != "cdn.example.com" {
		t.Errorf("S3CloudFrontDomain = %q, want %q", cfg.S3CloudFrontDomain, "cdn.example.com")
	}
}

func TestLoad_FrontendURL(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/testdb")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("FRONTEND_URL", "https://app.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.FrontendURL != "https://app.example.com" {
		t.Errorf("FrontendURL = %q, want %q", cfg.FrontendURL, "https://app.example.com")
	}
}

func TestLoad_AdminAndECR(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/testdb")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ADMIN_SECRET", "admin123")
	t.Setenv("ECR_REGISTRY", "123456789.dkr.ecr.us-east-1.amazonaws.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.AdminSecret != "admin123" {
		t.Errorf("AdminSecret = %q, want %q", cfg.AdminSecret, "admin123")
	}
	if cfg.ECRRegistry != "123456789.dkr.ecr.us-east-1.amazonaws.com" {
		t.Errorf("ECRRegistry = %q, want custom value", cfg.ECRRegistry)
	}
}

func TestLoad_SESEnvVars(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/testdb")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("SES_REGION", "eu-west-1")
	t.Setenv("SES_FROM_ADDRESS", "noreply@test.com")
	t.Setenv("SES_REPLY_TO", "support@test.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.SESRegion != "eu-west-1" {
		t.Errorf("SESRegion = %q, want %q", cfg.SESRegion, "eu-west-1")
	}
	if cfg.SESFromAddress != "noreply@test.com" {
		t.Errorf("SESFromAddress = %q, want %q", cfg.SESFromAddress, "noreply@test.com")
	}
	if cfg.SESReplyTo != "support@test.com" {
		t.Errorf("SESReplyTo = %q, want %q", cfg.SESReplyTo, "support@test.com")
	}
}

func TestLoad_JWTRefreshTTLCustom(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/testdb")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_REFRESH_TTL", "168h")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.JWTRefreshTTL != 168*time.Hour {
		t.Errorf("JWTRefreshTTL = %v, want 168h", cfg.JWTRefreshTTL)
	}
}

func TestLoad_InvalidDurationFallsBack(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/testdb")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_ACCESS_TTL", "not-a-duration")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	// Invalid duration should fall back to default (24h)
	if cfg.JWTAccessTTL != 24*time.Hour {
		t.Errorf("JWTAccessTTL = %v, want 24h (default fallback)", cfg.JWTAccessTTL)
	}
}
