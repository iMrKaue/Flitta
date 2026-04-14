package service

import (
	"flitta/internal/chatbot"
	"flitta/internal/model"
	"flitta/internal/repository"
	"flitta/internal/usecase"
	"flitta/internal/utils"
	"fmt"
	"strconv"
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

	if session.State == "" {
		session.State = model.StateIdle
	}

	if err != nil || session.Phone == "" {
		session = model.Session{
			Phone:    userID,
			ClientID: clientID,
			State:    model.StateIdle,
		}
	} else {
		session.ClientID = clientID
	}

	fmt.Println("ESTADO ATUAL:", session.State)

	if err != nil {
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
		return "Fluxo reiniciado ✅\nVocê pode *agendar*, ver *meus agendamentos*, *remarcar* ou *cancelar*."
	}

	intent := chatbot.DetectIntent(text)

	// =========================
	// 📅 LISTAR AGENDAMENTOS
	// =========================
	if intent == chatbot.IntentList {

		u := usecase.NewAppointmentUsecase()

		list, err := u.GetAppointmentsByCustomerPhone(clientID, userID)
		if err != nil || len(list) == 0 {
			return "Você não tem agendamentos ainda 😊"
		}

		session.State = model.StateChooseAppointment
		session.SelectedAppointmentID = 0
		repository.SaveSession(session)

		response := "📅 *Seus agendamentos:*\n\n"

		for i, a := range list {
			response += fmt.Sprintf("%d — 💇 %s\n📅 %s às %s\n\n", i+1, a.Service, a.Date, a.Time)
		}

		response += "\nDigite o número do agendamento."

		return response
	}

	// =========================
	// 🔢 ESCOLHER AGENDAMENTO
	// =========================
	if session.State == model.StateChooseAppointment {

		u := usecase.NewAppointmentUsecase()
		list, _ := u.GetAppointmentsByCustomerPhone(clientID, userID)

		index, err := strconv.Atoi(text)
		if err != nil || index < 1 || index > len(list) {
			return "Digite um número válido 😊"
		}

		selected := list[index-1]

		session.SelectedAppointmentID = selected.ID
		session.State = model.StateChooseAction
		repository.SaveSession(session)

		return "O que deseja fazer?\n\n1 - Remarcar\n2 - Cancelar"
	}

	// =========================
	// 🎯 ESCOLHER AÇÃO
	// =========================
	if session.State == model.StateChooseAction {

		if text == "1" {
			phone := utils.NormalizeCustomerPhone(userID)

			date, clientID, _, err := repository.GetAppointmentForReschedule(
				session.SelectedAppointmentID,
				phone,
			)

			if err != nil {
				return "Erro ao carregar horários 😢"
			}

			u := usecase.NewAppointmentUsecase()

			currentDate, _, _, _ := repository.GetAppointmentForReschedule(
				session.SelectedAppointmentID,
				utils.NormalizeCustomerPhone(userID),
			)

			currentTime := ""

			if currentDate == date {
				list, _ := u.GetAppointmentsByCustomerPhone(clientID, userID)
				for _, a := range list {
					if a.ID == session.SelectedAppointmentID {
						currentTime = a.Date
						break
					}
				}
			}

			slots, _ := u.GetAvailableSlots(clientID, date)

			var filtered []string
			for _, s := range slots {
				if s != currentTime {
					filtered = append(filtered, s)
				}
			}

			session.State = model.StateRescheduleTime
			repository.SaveSession(session)

			response := "🕒 Olha esses horários disponíveis pra você:\n\n"

			for i, s := range filtered {
				if i < 3 {
					response += "⭐ " + s + "\n"
				} else {
					response += "• " + s + "\n"
				}
			}

			response += "\nDigite o novo horário desejado 😊"

			return response
		}

		if text == "2" {
			session.State = model.StateCancelConfirm
			repository.SaveSession(session)
			return "Deseja cancelar? (sim/não)"
		}

		return "Escolha 1 para remarcar ou 2 para cancelar"
	}

	// =========================
	// 🔄 REMARCAR
	// =========================
	if session.State == model.StateRescheduleTime {

		u := usecase.NewAppointmentUsecase()

		if session.SelectedAppointmentID == 0 {
			repository.DeleteSession(userID)
			return "Erro ao identificar o agendamento. Tente novamente 🙏"
		}

		err := u.RescheduleAppointmentByID(
			session.SelectedAppointmentID,
			utils.NormalizeCustomerPhone(userID),
			text,
		)

		if err != nil {
			if strings.Contains(err.Error(), "ocupado") {
				return "Não foi possível remarcar 😢\nVerifique se o horário está disponível."
			}
			return "Não foi possível remarcar 😢"
		}
		repository.DeleteSession(userID)

		return "🔄 Agendamento atualizado com sucesso!"
	}

	// =========================
	// ❌ CANCELAR
	// =========================
	if session.State == model.StateCancelConfirm {

		if text == "sim" {

			u := usecase.NewAppointmentUsecase()

			err := u.CancelAppointment(
				session.SelectedAppointmentID,
				clientID,
				userID,
			)

			if err != nil {
				return "Erro ao cancelar 😢"
			}

			repository.DeleteSession(userID)

			return "❌ Agendamento cancelado com sucesso!"
		}

		repository.DeleteSession(userID)
		return "Cancelamento abortado 👍"
	}

	// =========================
	// 🤖 FLOW DE AGENDAMENTO
	// =========================
	if session.State == model.StateChooseAppointment ||
		session.State == model.StateChooseAction ||
		session.State == model.StateRescheduleTime ||
		session.State == model.StateCancelConfirm {

		return "Continue o fluxo digitando a opção 😊"
	}

	flow := chatbot.NewFlow()

	newSession, response := flow.Handle(session, text, intent)

	if err := repository.SaveSession(newSession); err != nil {
		fmt.Println("Erro ao salvar sessão:", err)
	}

	return response
}
