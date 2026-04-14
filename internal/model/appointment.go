package model

// Appointment é um agendamento no banco (cliente final × salão).
type Appointment struct {
	ID      int
	Name    string
	Service string
	Date    string
	Time    string
}
