package orgs

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

func (h *Handler) ListOrgs(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	orgs, err := h.q.ListOrgsForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgs)
}

func (h *Handler) CreateOrg(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	org, err := h.q.CreateOrg(r.Context(), db.CreateOrgParams{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		BillingEmail: pgtype.Text{},
	})
	if err != nil {
		http.Error(w, "name taken or error", http.StatusConflict)
		return
	}
	_, err = h.q.AddOrgMember(r.Context(), db.AddOrgMemberParams{
		OrgID:  org.ID,
		UserID: userID,
		Role:   "owner",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(org)
}

func (h *Handler) GetOrg(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	org, err := h.q.GetOrgByName(r.Context(), slug)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

func (h *Handler) UpdateOrg(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	org, err := h.q.GetOrgByName(r.Context(), slug)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var req struct {
		DisplayName *string `json:"display_name"`
		BillingEmail *string `json:"billing_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	displayName := org.DisplayName
	billingEmail := org.BillingEmail
	if req.DisplayName != nil {
		displayName = *req.DisplayName
	}
	if req.BillingEmail != nil {
		billingEmail = pgtype.Text{String: *req.BillingEmail, Valid: true}
	}
	_, err = h.q.UpdateOrg(r.Context(), db.UpdateOrgParams{
		ID:              org.ID,
		DisplayName:     displayName,
		BillingEmail:    billingEmail,
		SsoEnabled:      org.SsoEnabled,
		SamlMetadataUrl: org.SamlMetadataUrl,
		OidcIssuer:      org.OidcIssuer,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	org, err := h.q.GetOrgByName(r.Context(), slug)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	members, err := h.q.ListOrgMembers(r.Context(), org.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	org, err := h.q.GetOrgByName(r.Context(), slug)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err := h.q.RemoveOrgMember(r.Context(), db.RemoveOrgMemberParams{OrgID: org.ID, UserID: userID}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ConfigureSSO(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
