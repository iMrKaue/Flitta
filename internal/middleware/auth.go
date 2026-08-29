package middleware

import (
	"context"
	"errors"
	"flitta/internal/service"
	"log"
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

		if err := service.ValidateClientAccess(clientID); err != nil {
			if errors.Is(
				err,
				service.ErrClientAccessBlocked,
			) {
				http.Error(
					w,
					"Acesso ao estabelecimento suspenso ou expirado",
					http.StatusForbidden,
				)
				return
			}

			log.Printf(
				"erro ao validar acesso do estabelecimento %d: %v",
				clientID,
				err,
			)

			http.Error(
				w,
				"Erro interno",
				http.StatusInternalServerError,
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
