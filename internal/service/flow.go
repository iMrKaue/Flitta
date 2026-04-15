package service

import (
	"flitta/internal/chatbot"
	"flitta/internal/model"
	"flitta/internal/repository"
	"fmt"
	"strings"
)

func ProcessMessage(userID, text string) string {

	clientID, err := GetClientByPhone(userID)
	if err != nil {
		return "Empresa não encontrada. Entre em contato com o suporte."
	}

	text = strings.ToLower(strings.TrimSpace(text))

	// 🔹 sessão
	session, err := repository.GetSession(userID)

	if err != nil || session.Phone == "" {
		session = model.Session{
			Phone:    userID,
			ClientID: clientID,
			State:    model.StateIdle,
		}
	} else {
		session.ClientID = clientID
	}

	// 🔹 reset
	if text == "sair" {
		repository.DeleteSession(userID)
		return "Fluxo reiniciado ✅\nVocê pode agendar, listar, remarcar ou cancelar."
	}

	intent := chatbot.DetectIntent(text)

	flow := chatbot.NewFlow()

	newSession, response := flow.Handle(session, text, intent)

	if err := repository.SaveSession(newSession); err != nil {
		fmt.Println("Erro ao salvar sessão:", err)
	}

	return response
}
