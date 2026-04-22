package utils

import (
	"regexp"
	"strings"
	"time"
	"unicode"
)

// NormalizeCustomerPhone unifica o identificador vindo do WhatsApp/Twilio/JSON.
func NormalizeCustomerPhone(s string) string {
	s = strings.TrimSpace(s)
	for len(s) >= 9 && strings.EqualFold(s[:9], "whatsapp:") {
		s = strings.TrimSpace(s[9:])
	}
	return strings.TrimSpace(s)
}

func NormalizeText(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	return text
}

func ContainsAny(text string, terms ...string) bool {
	text = NormalizeText(text)

	for _, term := range terms {
		term = NormalizeText(term)

		// para termos curtos, exige palavra inteira
		if len(term) <= 2 {
			pattern := `(?:^|\s)` + regexp.QuoteMeta(term) + `(?:$|\s)`
			matched, _ := regexp.MatchString(pattern, text)
			if matched {
				return true
			}
			continue
		}

		if strings.Contains(text, term) {
			return true
		}
	}

	return false
}

func Capitalize(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}

	parts := strings.Fields(name)
	for i, part := range parts {
		runes := []rune(strings.ToLower(part))
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}

	return strings.Join(parts, " ")
}

func NormalizeHour(text string) string {
	text = NormalizeText(text)

	reHourFull := regexp.MustCompile(`(\d{1,2}):(\d{2})`)
	if match := reHourFull.FindStringSubmatch(text); len(match) == 3 {
		h := match[1]
		m := match[2]
		if len(h) == 1 {
			h = "0" + h
		}
		return h + ":" + m
	}

	reHourShort := regexp.MustCompile(`(\d{1,2})h`)
	if match := reHourShort.FindStringSubmatch(text); len(match) == 2 {
		h := match[1]
		if len(h) == 1 {
			h = "0" + h
		}
		return h + ":00"
	}

	reHourWithAs := regexp.MustCompile(`(?:^|\s)(?:as|às)\s*(\d{1,2})(?:\s|$)`)
	if match := reHourWithAs.FindStringSubmatch(text); len(match) == 2 {
		h := match[1]
		if len(h) == 1 {
			h = "0" + h
		}
		return h + ":00"
	}

	reOnlyNumber := regexp.MustCompile(`^\d{1,2}$`)
	if reOnlyNumber.MatchString(text) {
		if len(text) == 1 {
			text = "0" + text
		}
		return text + ":00"
	}

	return ""
}

func ExtractPeriod(text string) string {
	text = NormalizeText(text)

	switch {
	case ContainsAny(text, "fim da tarde", "final da tarde"):
		return "fim_tarde"
	case ContainsAny(text, "de manhã", "pela manhã", "de manha", "pela manha"):
		return "manha"
	case ContainsAny(text, "tarde", "a tarde", "à tarde", "de tarde"):
		return "tarde"
	case ContainsAny(text, "noite", "a noite", "à noite", "de noite"):
		return "noite"
	default:
		return ""
	}
}

func ParseDate(input string) string {
	input = NormalizeText(input)
	now := time.Now()

	if strings.Contains(input, "hoje") {
		return now.Format("2006-01-02")
	}

	if strings.Contains(input, "amanha") || strings.Contains(input, "amanhã") {
		return now.AddDate(0, 0, 1).Format("2006-01-02")
	}

	switch {
	case strings.Contains(input, "segunda-feira") || strings.Contains(input, "segunda feira") || strings.Contains(input, "segunda"):
		return nextWeekday(time.Monday).Format("2006-01-02")

	case strings.Contains(input, "terca-feira") || strings.Contains(input, "terça-feira") ||
		strings.Contains(input, "terca feira") || strings.Contains(input, "terça feira") ||
		strings.Contains(input, "terca") || strings.Contains(input, "terça"):
		return nextWeekday(time.Tuesday).Format("2006-01-02")

	case strings.Contains(input, "quarta-feira") || strings.Contains(input, "quarta feira") || strings.Contains(input, "quarta"):
		return nextWeekday(time.Wednesday).Format("2006-01-02")

	case strings.Contains(input, "quinta-feira") || strings.Contains(input, "quinta feira") || strings.Contains(input, "quinta"):
		return nextWeekday(time.Thursday).Format("2006-01-02")

	case strings.Contains(input, "sexta-feira") || strings.Contains(input, "sexta feira") || strings.Contains(input, "sexta"):
		return nextWeekday(time.Friday).Format("2006-01-02")

	case strings.Contains(input, "sabado-feira") || strings.Contains(input, "sábado-feira") ||
		strings.Contains(input, "sabado feira") || strings.Contains(input, "sábado feira") ||
		strings.Contains(input, "sabado") || strings.Contains(input, "sábado"):
		return nextWeekday(time.Saturday).Format("2006-01-02")

	case strings.Contains(input, "domingo"):
		return nextWeekday(time.Sunday).Format("2006-01-02")
	}

	return ""
}

func nextWeekday(target time.Weekday) time.Time {
	now := time.Now()
	daysAhead := int(target - now.Weekday())

	if daysAhead <= 0 {
		daysAhead += 7
	}

	return now.AddDate(0, 0, daysAhead)
}
