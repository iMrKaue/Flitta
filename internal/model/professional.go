package model

// Professional representa uma pessoa que realiza atendimentos
// dentro de um estabelecimento.
type Professional struct {
	ID       int            `json:"id"`
	ClientID int            `json:"client_id,omitempty"`
	Name     string         `json:"name"`
	Active   bool           `json:"active"`
	Services []SalonService `json:"services"`
}
