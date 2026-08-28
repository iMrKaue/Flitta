package middleware

import (
	"flitta/internal/config"
	"log"
	"net/http"
	"strings"

	twilioclient "github.com/twilio/twilio-go/client"
)

func ValidateTwilioWebhook(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next(w, r)
			return
		}

		if config.IsDevelopment() {
			next(w, r)
			return
		}

		contentType := strings.ToLower(
			strings.TrimSpace(r.Header.Get("Content-Type")),
		)

		if !strings.HasPrefix(
			contentType,
			"application/x-www-form-urlencoded",
		) {
			http.Error(
				w,
				"Tipo de conteúdo não suportado",
				http.StatusUnsupportedMediaType,
			)
			return
		}

		signature := strings.TrimSpace(
			r.Header.Get("X-Twilio-Signature"),
		)

		if signature == "" {
			http.Error(
				w,
				"Requisição não autorizada",
				http.StatusForbidden,
			)
			return
		}

		authToken := config.GetTwilioAuthToken()
		webhookURL := config.GetTwilioWebhookURL()

		if authToken == "" || webhookURL == "" {
			log.Println(
				"configuração Twilio incompleta para validação do webhook",
			)

			http.Error(
				w,
				"Erro interno",
				http.StatusInternalServerError,
			)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(
				w,
				"Requisição inválida",
				http.StatusBadRequest,
			)
			return
		}

		params := make(map[string]string, len(r.PostForm))

		for key, values := range r.PostForm {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}

		validator := twilioclient.NewRequestValidator(authToken)

		if !validator.Validate(
			webhookURL,
			params,
			signature,
		) {
			http.Error(
				w,
				"Requisição não autorizada",
				http.StatusForbidden,
			)
			return
		}

		next(w, r)
	}
}
