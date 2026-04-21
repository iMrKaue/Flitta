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
		enrichSessionFromNaturalText(&session, clientID, text, aiData)

		if session.Time != "" && session.Date != "" {
			session.State = model.StateTime
		}

		if session.Date != "" && session.Time == "" && strings.HasPrefix(session.SuggestedTime, "__period__:") {
			session.State = model.StateTime
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

func extractServiceFromText(text string, clientID int) string {
	text = strings.ToLower(strings.TrimSpace(text))
	services := repository.GetServices(clientID)

	for _, s := range services {
		serviceName := strings.ToLower(strings.TrimSpace(s.Name))
		if serviceName != "" && strings.Contains(text, serviceName) {
			return s.Name
		}
	}

	return ""
}

func extractNameFromText(text string) string {
	text = strings.TrimSpace(text)
	lower := strings.ToLower(text)

	patterns := []string{
		"meu nome é ",
		"meu nome e ",
		"eu sou ",
		"sou ",
	}

	for _, p := range patterns {
		if idx := strings.Index(lower, p); idx >= 0 {
			namePart := strings.TrimSpace(text[idx+len(p):])
			namePart = cutNameAtIntentMarkers(namePart)
			if namePart != "" {
				return namePart
			}
		}
	}

	return ""
}

func cutNameAtIntentMarkers(namePart string) string {
	cutters := []string{
		" e quero",
		", quero",
		" quero",
		" e marcar",
		", marcar",
		" marcar",
		" amanhã",
		" amanha",
		" e amanhã",
		" e amanha",
		" hoje",
		" e hoje",
		" segunda",
		" e segunda",
		" terça",
		" terca",
		" e terça",
		" e terca",
		" quarta",
		" e quarta",
		" quinta",
		" e quinta",
		" sexta",
		" e sexta",
		" sábado",
		" sabado",
		" e sábado",
		" e sabado",
		" domingo",
		" e domingo",
		" às ",
		" as ",
		" e às ",
		" e as ",
		" de manhã",
		" de manha",
		" à tarde",
		" a tarde",
		" no fim da tarde",
		" fim da tarde",
		" à noite",
		" a noite",
	}

	lower := strings.ToLower(namePart)
	cutPos := len(namePart)

	for _, cutter := range cutters {
		if pos := strings.Index(lower, cutter); pos >= 0 && pos < cutPos {
			cutPos = pos
		}
	}

	namePart = strings.TrimSpace(namePart[:cutPos])

	words := strings.Fields(namePart)
	if len(words) == 0 {
		return ""
	}
	if len(words) > 3 {
		words = words[:3]
	}

	return utils.Capitalize(strings.Join(words, " "))
}

func enrichSessionFromNaturalText(session *model.Session, clientID int, text string, aiData ai.AIData) {
	if session.Name == "" {
		if detectedName := extractNameFromText(text); detectedName != "" {
			session.Name = detectedName
		}
	}

	if session.Service == "" {
		if aiData.Service != "" {
			session.Service = aiData.Service
		} else if detectedService := extractServiceFromText(text, clientID); detectedService != "" {
			session.Service = detectedService
		}
	}

	if session.Date == "" {
		if aiData.Date != "" {
			session.Date = aiData.Date
		} else if parsedDate := utils.ParseDate(text); parsedDate != "" {
			session.Date = parsedDate
		}
	}

	if session.Time == "" && session.SuggestedTime == "" {
		if aiData.Time != "" {
			session.Time = aiData.Time
		} else if parsedHour := utils.NormalizeHour(text); parsedHour != "" {
			session.Time = parsedHour
		}
	}

	if session.Date != "" && session.Time == "" {
		if detectedPeriod := utils.ExtractPeriod(text); detectedPeriod != "" {
			session.SuggestedTime = "__period__:" + detectedPeriod
		}
	}
}
