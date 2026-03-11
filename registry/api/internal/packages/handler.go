package packages

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/auth"
	"github.com/rithul/hivemind/registry/api/internal/config"
	"github.com/rithul/hivemind/registry/api/internal/db"
	"github.com/rithul/hivemind/registry/api/internal/storage"
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
		Keywords                                                      []string
		Namespace                                                     *string
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
		Name:        name,
		Namespace:   ns,
		DisplayName: req.DisplayName,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
		Homepage:    pgtype.Text{String: req.Homepage, Valid: req.Homepage != ""},
		Repository:  pgtype.Text{String: req.Repository, Valid: req.Repository != ""},
		License:     pgtype.Text{String: req.License, Valid: req.License != ""},
		Keywords:    req.Keywords,
		OwnerUserID: ownerUser,
		OwnerOrgID:  ownerOrg,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			http.Error(w, "package name already exists", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		http.Error(w, "upload not configured (no storage)", http.StatusServiceUnavailable)
		return
	}
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	name := NormalizeName(chi.URLParam(r, "name"))
	if name == "" {
		// Fallback for generic upload endpoint (Twine/CLI)
		name = NormalizeName(r.FormValue("name"))
	}
	if name == "" {
		http.Error(w, "invalid package name", http.StatusBadRequest)
		return
	}
	maxBytes := int64(h.cfg.MaxUploadSizeMB) * 1024 * 1024
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		http.Error(w, "failed to parse form or file too large", http.StatusBadRequest)
		return
	}
	version := strings.TrimSpace(r.FormValue("version"))
	if version == "" {
		http.Error(w, "version required", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		// Twine uses "content" field
		file, header, err = r.FormFile("content")
		if err != nil {
			http.Error(w, "file required (file or content)", http.StatusBadRequest)
			return
		}
	}
	defer file.Close()
	body, err := io.ReadAll(io.LimitReader(file, maxBytes))
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}
	filename := header.Filename
	if filename == "" || strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}
	var ns pgtype.Text
	pkg, err := h.q.GetPackageByNamespaceName(r.Context(), db.GetPackageByNamespaceNameParams{Namespace: ns, Name: name})
	if err != nil {
		http.Error(w, "package not found", http.StatusNotFound)
		return
	}
	// Create version (pending)
	var up pgtype.UUID
	up.Bytes = userID
	up.Valid = true
	pv, err := h.q.CreatePackageVersion(r.Context(), db.CreatePackageVersionParams{
		PackageID:          pkg.ID,
		Version:            version,
		RequiresPython:     pgtype.Text{},
		RequiresHivemind:   pgtype.Text{},
		UploadedBy:         up,
		VerificationStatus: pgtype.Text{String: "pending", Valid: true},
	})
	if err != nil {
		http.Error(w, "version exists or error", http.StatusConflict)
		return
	}
	s3Key := storage.Key(pkg.Namespace.String, name, version, filename)
	if err := h.store.Upload(s3Key, body); err != nil {
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}
	sum := sha256.Sum256(body)
	sha256Sum := hex.EncodeToString(sum[:])
	filetype := "sdist"
	if strings.HasSuffix(strings.ToLower(filename), ".whl") {
		filetype = "wheel"
	}
	_, err = h.q.CreatePackageFile(r.Context(), db.CreatePackageFileParams{
		VersionID:     pv.ID,
		Filename:      filename,
		Filetype:      filetype,
		PythonVersion: pgtype.Text{},
		Abi:           pgtype.Text{},
		Platform:      pgtype.Text{},
		SizeBytes:     int64(len(body)),
		Sha256:        sha256Sum,
		Md5:           "",
		S3Key:         s3Key,
	})
	if err != nil {
		http.Error(w, "failed to record file", http.StatusInternalServerError)
		return
	}
	// Mark published so it appears in Simple index (minimal path; no verification run)
	_, _ = h.q.UpdatePackageVersionVerification(r.Context(), db.UpdatePackageVersionVerificationParams{
		ID:                 pv.ID,
		VerificationStatus: pgtype.Text{String: "passed", Valid: true},
		VerificationReport: nil,
		Published:          pgtype.Bool{Bool: true, Valid: true},
		ToolCount:          pgtype.Int4{},
		SigstoreBundle:     nil,
	})
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "publish not implemented", http.StatusNotImplemented)
}

func (h *Handler) UpdatePackage(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	name := NormalizeName(chi.URLParam(r, "name"))
	pkg, err := h.q.GetPackageByNamespaceName(r.Context(), db.GetPackageByNamespaceNameParams{Name: name})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !pkg.OwnerUserID.Valid || pkg.OwnerUserID.Bytes != userID {
		http.Error(w, "forbidden: not owner", http.StatusForbidden)
		return
	}

	var req struct {
		DisplayName string   `json:"display_name"`
		Description string   `json:"description"`
		Homepage    string   `json:"homepage"`
		Repository  string   `json:"repository"`
		License     string   `json:"license"`
		Keywords    []string `json:"keywords"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	updated, err := h.q.UpdatePackage(r.Context(), db.UpdatePackageParams{
		ID:          pkg.ID,
		DisplayName: req.DisplayName,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
		Homepage:    pgtype.Text{String: req.Homepage, Valid: req.Homepage != ""},
		Repository:  pgtype.Text{String: req.Repository, Valid: req.Repository != ""},
		License:     pgtype.Text{String: req.License, Valid: req.License != ""},
		Keywords:    req.Keywords,
		Verified:    pkg.Verified,
		Trusted:     pkg.Trusted,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *Handler) DeletePackage(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	name := NormalizeName(chi.URLParam(r, "name"))
	pkg, err := h.q.GetPackageByNamespaceName(r.Context(), db.GetPackageByNamespaceNameParams{Name: name})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !pkg.OwnerUserID.Valid || pkg.OwnerUserID.Bytes != userID {
		http.Error(w, "forbidden: not owner", http.StatusForbidden)
		return
	}

	if err := h.q.DeletePackage(r.Context(), pkg.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Yank(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	name := NormalizeName(chi.URLParam(r, "name"))
	version := chi.URLParam(r, "version")

	pkg, err := h.q.GetPackageByNamespaceName(r.Context(), db.GetPackageByNamespaceNameParams{Name: name})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !pkg.OwnerUserID.Valid || pkg.OwnerUserID.Bytes != userID {
		http.Error(w, "forbidden: not owner", http.StatusForbidden)
		return
	}

	pv, err := h.q.GetPackageVersion(r.Context(), db.GetPackageVersionParams{
		Namespace: pgtype.Text{},
		Name:      name,
		Version:   version,
	})
	if err != nil {
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}

	var req struct{ Reason string }
	json.NewDecoder(r.Body).Decode(&req)

	reason := pgtype.Text{String: req.Reason, Valid: req.Reason != ""}
	_, err = h.q.YankPackageVersion(r.Context(), db.YankPackageVersionParams{
		ID:         pv.ID,
		YankReason: reason,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteVersion(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	name := NormalizeName(chi.URLParam(r, "name"))
	version := chi.URLParam(r, "version")

	pkg, err := h.q.GetPackageByNamespaceName(r.Context(), db.GetPackageByNamespaceNameParams{Name: name})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !pkg.OwnerUserID.Valid || pkg.OwnerUserID.Bytes != userID {
		http.Error(w, "forbidden: not owner", http.StatusForbidden)
		return
	}

	pv, err := h.q.GetPackageVersion(r.Context(), db.GetPackageVersionParams{
		Namespace: pgtype.Text{},
		Name:      name,
		Version:   version,
	})
	if err != nil {
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}

	err = h.q.DeletePackageVersion(r.Context(), pv.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
