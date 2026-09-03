package model

// ProfessionalTimeOff representa uma indisponibilidade pontual de um profissional.
type ProfessionalTimeOff struct {
	ID             int    `json:"id,omitempty"`
	ClientID       int    `json:"client_id,omitempty"`
	ProfessionalID int    `json:"professional_id"`
	StartAt        string `json:"start_at"`
	EndAt          string `json:"end_at"`
	Reason         string `json:"reason,omitempty"`
}
