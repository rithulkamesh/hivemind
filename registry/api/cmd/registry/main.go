package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rithul/hivemind/registry/api/internal/auth"
	"github.com/rithul/hivemind/registry/api/internal/config"
	"github.com/rithul/hivemind/registry/api/internal/db"
	"github.com/rithul/hivemind/registry/api/internal/docker"
	"github.com/rithul/hivemind/registry/api/internal/email"
	"github.com/rithul/hivemind/registry/api/internal/health"
	"github.com/rithul/hivemind/registry/api/internal/middleware"
	"github.com/rithul/hivemind/registry/api/internal/orgs"
	"github.com/rithul/hivemind/registry/api/internal/packages"
	"github.com/rithul/hivemind/registry/api/internal/search"
	"github.com/rithul/hivemind/registry/api/internal/storage"
	"github.com/rithul/hivemind/registry/api/internal/users"
	"github.com/rithul/hivemind/registry/api/internal/webhooks"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("connect to database")
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, cfg.DatabaseURL); err != nil {
		log.Fatal().Err(err).Msg("run migrations")
	}

	if os.Getenv("MIGRATE_ONLY") == "1" {
		log.Info().Msg("migrations complete (MIGRATE_ONLY=1)")
		os.Exit(0)
	}

	queries := db.New(pool)

	// Better Auth: JWKS verifier for JWT validation (optional; else legacy JWT_SECRET)
	if cfg.JWKSURL != "" {
		go func() {
			for {
				jwks, err := auth.NewJWKSVerifier(context.Background(), cfg.JWKSURL)
				if err == nil {
					auth.SetGlobalJWKSVerifier(jwks)
					log.Info().Str("url", cfg.JWKSURL).Msg("JWKS verifier initialized")
					return
				}
				log.Warn().Err(err).Msg("failed to initialize JWKS verifier, retrying in 5s")
				time.Sleep(5 * time.Second)
			}
		}()
	}

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(log.Logger))
	r.Use(middleware.Recover)
	r.Use(middleware.CORSWithAllowlist(cfg.BaseURL, cfg.FrontendURL))

	// Public
	r.Get("/health", health.Liveness)
	r.Get("/ready", health.Readiness(pool))

	// PyPI simple index (public)
	verifier := packages.NewVerifier(queries, cfg.VerificationWorkers)
	var store packages.Storage
	if cfg.S3Bucket != "" {
		s3Store, err := storage.NewS3(ctx, cfg.S3Region, cfg.S3Bucket, cfg.S3CloudFrontDomain, cfg.S3Endpoint, cfg.S3PublicEndpoint)
		if err != nil {
			log.Warn().Err(err).Msg("S3 not configured; upload and file download disabled")
		} else {
			store = s3Store
			if cfg.S3Endpoint != "" {
				log.Info().Str("endpoint", cfg.S3Endpoint).Str("bucket", cfg.S3Bucket).Msg("S3 via custom endpoint (MinIO/LocalStack)")
			} else {
				log.Info().Str("bucket", cfg.S3Bucket).Str("region", cfg.S3Region).Msg("S3 configured (AWS)")
			}
		}
	}
	pkgHandler := packages.NewHandler(queries, store, nil, verifier, cfg)
	r.Get("/simple/", pkgHandler.SimpleIndex)
	r.Get("/simple/{name}/", pkgHandler.SimplePackageIndex)
	r.Get("/simple/{name}/{filename}", pkgHandler.SimpleFileRedirect)

	deviceAuth := auth.NewDeviceAuthManager(queries, cfg.BaseURL)
	r.Post("/api/v1/auth/device/request", middleware.DeviceRateLimit.Middleware(http.HandlerFunc(deviceAuth.RequestDeviceCode)).ServeHTTP)
	r.Post("/api/v1/auth/device/poll", middleware.DeviceRateLimit.Middleware(http.HandlerFunc(deviceAuth.PollDeviceCode)).ServeHTTP)

	// Public API
	searchHandler := search.NewHandler(queries)
	r.Get("/api/v1/packages", pkgHandler.ListPackages)
	dockerHandler := docker.NewHandler(queries)
	r.Get("/api/v1/packages/{name}/images", dockerHandler.ListImages)
	r.Get("/api/v1/packages/{name}/versions", pkgHandler.ListVersions)
	r.Get("/api/v1/packages/{name}", pkgHandler.GetPackage)
	r.Get("/api/v1/packages/{name}/{version}", pkgHandler.GetVersion)
	r.Get("/api/v1/packages/{name}/versions/{version}/status", pkgHandler.GetVersionStatus)
	r.Get("/api/v1/search", middleware.SearchRateLimit.Middleware(http.HandlerFunc(searchHandler.Search)).ServeHTTP)
	r.Get("/api/v1/stats", pkgHandler.Stats)
	webhookHandler := webhooks.NewHandler()
	r.Post("/webhooks/github", webhookHandler.GitHub)

	// Internal endpoints (Better Auth → API, protected by X-Internal-Secret)
	var emailSender email.Sender
	if cfg.SMTPHost != "" {
		smtpSender, err := email.NewSMTP(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom, cfg.SMTPUser, cfg.SMTPPassword)
		if err != nil {
			log.Warn().Err(err).Msg("SMTP not configured; internal email endpoints will no-op")
		} else {
			emailSender = smtpSender
			log.Info().Str("host", cfg.SMTPHost).Str("port", cfg.SMTPPort).Msg("email via SMTP (e.g. Mailhog)")
		}
	}
	if emailSender == nil && cfg.SESRegion != "" {
		ses, err := email.NewSES(ctx, cfg.SESRegion, cfg.SESFromAddress, cfg.SESReplyTo)
		if err != nil {
			log.Warn().Err(err).Msg("SES not configured; internal email endpoints will no-op")
		} else {
			emailSender = ses
		}
	}
	internalEmail := &email.InternalHandler{Secret: cfg.InternalSecret, Send: emailSender}
	r.Group(func(r chi.Router) {
		r.Use(internalEmail.RequireInternalSecret)
		r.Post("/internal/email/verify", internalEmail.ServeVerify)
		r.Post("/internal/email/send", internalEmail.ServeSend)
	})

	// Authenticated API (Better Auth JWT via JWKS or legacy JWT or API key)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(cfg.RateLimitRPS))
		r.Use(auth.RequireAuth(cfg, queries))

		// Read-scoped endpoints
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireScope("read"))
			r.Get("/api/v1/me", users.NewHandler(queries).GetMe)
			r.Put("/api/v1/me", users.NewHandler(queries).UpdateMe)
			r.Post("/api/v1/me/2fa/setup", users.NewHandler(queries).Setup2FA)
			r.Post("/api/v1/me/2fa/verify", users.NewHandler(queries).Verify2FA)
			r.Get("/api/v1/me/api-keys", users.NewHandler(queries).ListAPIKeys)
			r.Post("/api/v1/me/api-keys", users.NewHandler(queries).CreateAPIKey)
			r.Delete("/api/v1/me/api-keys/{id}", users.NewHandler(queries).RevokeAPIKey)
			r.Get("/api/v1/orgs", orgs.NewHandler(queries).ListOrgs)
			r.Get("/api/v1/orgs/{slug}", orgs.NewHandler(queries).GetOrg)
			r.Get("/api/v1/orgs/{slug}/members", orgs.NewHandler(queries).ListMembers)
			r.Get("/api/v1/orgs/{slug}/packages", pkgHandler.ListOrgPackages)
			r.Get("/api/v1/me/downloads", pkgHandler.MyDownloads)
			r.Get("/api/v1/packages/{name}/downloads", pkgHandler.PackageDownloads)
		})

		// Publish-scoped endpoints
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireScope("publish"))
			r.Post("/api/v1/auth/device/approve", deviceAuth.ApproveDevice)
			r.Post("/api/v1/packages", pkgHandler.CreatePackage)
			r.Put("/api/v1/packages/{name}", pkgHandler.UpdatePackage)
			r.Delete("/api/v1/packages/{name}", pkgHandler.DeletePackage)
			r.Post("/api/v1/packages/upload", middleware.UploadRateLimit.Middleware(http.HandlerFunc(pkgHandler.Upload)).ServeHTTP) // Generic upload (Twine/CLI)
			r.Post("/api/v1/packages/{name}/upload", middleware.UploadRateLimit.Middleware(http.HandlerFunc(pkgHandler.Upload)).ServeHTTP)
			r.Post("/api/v1/packages/{name}/{version}/publish", pkgHandler.Publish)
			r.Post("/api/v1/packages/{name}/{version}/yank", pkgHandler.Yank)
			r.Delete("/api/v1/packages/{name}/{version}", pkgHandler.DeleteVersion)
			r.Post("/api/v1/orgs", orgs.NewHandler(queries).CreateOrg)
			r.Put("/api/v1/orgs/{slug}", orgs.NewHandler(queries).UpdateOrg)
			r.Post("/api/v1/orgs/{slug}/members/invite", orgs.NewHandler(queries).InviteMember)
			r.Delete("/api/v1/orgs/{slug}/members/{userID}", orgs.NewHandler(queries).RemoveMember)
			r.Put("/api/v1/orgs/{slug}/sso", orgs.NewHandler(queries).ConfigureSSO)
		})
	})

	// Admin (require admin scope/role)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(cfg.RateLimitRPS))
		r.Use(auth.RequireAuth(cfg, queries))
		r.Use(auth.RequireAdmin)
		r.Get("/api/v1/admin/verification-queue", pkgHandler.VerificationQueue)
		r.Post("/api/v1/admin/packages/{name}/trust", pkgHandler.TrustPackage)
		r.Post("/api/v1/admin/packages/{name}/verify", pkgHandler.VerifyPackage)
		r.Post("/api/v1/admin/packages/{name}/remove", pkgHandler.RemovePackage)
		r.Get("/api/v1/admin/users", users.NewHandler(queries).AdminListUsers)
		r.Post("/api/v1/admin/users/{id}/ban", users.NewHandler(queries).AdminBanUser)
	})

	addr := ":" + cfg.Port
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		log.Info().Str("addr", addr).Msg("registry server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown")
	}
	log.Info().Msg("registry stopped")
}
