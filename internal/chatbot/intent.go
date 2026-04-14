package chatbot

import "strings"

type Intent string

const (
	IntentGreeting     Intent = "greeting"
	IntentSchedule     Intent = "schedule"
	IntentReschedule   Intent = "reschedule"
	IntentCancel       Intent = "cancel"
	IntentList         Intent = "list"
	IntentAvailability Intent = "availability"
	IntentPrice        Intent = "price"
	IntentUnknown      Intent = "unknown"
)

func DetectIntent(text string) Intent {
	text = strings.ToLower(text)

	switch {
	case strings.Contains(text, "remarcar"):
		return IntentReschedule
	case strings.Contains(text, "cancelar"):
		return IntentCancel
	case strings.Contains(text, "agendamento") || strings.Contains(text, "meus"):
		return IntentList
	case strings.Contains(text, "oi") || strings.Contains(text, "ola"):
		return IntentGreeting
	case strings.Contains(text, "agendar") || strings.Contains(text, "marcar"):
		return IntentSchedule
	case strings.Contains(text, "horario") || strings.Contains(text, "vaga"):
		return IntentAvailability
	case strings.Contains(text, "quanto") || strings.Contains(text, "preço"):
		return IntentPrice
	default:
		return IntentUnknown
	}
}