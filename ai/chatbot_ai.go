package ai

import (
	"strings"
	"time"
)

type AIData struct {
	Intent  string
	Service string
	Date    string
	Time    string
}

func ParseMessage(message string) AIData {

	msg := normalize(message)

	data := AIData{}

	// INTENT
	if contains(msg, "marcar", "agendar") {
		data.Intent = "schedule"
	}

	// SERVICE
	if contains(msg, "corte") {
		data.Service = "Corte"
	}
	if contains(msg, "escova") {
		data.Service = "Escova"
	}
	if contains(msg, "progressiva") {
		data.Service = "Progressiva"
	}

	// DATE
	if contains(msg, "amanha") {
		data.Date = tomorrow()
	}
	if contains(msg, "hoje") {
		data.Date = today()
	}

	// TIME (inteligência leve)
	if contains(msg, "manha") {
		data.Time = "10:00"
	}
	if contains(msg, "tarde") {
		data.Time = "14:00"
	}
	if contains(msg, "noite") {
		data.Time = "18:00"
	}

	return data
}

func contains(msg string, words ...string) bool {
	for _, w := range words {
		if strings.Contains(msg, w) {
			return true
		}
	}
	return false
}

func normalize(msg string) string {
	msg = strings.ToLower(msg)

	replacer := strings.NewReplacer(
		"á", "a",
		"ã", "a",
		"â", "a",
		"à", "a",

		"é", "e",
		"ê", "e",

		"í", "i",

		"ó", "o",
		"ô", "o",
		"õ", "o",

		"ú", "u",
	)

	return replacer.Replace(msg)
}

func today() string {
	return time.Now().Format("2006-01-02")
}

func tomorrow() string {
	return time.Now().AddDate(0, 0, 1).Format("2006-01-02")
}
