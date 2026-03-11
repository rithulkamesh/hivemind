package packages

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/auth"
	"github.com/rithul/hivemind/registry/api/internal/config"
	"github.com/rithul/hivemind/registry/api/internal/db"
)

// Storage provides S3 upload and presigned URLs (optional for read-only).
type Storage interface {
	PresignedDownloadURL(key string) (string, error)
	Upload(key string, body []byte) error
}

// EmailSender sends transactional emails (optional).
type EmailSender interface {
	SendPublishSuccess(to, name, version string) error
	SendPublishFailed(to, name, version, report string) error
}

type Handler struct {
	q      *db.Queries
	store  Storage
	email  EmailSender
	verify *Verifier
	cfg    *config.Config
}

func NewHandler(q *db.Queries, store Storage, email EmailSender, verify *Verifier, cfg *config.Config) *Handler {
	return &Handler{q: q, store: store, email: email, verify: verify, cfg: cfg}
}

// NormalizeName normalizes package name per PEP 503 (lowercase, replace [-_.] with -).
func NormalizeName(name string) string {
	b := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch c {
		case '_', '.':
			b = append(b, '-')
		default:
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			b = append(b, c)
		}
	}
	// strip leading/trailing '-'
	for len(b) > 0 && b[0] == '-' {
		b = b[1:]
	}
	for len(b) > 0 && b[len(b)-1] == '-' {
		b = b[:len(b)-1]
	}
	return string(b)
}

func (h *Handler) SimpleIndex(w http.ResponseWriter, r *http.Request) {
	RenderSimpleIndex(w, r, h.q, h.cfg.BaseURL)
}

func (h *Handler) SimplePackageIndex(w http.ResponseWriter, r *http.Request) {
	name := NormalizeName(chi.URLParam(r, "name"))
	RenderSimplePackageIndex(w, r, name, h.q, h.cfg.BaseURL)
}

func (h *Handler) SimpleFileRedirect(w http.ResponseWriter, r *http.Request) {
	name := NormalizeName(chi.URLParam(r, "name"))
	filename := chi.URLParam(r, "filename")
	if h.store == nil {
		http.Error(w, "downloads not configured", http.StatusServiceUnavailable)
		return
	}
	url, err := GetPresignedURLForFile(r.Context(), h.q, name, filename, h.store)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) ListPackages(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit
	var nf pgtype.Text
	if namespace != "" {
		nf.String = namespace
		nf.Valid = true
	}
	pkgs, err := h.q.ListPackages(r.Context(), db.ListPackagesParams{
		NamespaceFilter: nf,
		LimitVal:        int32(limit),
		OffsetVal:       int32(offset),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"packages": pkgs, "page": page})
}

func (h *Handler) GetPackage(w http.ResponseWriter, r *http.Request) {
	name := NormalizeName(chi.URLParam(r, "name"))
	pkg, err := h.q.GetPackageByNamespaceName(r.Context(), db.GetPackageByNamespaceNameParams{
		Namespace: pgtype.Text{},
		Name:      name,
	})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pkg)
}

func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	name := NormalizeName(chi.URLParam(r, "name"))
	version := chi.URLParam(r, "version")
	pv, err := h.q.GetPackageVersion(r.Context(), db.GetPackageVersionParams{
		Namespace: pgtype.Text{},
		Name:      name,
		Version:   version,
	})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pv)
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.q.GetGlobalStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) CreatePackage(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		Name, DisplayName, Description, Homepage, Repository, License string
		Keywords                                                     []string
		Namespace                                                    *string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	name := NormalizeName(req.Name)
	if name == "" {
		http.Error(w, "invalid name", http.StatusBadRequest)
		return
	}
	var ns pgtype.Text
	var ownerOrg pgtype.UUID
	if req.Namespace != nil && *req.Namespace != "" {
		ns.String = *req.Namespace
		ns.Valid = true
		// resolve org and set owner_org_id
	}
	var ownerUser pgtype.UUID
	ownerUser.Bytes = userID
	ownerUser.Valid = true
	_, err := h.q.CreatePackage(r.Context(), db.CreatePackageParams{
		Name:         name,
		Namespace:    ns,
		DisplayName:  req.DisplayName,
		Description:  pgtype.Text{String: req.Description, Valid: req.Description != ""},
		Homepage:     pgtype.Text{String: req.Homepage, Valid: req.Homepage != ""},
		Repository:   pgtype.Text{String: req.Repository, Valid: req.Repository != ""},
		License:      pgtype.Text{String: req.License, Valid: req.License != ""},
		Keywords:     req.Keywords,
		OwnerUserID:  ownerUser,
		OwnerOrgID:   ownerOrg,
	})
	if err != nil {
		http.Error(w, "package exists or error", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "upload not implemented", http.StatusNotImplemented)
}

func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "publish not implemented", http.StatusNotImplemented)
}

func (h *Handler) Yank(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) DeleteVersion(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) ListOrgPackages(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	org, err := h.q.GetOrgByName(r.Context(), slug)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var nf pgtype.Text
	nf.String = org.Name
	nf.Valid = true
	pkgs, err := h.q.ListPackages(r.Context(), db.ListPackagesParams{
		NamespaceFilter: nf,
		LimitVal:        100,
		OffsetVal:       0,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pkgs)
}

func (h *Handler) MyDownloads(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) PackageDownloads(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) VerificationQueue(w http.ResponseWriter, r *http.Request) {
	list, err := h.q.ListPendingVerifications(r.Context(), 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) TrustPackage(w http.ResponseWriter, r *http.Request) {
	name := NormalizeName(chi.URLParam(r, "name"))
	pkg, err := h.q.GetPackageByNamespaceName(r.Context(), db.GetPackageByNamespaceNameParams{Name: name})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	_, err = h.q.SetPackageTrusted(r.Context(), pkg.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) VerifyPackage(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) RemovePackage(w http.ResponseWriter, r *http.Request) {
	name := NormalizeName(chi.URLParam(r, "name"))
	pkg, err := h.q.GetPackageByNamespaceName(r.Context(), db.GetPackageByNamespaceNameParams{Name: name})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err := h.q.DeletePackage(r.Context(), pkg.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
