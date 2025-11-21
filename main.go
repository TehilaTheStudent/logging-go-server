// CORS middleware that allows origins correctly and handles preflight requests.
// Behavior:
// - If ALLOWED_ORIGINS is set (comma-separated), only those origins are allowed.
// - If ALLOWED_ORIGINS is empty, treat as "allow all origins".
//   - If ALLOW_CREDENTIALS=="true", echo the request Origin (required to allow credentials).
//   - Otherwise, use Access-Control-Allow-Origin: "*" (no credentials).
func corsMiddleware(next http.Handler) http.Handler {
	allowedCSV := os.Getenv("ALLOWED_ORIGINS") // e.g. "https://example.com,https://foo.bar"
	allowCredentials := os.Getenv("ALLOW_CREDENTIALS") == "true"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// If an Origin header is present, treat as a browser CORS request.
		if origin != "" {
			if allowedCSV == "" {
				// Allow all origins logically.
				if allowCredentials {
					// Must echo the Origin when credentials are allowed.
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Vary", "Origin")
				} else {
					// No credentials: safe to allow all origins with "*"
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}
			} else {
				// We have an allowlist: only allow if origin is in the list.
				if originAllowed(origin, allowedCSV) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					if allowCredentials {
						w.Header().Set("Access-Control-Allow-Credentials", "true")
					}
				}
				// If not allowed, do not set Access-Control-Allow-* headers (browser will block).
			}
		} else {
			// No Origin header (likely non-browser request). Optionally set wildcard for non-browser clients.
			// We'll set wildcard only if there's no allowlist.
			if allowedCSV == "" && !allowCredentials {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
		}

		// Allowed methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		// Expose these headers to browser JS
		w.Header().Set("Access-Control-Expose-Headers", "X-Served-By, X-Timestamp, Content-Length")

		// Echo requested headers if present, otherwise provide a sensible default
		if reqHeaders := r.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
		} else {
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Api-Key, Accept")
		}

		// For preflight requests, respond with 204 No Content and return
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
