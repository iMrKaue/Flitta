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
				session.State = model.StateDate
			} else if parsedDate := utils.ParseDate(text); parsedDate != "" {
				session.Date = parsedDate
				session.State = model.StateDate
			}

		}

		if session.Time == "" && session.SuggestedTime == "" {
			if aiData.Time != "" {
				session.Time = aiData.Time
			} else if parsedHour := utils.NormalizeHour(text); parsedHour != "" {
				session.Time = parsedHour
			}
		}

		detectedPeriod := utils.ExtractPeriod(text)

		if session.Time != "" && session.Date != "" {
			session.State = model.StateTime
		}

		if session.Date != "" && session.Time == "" && detectedPeriod != "" {
			session.SuggestedTime = "__period__:" + detectedPeriod
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
		"sou",
		"eu sou",
	}

	for _, p := range patterns {
		if strings.Contains(lower, p) {
			idx := strings.Index(lower, p)
			if idx >= 0 {
				namePart := strings.TrimSpace(text[idx+len(p):])

				// corta se a frase continuar com outras intenções
				cutters := []string{
					" e quero",
					", quero",
					" quero",
					" amanhã",
					" amanha",
					" hoje",
					" segunda",
					" terça",
					" terca",
					" quarta",
					" quinta",
					" sexta",
					" sábado",
					" sabado",
					" domingo",
					" às ",
					" as ",
					" de manhã",
					" de manha",
					" à tarde",
					" a tarde",
					" à noite",
					" a noite",
				}

				lowerNamePart := strings.ToLower(namePart)
				cutPos := len(namePart)

				for _, cutter := range cutters {
					if pos := strings.Index(lowerNamePart, cutter); pos >= 0 && pos < cutPos {
						cutPos = pos
					}
				}

				namePart = strings.TrimSpace(namePart[:cutPos])

				words := strings.Fields(namePart)
				if len(words) == 0 {
					return ""
				}

				// aceita no máximo 3 palavras para nome
				if len(words) > 3 {
					words = words[:3]
				}

				return utils.Capitalize(strings.Join(words, " "))
			}
		}
	}

	return ""
}
