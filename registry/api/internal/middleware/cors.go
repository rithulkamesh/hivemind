package middleware

import (
	"net/http"
	"net/url"
)

// CORSWithAllowlist adds CORS headers restricted to an explicit set of allowed origins.
// Pass the frontend URL (e.g. "https://registry.hivemind.rithul.dev") and the API base URL.
// Only exact-match origins receive credentialed CORS headers.
func CORSWithAllowlist(allowedOrigins ...string) func(next http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if o != "" {
			// Normalise: strip trailing slash, store origin (scheme+host+port)
			if u, err := url.Parse(o); err == nil && u.Scheme != "" {
				allowed[u.Scheme+"://"+u.Host] = true
			} else {
				allowed[o] = true
			}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With, Accept, X-API-Key")
			w.Header().Set("Access-Control-Max-Age", "86400")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
