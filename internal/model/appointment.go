package model

// Appointment é um agendamento no banco (cliente final × estabelecimento).
type Appointment struct {
	ID            int    `json:"id"`
	ClientID      int    `json:"client_id,omitempty"`
	Name          string `json:"name"`
	Service       string `json:"service"`
	Date          string `json:"date"`
	Time          string `json:"time"`
	CustomerPhone string `json:"customer_phone,omitempty"`
	ReminderSent  bool   `json:"reminder_sent,omitempty"`
	Status        string `json:"status,omitempty"`
}
