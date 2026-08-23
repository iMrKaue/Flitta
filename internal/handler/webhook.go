package handler

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flitta/internal/middleware"
	"flitta/internal/repository"
	"flitta/internal/service"
	"log"
	"net/http"
	"strings"
)

type Message struct {
	From   string `json:"from"`
	UserID string `json:"user_id"`
	To     string `json:"to"`
	Text   string `json:"text"`
}

type twimlResponse struct {
	XMLName xml.Name `xml:"Response"`
	Message string   `xml:"Message"`
}

func writeTwiMLResponse(
	w http.ResponseWriter,
	message string,
) error {
	payload, err := xml.Marshal(twimlResponse{
		Message: message,
	})
	if err != nil {
		return err
	}

	w.Header().Set(
		"Content-Type",
		"application/xml; charset=utf-8",
	)

	_, err = w.Write(payload)
	return err
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	from := strings.TrimSpace(r.FormValue("From"))
	to := strings.TrimSpace(r.FormValue("To"))
	text := strings.TrimSpace(r.FormValue("Body"))

	isTwilioRequest := from != "" || to != "" || text != ""

	// Fallback para testes locais/Postman.
	if !isTwilioRequest {
		var msg Message

		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		from = strings.TrimSpace(msg.From)

		// Compatibilidade temporária com o formato antigo.
		if from == "" {
			from = strings.TrimSpace(msg.UserID)
		}

		to = strings.TrimSpace(msg.To)
		text = strings.TrimSpace(msg.Text)
	}

	if from == "" || to == "" || text == "" {
		http.Error(
			w,
			"From, To e mensagem são obrigatórios",
			http.StatusBadRequest,
		)
		return
	}

	customerPhone := service.NormalizeWhatsAppNumber(from)
	channelAddress := service.NormalizeWhatsAppNumber(to)

	clientID, err := repository.GetClientIDByCommunicationChannel(
		"twilio",
		"whatsapp",
		channelAddress,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(
				w,
				"Canal de comunicação não configurado",
				http.StatusNotFound,
			)
			return
		}

		log.Printf(
			"erro ao resolver canal de comunicação %s: %v",
			channelAddress,
			err,
		)

		http.Error(
			w,
			"Erro interno",
			http.StatusInternalServerError,
		)
		return
	}

	response := service.ProcessMessage(
		clientID,
		customerPhone,
		text,
	)

	if isTwilioRequest {
		if err := writeTwiMLResponse(w, response); err != nil {
			log.Printf(
				"erro ao gerar resposta TwiML: %v",
				err,
			)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	json.NewEncoder(w).Encode(map[string]string{
		"response": response,
	})
}

func GetAppointmentsHandler(w http.ResponseWriter, r *http.Request) {

	clientID := r.Context().Value(middleware.ClientIDKey).(int)

	appointments, err := repository.GetAppointmentsByClient(clientID)
	if err != nil {
		http.Error(w, "Erro ao buscar", 500)
		return
	}

	json.NewEncoder(w).Encode(appointments)
}
