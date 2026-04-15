package chatbot

import (
	"flitta/internal/model"
	"flitta/internal/repository"
	"flitta/internal/usecase"
	"flitta/internal/utils"
	"fmt"
	"strconv"
	"strings"
)

type Flow struct {
	appointments *usecase.AppointmentUsecase
}

func NewFlow() *Flow {
	return &Flow{appointments: usecase.NewAppointmentUsecase()}
}

func (f *Flow) Handle(session model.Session, text string, intent Intent) (model.Session, string) {

	switch session.State {

	case model.StateIdle:
		return f.handleStart(session, intent)

	case model.StateName:
		return f.handleName(session, text)

	case model.StateService:
		return f.handleService(session, text)

	case model.StateDate:
		return f.handleDate(session, text)

	case model.StateTime:
		return f.handleTime(session, text)

	case model.StateConfirm:
		return f.handleConfirm(session, text)

	case model.StateChooseAppointment:
		return f.handleChooseAppointment(session, text)

	case model.StateChooseAction:
		return f.handleChooseAction(session, text)

	case model.StateRescheduleTime:
		return f.handleRescheduleTime(session, text)

	case model.StateCancelConfirm:
		return f.handleCancel(session, text)
	default:
		session.State = model.StateIdle // ✅ corrigido
		return session, "Não entendi 😅\nVamos começar novamente."
	}
}

func (f *Flow) handleStart(session model.Session, intent Intent) (model.Session, string) {

	switch intent {

	case IntentGreeting:
		session.State = model.StateName // ✅ corrigido
		return session, "Olá! Seja bem-vindo(a) ao salão 💇‍♀️\nQual seu nome?"

	case IntentSchedule:
		session.State = model.StateName // ✅ corrigido
		return session, "👋 Olá! Vamos agendar seu horário 😊\n\nQual seu nome?"

	case IntentList:
		session.State = model.StateChooseAppointment
		return session, f.listAppointments(session)

	default:
		session.State = model.StateName // ✅ corrigido
		return session, "Vamos começar 😊\nQual seu nome?"
	}
}

func (f *Flow) handleName(session model.Session, text string) (model.Session, string) {

	if len(text) < 3 ||
		strings.Contains(text, "quero") ||
		strings.Contains(text, "agendar") ||
		strings.Contains(text, "horario") {

		return session, "Pode me dizer seu nome? 😊"
	}

	session.Name = utils.Capitalize(text)
	session.State = model.StateService // ✅ corrigido

	services := repository.GetServices(session.ClientID)

	response := "Escolha o serviço:\n"

	for i, a := range services {
		response += fmt.Sprintf("%d - %s\n", i+1, a.Name)
	}

	return session, response
}

func (f *Flow) handleService(session model.Session, text string) (model.Session, string) {

	services := repository.GetServices(session.ClientID)

	index, err := strconv.Atoi(text)
	if err == nil && index >= 1 && index <= len(services) {
		session.Service = services[index-1].Name
		session.State = model.StateDate // ✅ corrigido
		return session, "Qual dia você deseja?"
	}

	for _, s := range services {
		if strings.Contains(strings.ToLower(text), strings.ToLower(s.Name)) {
			session.Service = s.Name
			session.State = model.StateDate // ✅ corrigido
			return session, "Qual dia você deseja?"
		}
	}

	return session, "Não entendi 😅\nEscolha um dos serviços listados."
}

func (f *Flow) handleDate(session model.Session, text string) (model.Session, string) {

	session.Date = utils.ParseDate(text)
	session.State = model.StateTime // ✅ corrigido

	slots, _ := f.appointments.GetAvailableSlots(session.ClientID, session.Date)

	response := "🕒 Olha os horários disponíveis para " + session.Date + ":\n\n"

	for i, s := range slots {
		if i < 3 {
			response += "⭐ " + s + "\n"
		} else {
			response += "• " + s + "\n"
		}
	}

	response += "\nDigite ou escolha um horário 😊"

	return session, response
}

func (f *Flow) handleTime(session model.Session, text string) (model.Session, string) {

	text = strings.TrimSpace(strings.ToLower(text))

	text = strings.ReplaceAll(text, "h", "")
	text = strings.ReplaceAll(text, ":", "")
	text = strings.ReplaceAll(text, "às", "")
	text = strings.TrimSpace(text)

	if len(text) == 2 {
		text = text + ":00"
	}
	if len(text) == 3 {
		text = text[:1] + ":" + text[1:]
	}

	slots, _ := f.appointments.GetAvailableSlots(session.ClientID, session.Date)

	valid := false

	for _, s := range slots {
		if strings.TrimSpace(s) == text {
			valid = true
			break
		}
	}

	// 🔴 SE NÃO FOR VÁLIDO
	if !valid {

		closest := ""
		if len(slots) > 0 {
			closest = slots[0]
		}

		return session,
			fmt.Sprintf(
				"😅 Esse horário não está disponível.\n\nQue tal %s?\n\nOu escolha outro abaixo 👇",
				closest,
			)
	}

	// 🟢 SE FOR VÁLIDO (continua fluxo)
	session.Time = text
	session.State = model.StateConfirm

	return session,
		"📋 Confirmação do seu horário:\n\n" +
			"👤 " + session.Name + "\n" +
			"💇 " + session.Service + "\n" +
			"📅 " + session.Date + "\n" +
			"🕒 " + session.Time + "\n\nConfirmar? (sim/não)"
}

func (f *Flow) handleConfirm(session model.Session, text string) (model.Session, string) {

	if text == "sim" {

		err := f.appointments.CreateAppointment(
			session.ClientID,
			session.Phone,
			session.Name,
			session.Service,
			session.Date,
			session.Time,
		)

		if err != nil {
			return session, "Esse horário acabou de ser ocupado 😢\nEscolha outro horário."
		}

		session.State = model.StateIdle // ✅ corrigido

		return session,
			fmt.Sprintf(
				"✅ Agendamento confirmado!\n\n👤 %s\n💇 %s\n📅 %s\n🕒 %s\n\nTe esperamos no horário combinado 😊",
				session.Name, session.Service, session.Date, session.Time,
			)
	}

	session.State = model.StateIdle // ✅ corrigido
	return session, "Agendamento cancelado. Podemos começar novamente 😊"
}

func (f *Flow) listAppointments(session model.Session) string {

	list, err := f.appointments.GetAppointmentsByCustomerPhone(
		session.ClientID,
		session.Phone,
	)

	if err != nil || len(list) == 0 {
		return "Você não tem agendamentos ainda 😊"
	}

	response := "📅 Seus agendamentos:\n\n"

	for i, a := range list {
		response += fmt.Sprintf("%d - 💇 %s\n📅 %s às %s\n\n",
			i+1, a.Service, a.Date, a.Time)
	}

	response += "Digite o número do agendamento."

	return response
}

func (f *Flow) handleChooseAppointment(session model.Session, text string) (model.Session, string) {

	list, _ := f.appointments.GetAppointmentsByCustomerPhone(
		session.ClientID,
		session.Phone,
	)

	index, err := strconv.Atoi(text)
	if err != nil || index < 1 || index > len(list) {
		return session, "Digite um número válido 😊"
	}

	selected := list[index-1]

	session.SelectedAppointmentID = selected.ID
	session.State = model.StateChooseAction

	return session, "O que deseja fazer?\n\n1 - Remarcar\n2 - Cancelar"
}

func (f *Flow) handleChooseAction(session model.Session, text string) (model.Session, string) {

	if text == "1" {
		session.State = model.StateRescheduleTime
		return session, "Digite o novo horário desejado 😊"
	}

	if text == "2" {
		session.State = model.StateCancelConfirm
		return session, "Deseja cancelar? (sim/não)"
	}

	return session, "Escolha 1 para remarcar ou 2 para cancelar"
}

func (f *Flow) handleRescheduleTime(session model.Session, text string) (model.Session, string) {

	if session.SelectedAppointmentID == 0 {
		session.State = model.StateIdle
		return session, "Erro ao identificar agendamento. Tente novamente 🙏"
	}

	err := f.appointments.RescheduleAppointmentByID(
		session.SelectedAppointmentID,
		session.Phone,
		text,
	)

	if err != nil {
		return session, "Não foi possível remarcar 😢\nTente outro horário."
	}

	session.State = model.StateIdle

	return session, "🔄 Agendamento atualizado com sucesso!"
}

func (f *Flow) handleCancel(session model.Session, text string) (model.Session, string) {

	if text == "sim" {

		err := f.appointments.CancelAppointment(
			session.SelectedAppointmentID,
			session.ClientID,
			session.Phone,
		)

		if err != nil {
			return session, "Erro ao cancelar 😢"
		}

		session.State = model.StateIdle
		return session, "❌ Agendamento cancelado com sucesso!"
	}

	session.State = model.StateIdle
	return session, "Cancelamento abortado 👍"
}