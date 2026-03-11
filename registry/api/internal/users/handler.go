package users

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/auth"
	"github.com/rithul/hivemind/registry/api/internal/db"
)

type Handler struct {
	q *db.Queries
}

func NewHandler(q *db.Queries) *Handler {
	return &Handler{q: q}
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ctx := r.Context()
	userIDStr := userID.String()
	email := auth.GetEmail(ctx)

	// Prefer registry_profiles (Better Auth flow)
	profile, err := h.q.GetRegistryProfileByUserID(ctx, userIDStr)
	if err == nil {
		me := meResponse{
			ID:       userIDStr,
			Email:    email,
			Username: profile.Username,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(me)
		return
	}
	// First time: create profile if we have email (Better Auth)
	if email != "" {
		username := email
		if len(username) > 64 {
			username = username[:64]
		}
		profile, err = h.q.CreateRegistryProfile(ctx, db.CreateRegistryProfileParams{
			UserID:   userIDStr,
			Username: username,
		})
		if err == nil {
			me := meResponse{
				ID:       userIDStr,
				Email:    email,
				Username: profile.Username,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(me)
			return
		}
	}
	// Legacy: users table (e.g. API key or old JWT)
	user, err := h.q.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	me := meResponse{
		ID:       user.ID.String(),
		Email:    user.Email,
		Username: user.Username,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(me)
}

type meResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ctx := r.Context()
	userIDStr := userID.String()
	var req struct {
		Email    *string `json:"email"`
		Username *string `json:"username"`
		Bio      *string `json:"bio"`
		Website  *string `json:"website"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// Prefer registry_profiles (Better Auth)
	profile, err := h.q.GetRegistryProfileByUserID(ctx, userIDStr)
	if err == nil {
		username := profile.Username
		if req.Username != nil {
			username = *req.Username
		}
		bio := profile.Bio
		if req.Bio != nil {
			bio = pgtype.Text{String: *req.Bio, Valid: true}
		}
		website := profile.Website
		if req.Website != nil {
			website = pgtype.Text{String: *req.Website, Valid: true}
		}
		_, err = h.q.UpdateRegistryProfile(ctx, db.UpdateRegistryProfileParams{
			UserID:   userIDStr,
			Username: username,
			Bio:      bio,
			Website:  website,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	// Legacy: users table
	user, err := h.q.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Username != nil {
		user.Username = *req.Username
	}
	_, err = h.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:            userID,
		Email:         user.Email,
		Username:      user.Username,
		PasswordHash:  user.PasswordHash,
		EmailVerified: user.EmailVerified,
		TotpSecret:    user.TotpSecret,
		TotpEnabled:   user.TotpEnabled,
		LastLoginAt:   user.LastLoginAt,
		Banned:        user.Banned,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Setup2FA(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) Verify2FA(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	keys, err := h.q.ListAPIKeysForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Helper to ensure we return [] instead of null if empty
	if keys == nil {
		keys = []db.ListAPIKeysForUserRow{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func (h *Handler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	raw, hash, prefix, err := auth.GenerateAPIKey()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var orgID pgtype.UUID
	_, err = h.q.CreateAPIKey(r.Context(), db.CreateAPIKeyParams{
		UserID:    userID,
		OrgID:     orgID,
		Name:      req.Name,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Scopes:    req.Scopes,
		ExpiresAt: pgtype.Timestamptz{},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"key":    raw,
		"prefix": prefix,
	})
}

func (h *Handler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	// Optional: verify key belongs to user
	_ = userID
	if err := h.q.RevokeAPIKey(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	list, err := h.q.ListUsersAdmin(r.Context(), db.ListUsersAdminParams{Limit: 100, Offset: 0})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) AdminBanUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.q.SetUserBanned(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
