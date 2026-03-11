package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rithul/hivemind/registry/api/internal/config"
	"github.com/rithul/hivemind/registry/api/internal/db"
	"github.com/rithul/hivemind/registry/api/internal/health"
	"github.com/rithul/hivemind/registry/api/internal/middleware"
	"github.com/rithul/hivemind/registry/api/internal/packages"
	"github.com/rithul/hivemind/registry/api/internal/auth"
	"github.com/rithul/hivemind/registry/api/internal/users"
	"github.com/rithul/hivemind/registry/api/internal/orgs"
	"github.com/rithul/hivemind/registry/api/internal/search"
	"github.com/rithul/hivemind/registry/api/internal/docker"
	"github.com/rithul/hivemind/registry/api/internal/webhooks"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
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

	queries := db.New(pool)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(log.Logger))
	r.Use(middleware.Recover)
	r.Use(middleware.CORS())

	// Public
	r.Get("/health", health.Liveness)
	r.Get("/ready", health.Readiness(pool))

	// PyPI simple index (public)
	verifier := packages.NewVerifier(queries, cfg.VerificationWorkers)
	pkgHandler := packages.NewHandler(queries, nil, nil, verifier, cfg)
	r.Get("/simple/", pkgHandler.SimpleIndex)
	r.Get("/simple/{name}/", pkgHandler.SimplePackageIndex)
	r.Get("/simple/{name}/{filename}", pkgHandler.SimpleFileRedirect)

	// Public API
	searchHandler := search.NewHandler(queries)
	r.Get("/api/v1/packages", pkgHandler.ListPackages)
	dockerHandler := docker.NewHandler(queries)
	r.Get("/api/v1/packages/{name}/images", dockerHandler.ListImages)
	r.Get("/api/v1/packages/{name}", pkgHandler.GetPackage)
	r.Get("/api/v1/packages/{name}/{version}", pkgHandler.GetVersion)
	r.Get("/api/v1/search", searchHandler.Search)
	r.Get("/api/v1/stats", pkgHandler.Stats)
	webhookHandler := webhooks.NewHandler()
	r.Post("/webhooks/github", webhookHandler.GitHub)

	// Auth routes
	authHandler := auth.NewHandler(queries, cfg)
	r.Get("/auth/github", authHandler.GitHubLogin)
	r.Get("/auth/github/callback", authHandler.GitHubCallback)
	r.Get("/auth/google", authHandler.GoogleLogin)
	r.Get("/auth/google/callback", authHandler.GoogleCallback)
	r.Post("/auth/login", authHandler.Login)
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/verify-email", authHandler.VerifyEmail)
	r.Post("/auth/resend-verification", authHandler.ResendVerification)
	r.Post("/auth/refresh", authHandler.Refresh)
	r.Post("/auth/logout", authHandler.Logout)
	r.Get("/auth/saml/{orgSlug}/metadata", authHandler.SAMLMetadata)
	r.Post("/auth/saml/{orgSlug}/acs", authHandler.SAMLACS)
	r.Get("/auth/oidc/{orgSlug}/login", authHandler.OIDCLogin)
	r.Get("/auth/oidc/{orgSlug}/callback", authHandler.OIDCCallback)

	// Authenticated API (JWT or API key)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(cfg.RateLimitRPS))
		r.Use(auth.RequireAuth(cfg))
		r.Get("/api/v1/me", users.NewHandler(queries).GetMe)
		r.Put("/api/v1/me", users.NewHandler(queries).UpdateMe)
		r.Post("/api/v1/me/2fa/setup", users.NewHandler(queries).Setup2FA)
		r.Post("/api/v1/me/2fa/verify", users.NewHandler(queries).Verify2FA)
		r.Get("/api/v1/me/api-keys", users.NewHandler(queries).ListAPIKeys)
		r.Post("/api/v1/me/api-keys", users.NewHandler(queries).CreateAPIKey)
		r.Delete("/api/v1/me/api-keys/{id}", users.NewHandler(queries).RevokeAPIKey)

		r.Post("/api/v1/packages", pkgHandler.CreatePackage)
		r.Post("/api/v1/packages/{name}/upload", pkgHandler.Upload)
		r.Post("/api/v1/packages/{name}/{version}/publish", pkgHandler.Publish)
		r.Post("/api/v1/packages/{name}/{version}/yank", pkgHandler.Yank)
		r.Delete("/api/v1/packages/{name}/{version}", pkgHandler.DeleteVersion)

		r.Get("/api/v1/orgs", orgs.NewHandler(queries).ListOrgs)
		r.Post("/api/v1/orgs", orgs.NewHandler(queries).CreateOrg)
		r.Get("/api/v1/orgs/{slug}", orgs.NewHandler(queries).GetOrg)
		r.Put("/api/v1/orgs/{slug}", orgs.NewHandler(queries).UpdateOrg)
		r.Get("/api/v1/orgs/{slug}/members", orgs.NewHandler(queries).ListMembers)
		r.Post("/api/v1/orgs/{slug}/members/invite", orgs.NewHandler(queries).InviteMember)
		r.Delete("/api/v1/orgs/{slug}/members/{userID}", orgs.NewHandler(queries).RemoveMember)
		r.Put("/api/v1/orgs/{slug}/sso", orgs.NewHandler(queries).ConfigureSSO)

		r.Get("/api/v1/orgs/{slug}/packages", pkgHandler.ListOrgPackages)
		r.Get("/api/v1/me/downloads", pkgHandler.MyDownloads)
		r.Get("/api/v1/packages/{name}/downloads", pkgHandler.PackageDownloads)
	})

	// Admin (require admin scope/role)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(cfg.RateLimitRPS))
		r.Use(auth.RequireAuth(cfg))
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
