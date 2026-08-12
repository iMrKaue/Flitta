package model

// SalonService representa um serviço ou atendimento cadastrado pelo estabelecimento.
// O nome será refatorado futuramente para Service.
type SalonService struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Duration int     `json:"duration"`
	Price    float64 `json:"price"`
}
