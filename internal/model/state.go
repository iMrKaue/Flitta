package model

type State string

const (
	// geral
	StateIdle State = ""

	// gerenciamento
	StateChooseAppointment State = "choose_appointments"
	StateChooseAction      State = "choose_action"
	StateRescheduleTime    State = "reschedule_time"
	StateCancelConfirm     State = "cancel_confirm"

	//chatbot
	StateName    State = "name"
	StateService State = "service"
	StateDate    State = "date"
	StateTime    State = "time"
	StateConfirm State = "confirm"
)
