package model

// SalonService é um serviço oferecido pelo salão (catálogo), não o pacote internal/service.
type SalonService struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
