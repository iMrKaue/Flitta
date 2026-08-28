package middleware

import (
	"flitta/internal/config"
	"net/http"
	"net/url"
	"strings"
)

func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))

		// Requisições sem Origin não são controladas por CORS.
		// Ex.: Twilio, curl, Postman e comunicação server-to-server.
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		if !isAllowedOrigin(origin) {
			http.Error(
				w,
				"Origin não permitida",
				http.StatusForbidden,
			)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, DELETE, OPTIONS",
		)
		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Authorization",
		)

		w.Header().Add("Vary", "Origin")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(origin string) bool {
	origin = strings.TrimSuffix(
		strings.TrimSpace(origin),
		"/",
	)

	if config.IsDevelopment() && isLocalDevelopmentOrigin(origin) {
		return true
	}

	for _, allowedOrigin := range config.GetCORSAllowedOrigins() {
		if origin == allowedOrigin {
			return true
		}
	}

	return false
}

func isLocalDevelopmentOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}

	if parsed.Scheme != "http" &&
		parsed.Scheme != "https" {
		return false
	}

	switch parsed.Hostname() {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}
