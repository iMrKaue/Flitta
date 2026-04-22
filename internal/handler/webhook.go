package handler

import (
	"encoding/json"
	"flitta/internal/middleware"
	"flitta/internal/repository"
	"flitta/internal/service"
	"net/http"
)

type Message struct {
	UserID string `json:"user_id"`
	Text   string `json:"text"`
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {

	// tenta pegar do Twilio
	userID := r.FormValue("From")
	text := r.FormValue("Body")

	// se não veio do twilio -> tenta
	if userID == "" || text == "" {
		var msg struct {
			UserID string `json:"user_id"`
			Text   string `json:"text"`
		}

		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		userID = msg.UserID
		text = msg.Text
	}

	response := service.ProcessMessage(userID, text)

	// resposta depende da origem
	if r.FormValue("From") != "" {
		// Twilio -> XML
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`
		<Response>
			<Message>` + response + `</Message>
		</Response>
		`))
	} else {
		// Postman -> JSON
		json.NewEncoder(w).Encode(map[string]string{
			"response": response,
		})
	}
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
