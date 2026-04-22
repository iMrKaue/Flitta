package chatbot

import (
	"flitta/internal/model"
	"flitta/internal/repository"
	"flitta/internal/usecase"
	"flitta/internal/utils"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Flow struct {
	appointments *usecase.AppointmentUsecase
}

func NewFlow() *Flow {
	return &Flow{appointments: usecase.NewAppointmentUsecase()}
}

func (f *Flow) Handle(session *model.Session, text string) (*model.Session, string) {

	switch session.State {

	case model.StateIdle:
		return f.handleStart(session, text)

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
		session.State = model.StateIdle
		return session, "Não entendi 😅\nVamos começar novamente."
	}
}

func (f *Flow) handleStart(session *model.Session, text string) (*model.Session, string) {

	if session.Name != "" || session.Service != "" || session.Date != "" || session.Time != "" {

		if session.Service == "" {
			session.State = model.StateService

			services := repository.GetServices(session.ClientID)
			response := "Qual serviço você deseja?\n"
			for i, s := range services {
				response += fmt.Sprintf("%d - %s\n", i+1, s.Name)
			}

			return session, response
		}

		if session.Date == "" {
			session.State = model.StateDate
			return session, "Qual dia você deseja?"
		}

		if session.Time == "" {
			session.State = model.StateTime

			if strings.HasPrefix(session.SuggestedTime, "__period__:") {
				return f.handleTime(session, text)
			}

			if utils.NormalizeHour(text) != "" {
				return f.handleTime(session, text)
			}

			return f.presentAvailableSlots(session)
		}

		if session.Name == "" {
			session.State = model.StateName
			return session, "Qual é o seu nome?"
		}

		session.State = model.StateConfirm
		return session,
			"📋 Confirmação do seu horário:\n\n" +
				"👤 " + session.Name + "\n" +
				"💇 " + session.Service + "\n" +
				"📅 " + session.Date + "\n" +
				"🕒 " + session.Time + "\n\nConfirmar? (sim/não)"
	}

	intent := DetectIntent(text)

	switch intent {

	case IntentGreeting:
		session.State = model.StateName // ✅ corrigido
		return session, "Olá! Seja bem-vindo(a) 😊\nQual é o seu nome?"

	case IntentSchedule:
		session.State = model.StateName // ✅ corrigido
		return session, "Perfeito 😊 Vamos agendar seu horário.\nQual é o seu nome?"

	case IntentList:
		session.State = model.StateChooseAppointment
		return session, f.listAppointments(session)

	default:
		return session, "Não entendi direito 😊\nVocê quer agendar, remarcar, cancelar ou ver seus agendamentos?"
	}
}

func (f *Flow) handleName(session *model.Session, text string) (*model.Session, string) {

	if len(text) < 3 ||
		strings.Contains(text, "quero") ||
		strings.Contains(text, "agendar") ||
		strings.Contains(text, "horario") {

		return session, "Pode me dizer seu nome? 😊"
	}

	session.Name = utils.Capitalize(text)

	if session.Service == "" {
		session.State = model.StateService

		services := repository.GetServices(session.ClientID)

		response := "Perfeito 😊\nEscolha um serviço abaixo:\n"
		for i, a := range services {
			response += fmt.Sprintf("%d - %s\n", i+1, a.Name)
		}

		return session, response
	}

	if session.Date == "" {
		session.State = model.StateDate
		return session, "Qual dia você deseja? 😊"
	}

	if session.Time == "" {
		return f.presentAvailableSlots(session)
	}

	session.State = model.StateConfirm
	return session,
		"📋 Confirmação do seu horário:\n\n" +
			"👤 " + session.Name + "\n" +
			"💇 " + session.Service + "\n" +
			"📅 " + session.Date + "\n" +
			"🕒 " + session.Time + "\n\nConfirmar? (sim/não)"

}

func (f *Flow) handleService(session *model.Session, text string) (*model.Session, string) {
	services := repository.GetServices(session.ClientID)

	index, err := strconv.Atoi(text)
	if err == nil && index >= 1 && index <= len(services) {
		session.Service = services[index-1].Name

		if session.Date == "" {
			session.State = model.StateDate
			return session, "Qual dia você deseja?"
		}

		if session.Time == "" {
			return f.presentAvailableSlots(session)
		}

		session.State = model.StateConfirm
		return session,
			"📋 Confirmação do seu horário:\n\n" +
				"👤 " + session.Name + "\n" +
				"💇 " + session.Service + "\n" +
				"📅 " + session.Date + "\n" +
				"🕒 " + session.Time + "\n\nConfirmar? (sim/não)"
	}

	for _, s := range services {
		if strings.Contains(strings.ToLower(text), strings.ToLower(s.Name)) {
			session.Service = s.Name

			if session.Date == "" {
				session.State = model.StateDate
				return session, "Qual dia você deseja?"
			}

			if session.Time == "" {
				return f.presentAvailableSlots(session)
			}

			session.State = model.StateConfirm
			return session,
				"📋 Confirmação do seu horário:\n\n" +
					"👤 " + session.Name + "\n" +
					"💇 " + session.Service + "\n" +
					"📅 " + session.Date + "\n" +
					"🕒 " + session.Time + "\n\nConfirmar? (sim/não)"
		}
	}

	response := "Não entendi qual serviço você deseja 😅\nEscolha um dos serviços abaixo:\n"
	for i, s := range services {
		response += fmt.Sprintf("%d - %s\n", i+1, s.Name)
	}

	return session, response
}

func (f *Flow) handleDate(session *model.Session, text string) (*model.Session, string) {
	parsed := utils.ParseDate(text)

	if parsed != "" {
		session.Date = parsed
	}

	if session.Date == "" {
		return session, "Não entendi o dia 😅\nVocê pode me dizer, por exemplo: hoje, amanhã, sexta-feira."
	}

	if session.Time != "" {
		if session.Name == "" {
			session.State = model.StateName
			return session, "Perfeito 😊 Qual é o seu nome?"
		}

		if session.Service == "" {
			session.State = model.StateService

			services := repository.GetServices(session.ClientID)
			response := "Perfeito 😊\nEscolha um serviço abaixo:\n"
			for i, s := range services {
				response += fmt.Sprintf("%d - %s\n", i+1, s.Name)
			}

			return session, response
		}

		session.State = model.StateConfirm
		return session,
			"📋 Confirmação do seu horário:\n\n" +
				"👤 " + session.Name + "\n" +
				"💇 " + session.Service + "\n" +
				"📅 " + session.Date + "\n" +
				"🕒 " + session.Time + "\n\nConfirmar? (sim/não)"
	}

	return f.presentAvailableSlots(session)
}

// presentAvailableSlots assume session.Date preenchido; coloca em StateTime e lista horários + sugestão.
func (f *Flow) presentAvailableSlots(session *model.Session) (*model.Session, string) {

	session.State = model.StateTime

	slots, _ := f.appointments.GetAvailableSlots(session.ClientID, session.Date)

	if len(slots) == 0 {
		session.State = model.StateDate
		session.SuggestedTime = ""
		return session, "😢 Não há horários livres nessa data.\nQual outro dia você prefere?"
	}

	suggestedOK := false
	for _, s := range slots {
		if session.SuggestedTime != "" && s == session.SuggestedTime {
			suggestedOK = true
			break
		}
	}
	if !suggestedOK {
		session.SuggestedTime = slots[0]
	}

	response := "🕒 Olha os horários disponíveis para " + session.Date + ":\n\n"
	response += "💡 Tenho um horário às " + session.SuggestedTime + " (recomendado), quer esse? 😊\n\n"

	shown := 0
	for _, s := range slots {
		if s == session.SuggestedTime {
			continue
		}
		if shown < 3 {
			response += "⭐ " + s + "\n"
		} else {
			response += "• " + s + "\n"
		}
		shown++
	}

	response += "\nDigite ou escolha um horário 😊"

	return session, response
}

func (f *Flow) presentAvailableSlotsExcluding(session *model.Session, excluded string) (*model.Session, string) {
	slots, _ := f.appointments.GetAvailableSlots(session.ClientID, session.Date)

	var filtered []string
	for _, slot := range slots {
		if strings.TrimSpace(slot) != strings.TrimSpace(excluded) {
			filtered = append(filtered, slot)
		}
	}

	if len(filtered) == 0 {
		return session, "Não encontrei outros horários disponíveis nesse dia 😢"
	}

	session.SuggestedTime = filtered[0]

	response := fmt.Sprintf(
		"🕒 Tudo bem 😊 Tenho outros horários disponíveis para %s:\n\n💡 Posso te sugerir %s. Quer esse?\n\n",
		session.Date,
		filtered[0],
	)

	limit := len(filtered)
	if limit > 8 {
		limit = 8
	}

	for i := 0; i < limit; i++ {
		if i < 3 {
			response += "⭐ " + filtered[i] + "\n"
		} else {
			response += "• " + filtered[i] + "\n"
		}
	}

	response += "\nDigite ou escolha um horário 😊"
	return session, response
}

func (f *Flow) handleTime(session *model.Session, text string) (*model.Session, string) {
	intent := AnalyzeIntent(text)

	if strings.HasPrefix(session.SuggestedTime, "__period__:") && intent.Period == "" {
		intent.Period = strings.TrimPrefix(session.SuggestedTime, "__period__:")
		session.SuggestedTime = ""
	}

	if session.SuggestedTime != "" && intent.IsPositive {
		text = session.SuggestedTime
	}

	if intent.AsksOtherDay {
		session.Date = ""
		session.Time = ""
		session.SuggestedTime = ""
		session.State = model.StateDate
		return session, "Certo 😊 Qual outro dia você deseja?"
	}

	if intent.AsksEarlier {
		slots, _ := f.appointments.GetAvailableSlots(session.ClientID, session.Date)

		base := session.SuggestedTime
		if base == "" {
			base = session.Time
		}

		if base != "" {
			earlier := findEarlierSlot(slots, base)
			if earlier != "" {
				session.SuggestedTime = earlier
				return session, fmt.Sprintf("Perfeito 😊 Tenho um horário mais cedo às %s. Quer esse?", earlier)
			}
		}

		return session, "Não tenho um horário mais cedo disponível 😊"
	}

	if intent.AsksLater {
		slots, _ := f.appointments.GetAvailableSlots(session.ClientID, session.Date)

		base := session.SuggestedTime
		if base == "" {
			base = session.Time
		}

		if base != "" {
			later := findLaterSlot(slots, base)
			if later != "" {
				session.SuggestedTime = later
				return session, fmt.Sprintf("Perfeito 😊 Tenho um horário mais tarde às %s. Quer esse?", later)
			}
		}

		return session, "Não tenho um horário mais tarde disponível 😊"
	}

	if intent.Period != "" {
		slots, _ := f.appointments.GetAvailableSlots(session.ClientID, session.Date)
		filtered := filterSlotsByPeriod(slots, intent.Period)

		if len(filtered) == 0 {
			if intent.Period == "noite" {
				return session, "Não encontrei horários à noite nesse dia 😊 O salão atende até o fim da tarde."
			}
			return session, fmt.Sprintf("Não encontrei horários %s nesse dia 😊", periodLabel(intent.Period))
		}

		session.SuggestedTime = filtered[0]

		response := fmt.Sprintf(
			"🕒 Tenho estes horários %s para %s:\n\n💡 Posso te sugerir %s. Quer esse?\n\n",
			periodLabel(intent.Period),
			session.Date,
			filtered[0],
		)

		limit := len(filtered)
		if limit > 8 {
			limit = 8
		}

		for i := 0; i < limit; i++ {
			if i < 3 {
				response += "⭐ " + filtered[i] + "\n"
			} else {
				response += "• " + filtered[i] + "\n"
			}
		}

		response += "\nDigite ou escolha um horário 😊"
		return session, response
	}

	text = strings.TrimSpace(strings.ToLower(text))

	if normalized := utils.NormalizeHour(text); normalized != "" {
		text = normalized
	} else {
		text = strings.ReplaceAll(text, "às", "")
		text = strings.ReplaceAll(text, "h", "")
		text = strings.TrimSpace(text)

		if len(text) == 2 {
			text = text + ":00"
		}

		if len(text) == 3 {
			text = text[:1] + ":" + text[1:]
		}
	}

	slots, _ := f.appointments.GetAvailableSlots(session.ClientID, session.Date)

	valid := false
	for _, s := range slots {
		if strings.TrimSpace(s) == strings.TrimSpace(text) {
			valid = true
			break
		}
	}

	if !valid {
		if len(slots) == 0 {
			return session, "Não encontrei horários disponíveis para esse dia 😢\nVocê pode tentar outro dia."
		}

		closest := slots[0]
		session.SuggestedTime = closest

		response := fmt.Sprintf(
			"😅 Esse horário não está disponível.\n\nQue tal %s?\n\nOu escolha outro abaixo 👇\n\n",
			closest,
		)

		limit := len(slots)
		if limit > 8 {
			limit = 8
		}

		for i := 0; i < limit; i++ {
			if i < 3 {
				response += "⭐ " + slots[i] + "\n"
			} else {
				response += "• " + slots[i] + "\n"
			}
		}

		response += "\nDigite ou escolha um horário 😊"
		return session, response
	}

	session.Time = text
	session.SuggestedTime = ""

	if session.Name == "" {
		session.State = model.StateName
		return session, "Perfeito 😊 Qual é o seu nome?"
	}

	if session.Service == "" {
		session.State = model.StateService

		services := repository.GetServices(session.ClientID)
		response := "Qual serviço você deseja?\n"
		for i, s := range services {
			response += fmt.Sprintf("%d - %s\n", i+1, s.Name)
		}

		return session, response
	}

	session.State = model.StateConfirm
	return session,
		"📋 Confirmação do seu horário:\n\n" +
			"👤 " + session.Name + "\n" +
			"💇 " + session.Service + "\n" +
			"📅 " + session.Date + "\n" +
			"🕒 " + session.Time + "\n\nConfirmar? (sim/não)"
}

func (f *Flow) handleConfirm(session *model.Session, text string) (*model.Session, string) {
	intent := AnalyzeIntent(text)

	if intent.IsPositive {
		name := session.Name
		service := session.Service
		date := session.Date
		timeValue := session.Time

		err := f.appointments.CreateAppointment(
			session.ClientID,
			session.Phone,
			name,
			service,
			date,
			timeValue,
		)

		if err != nil {
			return session, "Esse horário acabou de ser ocupado 😢\nEscolha outro horário."
		}

		session.State = model.StateIdle
		session.Name = ""
		session.Service = ""
		session.Date = ""
		session.Time = ""
		session.SuggestedTime = ""

		return session,
			fmt.Sprintf(
				"✅ Agendamento confirmado!\n\n👤 %s\n💇 %s\n📅 %s\n🕒 %s\n\nTe esperamos no horário combinado 😊",
				name, service, date, timeValue,
			)
	}

	if intent.IsNegative {
		rejectedTime := session.Time

		session.State = model.StateTime
		session.Time = ""
		session.SuggestedTime = ""

		return f.presentAvailableSlotsExcluding(session, rejectedTime)
	}

	return session, "Não entendi sua resposta 😊\nResponda com sim para confirmar ou não para escolher outro horário."
}

func (f *Flow) listAppointments(session *model.Session) string {

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

func (f *Flow) handleChooseAppointment(session *model.Session, text string) (*model.Session, string) {

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

func (f *Flow) handleChooseAction(session *model.Session, text string) (*model.Session, string) {

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

func (f *Flow) handleRescheduleTime(session *model.Session, text string) (*model.Session, string) {

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

func (f *Flow) handleCancel(session *model.Session, text string) (*model.Session, string) {

	intent := AnalyzeIntent(text)

	if intent.IsPositive {

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
	return session, "Não entendi sua resposta 😊\nResponda com sim para cancelar ou não para continuar com seu agendamento."
}

var (
	reAffirmPodeSer = regexp.MustCompile(`pode\s+ser`)
	reAffirmNegacao = regexp.MustCompile(`\b(não|nao)\b`)
)

// affirmsSuggestedSlot detecta se o usuário aceitou o horário recomendado (evita "pode" sozinho; aceita "pode  ser", "pode ser!", etc.).
func affirmsSuggestedSlot(text string) bool {
	t := strings.ToLower(strings.TrimSpace(text))
	if t == "" {
		return false
	}
	if reAffirmNegacao.MatchString(t) {
		return false
	}
	t = strings.Trim(t, ".,!?;…")
	if t == "sim" || t == "s" || t == "ok" || t == "okay" {
		return true
	}
	if strings.Contains(t, " sim") || strings.HasPrefix(t, "sim ") {
		return true
	}
	if reAffirmPodeSer.MatchString(t) {
		return true
	}
	return strings.Contains(t, "fechado") ||
		strings.Contains(t, "confirmo") ||
		strings.Contains(t, "claro") ||
		strings.Contains(t, "isso mesmo") ||
		strings.Contains(t, "esse mesmo") ||
		t == "isso" ||
		strings.Contains(t, "beleza") ||
		strings.Contains(t, "combinado")
}

func findEarlierSlot(slots []string, current string) string {
	for i, slot := range slots {
		if strings.TrimSpace(slot) == strings.TrimSpace(current) {
			if i > 0 {
				return slots[i-1]
			}
			return ""
		}
	}
	return ""
}

func findLaterSlot(slots []string, current string) string {
	for i, slot := range slots {
		if strings.TrimSpace(slot) == strings.TrimSpace(current) {
			if i < len(slots)-1 {
				return slots[i+1]
			}
			return ""
		}
	}
	return ""
}

func filterSlotsByPeriod(slots []string, period string) []string {
	var filtered []string

	for _, slot := range slots {
		switch period {
		case "manha":
			if slot >= "09:00" && slot < "12:00" {
				filtered = append(filtered, slot)
			}
		case "tarde":
			if slot >= "12:00" && slot < "16:00" {
				filtered = append(filtered, slot)
			}
		case "fim_tarde":
			if slot >= "16:00" && slot < "18:00" {
				filtered = append(filtered, slot)
			}
		case "noite":
			if slot >= "18:00" && slot <= "23:59" {
				filtered = append(filtered, slot)
			}
		}
	}

	return filtered
}

func periodLabel(period string) string {
	switch period {
	case "manha":
		return "de manhã"
	case "tarde":
		return "à tarde"
	case "fim_tarde":
		return "no fim da tarde"
	case "noite":
		return "à noite"
	default:
		return "nesse período"
	}
}
