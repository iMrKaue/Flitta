package middleware

import (
	"context"
	"flitta/internal/config"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

type contextKey string

const ClientIDKey contextKey = "client_id"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Token obrigatório", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return config.GetJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Erro no token", http.StatusUnauthorized)
		}

		clientID := int(claims["client_id"].(float64))

		ctx := context.WithValue(r.Context(), ClientIDKey, clientID)

		next(w, r.WithContext(ctx))
	}
}
