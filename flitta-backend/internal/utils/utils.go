package utils

import (
	"strings"
	"time"
)

// NormalizeCustomerPhone unifica o identificador vindo do WhatsApp/Twilio/JSON.
func NormalizeCustomerPhone(s string) string {
	s = strings.TrimSpace(s)
	for len(s) >= 9 && strings.EqualFold(s[:9], "whatsapp:") {
		s = strings.TrimSpace(s[9:])
	}
	return strings.TrimSpace(s)
}

func Capitalize(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(string(name[0])) + strings.ToLower(name[1:])
}

func ParseDate(input string) string {

	input = strings.ToLower(strings.TrimSpace(input))

	now := time.Now()

	switch input {

	case "hoje":
		return now.Format("2006-01-02")

	case "amanha", "amanhã":
		return now.AddDate(0, 0, 1).Format("2006-01-02")

	case "segunda":
		return nextWeekday(time.Monday).Format("2006-01-02")

	case "terca", "terça":
		return nextWeekday(time.Tuesday).Format("2006-01-02")

	case "quarta":
		return nextWeekday(time.Wednesday).Format("2006-01-02")

	case "quinta":
		return nextWeekday(time.Thursday).Format("2006-01-02")

	case "sexta":
		return nextWeekday(time.Friday).Format("2006-01-02")

	case "sabado", "sábado":
		return nextWeekday(time.Saturday).Format("2006-01-02")

	case "domindo":
		return nextWeekday(time.Sunday).Format("2006-01-02")
	}

	return input
}

func nextWeekday(target time.Weekday) time.Time {
	now := time.Now()

	daysAhead := int(target - now.Weekday())

	if daysAhead <= 0 {
		daysAhead += 7
	}

	return now.AddDate(0, 0, daysAhead)
}
