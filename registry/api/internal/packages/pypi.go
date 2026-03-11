package packages

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/db"
)

func jsonEncoder(w http.ResponseWriter) *json.Encoder {
	return json.NewEncoder(w)
}

// RenderSimpleIndex writes PEP 503 simple index root (HTML or JSON per Accept).
func RenderSimpleIndex(w http.ResponseWriter, r *http.Request, q *db.Queries, baseURL string) {
	if wantsJSON(r) {
		renderSimpleIndexJSON(w, r, q, baseURL)
		return
	}
	// HTML
	pkgs, err := q.ListPackages(r.Context(), db.ListPackagesParams{
		NamespaceFilter: pgtype.Text{},
		LimitVal:        5000,
		OffsetVal:       0,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<!DOCTYPE html><html><head><meta name=\"pypi:repository-version\" content=\"1.0\"></head><body>\n"))
	for _, p := range pkgs {
		w.Write([]byte("<a href=\"" + baseURL + "/simple/" + p.Name + "/\">" + p.Name + "</a><br>\n"))
	}
	w.Write([]byte("</body></html>"))
}

func renderSimpleIndexJSON(w http.ResponseWriter, r *http.Request, q *db.Queries, baseURL string) {
	pkgs, err := q.ListPackages(r.Context(), db.ListPackagesParams{
		NamespaceFilter: pgtype.Text{},
		LimitVal:        5000,
		OffsetVal:       0,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	names := make([]map[string]string, 0, len(pkgs))
	for _, p := range pkgs {
		names = append(names, map[string]string{"name": p.Name})
	}
	w.Header().Set("Content-Type", "application/vnd.pypi.simple.v1+json")
	enc := jsonEncoder(w)
	enc.Encode(map[string]interface{}{"projects": names})
}

// RenderSimplePackageIndex writes PEP 503 package index (links to files).
func RenderSimplePackageIndex(w http.ResponseWriter, r *http.Request, name string, q *db.Queries, baseURL string) {
	if wantsJSON(r) {
		renderSimplePackageIndexJSON(w, r, name, q, baseURL)
		return
	}
	pkg, err := q.GetPackageByNamespaceName(r.Context(), db.GetPackageByNamespaceNameParams{Name: name})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	vers, err := q.ListVersionsForPackage(r.Context(), pkg.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<!DOCTYPE html><html><head><meta name=\"pypi:repository-version\" content=\"1.1\"></head><body>\n"))
	for _, pv := range vers {
		if pv.Published.Valid && !pv.Published.Bool {
			continue
		}
		files, _ := q.ListFilesForVersion(r.Context(), pv.ID)
		for _, f := range files {
			yanked := ""
			if pv.Yanked.Valid && pv.Yanked.Bool && pv.YankReason.Valid {
				yanked = " data-yanked=\"" + escapeHTML(pv.YankReason.String) + "\""
			} else if pv.Yanked.Valid && pv.Yanked.Bool {
				yanked = " data-yanked=\"\""
			}
			reqPy := ""
			if pv.RequiresPython.Valid {
				reqPy = " data-requires-python=\"" + escapeHTML(pv.RequiresPython.String) + "\""
			}
			href := baseURL + "/simple/" + name + "/" + f.Filename + "#sha256=" + f.Sha256
			w.Write([]byte("<a href=\"" + href + "\"" + reqPy + yanked + ">" + f.Filename + "</a><br>\n"))
		}
	}
	w.Write([]byte("</body></html>"))
}

func renderSimplePackageIndexJSON(w http.ResponseWriter, r *http.Request, name string, q *db.Queries, baseURL string) {
	pkg, err := q.GetPackageByNamespaceName(r.Context(), db.GetPackageByNamespaceNameParams{Name: name})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	vers, err := q.ListVersionsForPackage(r.Context(), pkg.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var files []map[string]interface{}
	for _, pv := range vers {
		if pv.Published.Valid && !pv.Published.Bool {
			continue
		}
		flist, _ := q.ListFilesForVersion(r.Context(), pv.ID)
		for _, f := range flist {
			entry := map[string]interface{}{
				"filename": f.Filename,
				"url":      baseURL + "/simple/" + name + "/" + f.Filename,
				"hashes":   map[string]string{"sha256": f.Sha256},
			}
			if pv.RequiresPython.Valid {
				entry["requires-python"] = pv.RequiresPython.String
			}
			if pv.Yanked.Valid && pv.Yanked.Bool {
				entry["yanked"] = true
				if pv.YankReason.Valid {
					entry["yanked-reason"] = pv.YankReason.String
				}
			}
			files = append(files, entry)
		}
	}
	w.Header().Set("Content-Type", "application/vnd.pypi.simple.v1+json")
	enc := jsonEncoder(w)
	enc.Encode(map[string]interface{}{"name": name, "files": files})
}

func wantsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "application/vnd.pypi.simple.v1+json") || strings.Contains(accept, "application/json")
}

func escapeHTML(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;").Replace(s)
}

// GetPresignedURLForFile finds the file by package name and filename and returns a presigned download URL.
func GetPresignedURLForFile(ctx context.Context, q *db.Queries, name, filename string, store Storage) (string, error) {
	pkg, err := q.GetPackageByNamespaceName(ctx, db.GetPackageByNamespaceNameParams{Name: name})
	if err != nil {
		return "", err
	}
	vers, err := q.ListVersionsForPackage(ctx, pkg.ID)
	if err != nil {
		return "", err
	}
	for _, pv := range vers {
		f, err := q.GetPackageFileByVersionAndFilename(ctx, db.GetPackageFileByVersionAndFilenameParams{
			VersionID: pv.ID,
			Filename:  filename,
		})
		if err != nil {
			continue
		}
		return store.PresignedDownloadURL(f.S3Key)
	}
	return "", nil
}
