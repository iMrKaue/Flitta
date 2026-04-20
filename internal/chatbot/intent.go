package chatbot

import (
	"flitta/internal/utils"
)

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

type IntentResult struct {
	MainIntent   Intent
	IsPositive   bool
	IsNegative   bool
	AsksEarlier  bool
	AsksLater    bool
	AsksOtherDay bool
	WantsHuman   bool
	Period       string
}

func AnalyzeIntent(text string) IntentResult {
	t := utils.NormalizeText(text)

	result := IntentResult{
		MainIntent: IntentUnknown,
		Period:     utils.ExtractPeriod(t),
	}

	switch {
	case utils.ContainsAny(t, "remarcar", "reagendar", "mudar horario", "trocar horario", "alterar horario"):
		result.MainIntent = IntentReschedule

	case utils.ContainsAny(t, "cancelar", "desmarcar", "nao vou conseguir", "não vou conseguir"):
		result.MainIntent = IntentCancel

	case utils.ContainsAny(t, "meus agendamentos", "meu agendamento", "agendamentos", "listar agendamentos"):
		result.MainIntent = IntentList

	case utils.ContainsAny(t, "oi", "ola", "olá", "bom dia", "boa tarde", "boa noite"):
		result.MainIntent = IntentGreeting

	case utils.ContainsAny(t, "agendar", "marcar", "quero agendar", "quero marcar"):
		result.MainIntent = IntentSchedule

	case utils.ContainsAny(t, "horario", "horário", "vaga", "disponivel", "disponível", "tem horario", "tem horário"):
		result.MainIntent = IntentAvailability

	case utils.ContainsAny(t, "quanto custa", "quanto e", "quanto é", "preco", "preço", "valor"):
		result.MainIntent = IntentPrice
	}

	if utils.ContainsAny(t,
		"sim", "s", "ok", "okay", "pode ser", "fechado", "confirmo",
		"claro", "isso mesmo", "beleza", "combinado", "esse mesmo", "pode",
	) {
		result.IsPositive = true
	}

	if utils.ContainsAny(t,
		"nao", "não", "esse nao", "esse não", "outro horario", "outro horário",
		"nao quero", "não quero",
	) {
		result.IsNegative = true
	}

	if utils.ContainsAny(t, "mais cedo", "um pouco mais cedo", "horario mais cedo", "horário mais cedo") {
		result.AsksEarlier = true
	}

	if utils.ContainsAny(t, "mais tarde", "um pouco mais tarde", "horario mais tarde", "horário mais tarde") {
		result.AsksLater = true
	}

	if utils.ContainsAny(t, "outro dia", "outra data", "outro dia entao", "outro dia então") {
		result.AsksOtherDay = true
	}

	if utils.ContainsAny(t, "atendente", "humano", "pessoa", "falar com alguem", "falar com alguém") {
		result.WantsHuman = true
	}

	if result.IsNegative && result.IsPositive {
		result.IsPositive = false
	}

	return result
}

func DetectIntent(text string) Intent {
	return AnalyzeIntent(text).MainIntent
}
