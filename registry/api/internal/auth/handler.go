package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/config"
	"github.com/rithul/hivemind/registry/api/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	q    *db.Queries
	cfg  *config.Config
}

func NewHandler(q *db.Queries, cfg *config.Config) *Handler {
	return &Handler{q: q, cfg: cfg}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	user, err := h.q.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if user.PasswordHash.String == "" {
		http.Error(w, "use OAuth to sign in", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	access, _ := IssueAccessToken(h.cfg.JWTSecret, user.ID.String(), user.Username, h.cfg.JWTAccessTTL)
	refresh, _ := IssueRefreshToken(h.cfg.JWTSecret, user.ID.String(), user.Username, h.cfg.JWTRefreshTTL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(h.cfg.JWTAccessTTL.Seconds()),
	})
}

type registerRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_, err = h.q.CreateUser(r.Context(), db.CreateUserParams{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "email or username already in use", http.StatusConflict)
			return
		}
		slog.Error("register create user", "err", err)
		http.Error(w, "registration failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	claims, err := VerifyToken(h.cfg.JWTSecret, req.RefreshToken)
	if err != nil || claims.Subject != "refresh" {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}
	access, _ := IssueAccessToken(h.cfg.JWTSecret, claims.UserID, claims.Username, h.cfg.JWTAccessTTL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResponse{AccessToken: access, ExpiresIn: int(h.cfg.JWTAccessTTL.Seconds())})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := "https://github.com/login/oauth/authorize?client_id=" + h.cfg.GitHubClientID + "&scope=read:user%20user:email"
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	// Exchange code for token, get user, upsert oauth_identity + user, issue JWT
	http.Error(w, "OAuth callback not implemented", http.StatusNotImplemented)
}

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) SAMLMetadata(w http.ResponseWriter, r *http.Request) {
	orgSlug := chi.URLParam(r, "orgSlug")
	_ = orgSlug
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) SAMLACS(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) OIDCLogin(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
