package service

import (
	"flitta/ai"
	"flitta/internal/chatbot"
	"flitta/internal/model"
	"flitta/internal/repository"
	"flitta/internal/utils"
	"fmt"
	"strings"
)

func ProcessMessage(userID, text string) string {
	println("1. entrou em ProcessMessage")

	clientID, err := GetClientByPhone(userID)
	println("2. voltou de GetClientByPhone")

	if err != nil {
		return "Empresa não encontrada. Entre em contato com o suporte."
	}

	text = strings.ToLower(strings.TrimSpace(text))
	println("4. texto normalizado:", text)

	// 🔹 sessão
	session, err := repository.GetSession(userID)
	println("5. voltou de GetSession")

	if err != nil || session.Phone == "" {
		session = model.Session{
			Phone:    userID,
			ClientID: clientID,
			State:    model.StateIdle,
		}
	} else {
		session.ClientID = clientID

		if session.State == "" {
			session.State = model.StateIdle
		}
	}

	println("6. antes do ParseMessage")
	aiData := ai.ParseMessage(text)
	println("7. depois do ParseMessage")

	if session.State == model.StateIdle {

		if aiData.Service != "" && session.Service == "" {
			session.Service = aiData.Service
		}

		if session.Date == "" {
			if aiData.Date != "" {
				session.Date = aiData.Date
				session.State = model.StateDate
			} else if parsedDate := utils.ParseDate(text); parsedDate != "" {
				session.Date = parsedDate
				session.State = model.StateDate
			}
			
		}

		if aiData.Time != "" && session.SuggestedTime == "" && session.Time == "" {
			session.SuggestedTime = aiData.Time
		}
	}

	// 🔹 reset
	if text == "sair" {
		repository.DeleteSession(userID)
		return "Fluxo reiniciado ✅\nVocê pode agendar, listar, remarcar ou cancelar."
	}

	println("8. antes do Handle. state=", session.State, "name=", session.Name, "service=", session.Service, "date=", session.Date, "time=", session.Time)

	flow := chatbot.NewFlow()
	newSession, response := flow.Handle(&session, text)
	println("9. depois do Handle:", response)

	if err := repository.SaveSession(*newSession); err != nil {
		println("10. erro ao salvar sessão")
		fmt.Println("Erro ao salvar sessão:", err)
	}

	println("11. fim do ProcessMessage")
	return response
}
