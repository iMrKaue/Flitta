package middleware

import (
	"context"
	"flitta/internal/service"
	"net/http"
	"strings"
)

type contextKey string

const ClientIDKey contextKey = "client_id"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(
			r.Header.Get("Authorization"),
		)

		if authHeader == "" {
			http.Error(
				w,
				"Token obrigatório",
				http.StatusUnauthorized,
			)
			return
		}

		parts := strings.Fields(authHeader)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			strings.TrimSpace(parts[1]) == "" {

			http.Error(
				w,
				"Token inválido",
				http.StatusUnauthorized,
			)
			return
		}

		clientID, err := service.ValidateToken(parts[1])
		if err != nil {
			http.Error(
				w,
				"Token inválido",
				http.StatusUnauthorized,
			)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			ClientIDKey,
			clientID,
		)

		next(w, r.WithContext(ctx))
	}
}
