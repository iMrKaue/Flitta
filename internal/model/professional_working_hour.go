package model

// ProfessionalWorkingHour representa uma faixa recorrente de trabalho
// de um profissional em um dia da semana.
//
// Weekday segue o padrão ISO:
// 1 = segunda-feira
// 2 = terça-feira
// 3 = quarta-feira
// 4 = quinta-feira
// 5 = sexta-feira
// 6 = sábado
// 7 = domingo
type ProfessionalWorkingHour struct {
	ID      int    `json:"id,omitempty"`
	Weekday int    `json:"weekday"`
	Start   string `json:"start"`
	End     string `json:"end"`
}
