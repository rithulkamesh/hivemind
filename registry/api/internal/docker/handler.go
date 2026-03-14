package docker

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/db"
)

type Handler struct {
	q *db.Queries
}

func NewHandler(q *db.Queries) *Handler {
	return &Handler{q: q}
}

func (h *Handler) ListImages(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	pkg, err := h.q.GetPackageByNamespaceName(r.Context(), db.GetPackageByNamespaceNameParams{
		Namespace: pgtype.Text{},
		Name:      name,
	})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	images, err := h.q.ListDockerImagesForPackage(r.Context(), pkg.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if images == nil {
		images = []db.DockerImage{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}
