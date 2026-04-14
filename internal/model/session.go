package model

type Session struct {
	Phone    string
	ClientID int
	State    State
	Name     string
	Service  string
	Date     string
	Time     string

	SelectedAppointmentID int
}
