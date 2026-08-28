package service

import (
	"errors"
	"flitta/internal/database"
	"flitta/internal/repository"
	"flitta/internal/utils"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

var (
	ErrPhoneAlreadyRegistered = errors.New(
		"telefone já cadastrado",
	)

	ErrEmailAlreadyRegistered = errors.New(
		"email já cadastrado",
	)
)

func GetClientByPhone(phone string) (int, error) {

	phone = utils.NormalizeCustomerPhone(phone)

	var id int

	err := database.DB.QueryRow(
		"SELECT id FROM clients WHERE phone = $1",
		phone,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func RegisterClient(
	name string,
	phone string,
	email string,
	password string,
	businessType string,
) (string, error) {
	name = strings.TrimSpace(name)

	phone = utils.NormalizeCustomerPhone(phone)

	email = strings.ToLower(
		strings.TrimSpace(email),
	)

	businessType = NormalizeBusinessType(
		businessType,
	)

	hashed, err := HashPassword(password)
	if err != nil {
		return "", fmt.Errorf(
			"gerar hash da senha: %w",
			err,
		)
	}

	defaultServices := DefaultServicesByBusinessType(
		businessType,
	)

	clientID, err := repository.CreateClientWithDefaults(
		name,
		phone,
		email,
		hashed,
		businessType,
		defaultServices,
	)

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			switch pqErr.Constraint {
			case "clients_phone_key":
				return "", ErrPhoneAlreadyRegistered

			case "clients_email_normalized_unique":
				return "", ErrEmailAlreadyRegistered
			}
		}

		return "", err
	}

	token, err := GenerateToken(clientID)
	if err != nil {
		return "", fmt.Errorf(
			"gerar token após cadastro: %w",
			err,
		)
	}

	return token, nil
}

func NormalizeBusinessType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))

	switch value {
	case "beauty", "salon", "salao", "salão", "estetica", "estética":
		return "beauty"
	case "barber", "barbearia":
		return "barber"
	case "clinic", "clinica", "clínica":
		return "clinic"
	case "gym", "academia", "personal":
		return "gym"
	case "petshop", "pet", "pet_shop":
		return "petshop"
	case "other", "outro", "":
		return "other"
	default:
		return "other"
	}
}

func DefaultServicesByBusinessType(businessType string) []string {
	businessType = NormalizeBusinessType(businessType)

	switch businessType {
	case "beauty":
		return []string{"Corte", "Escova", "Progressiva"}

	case "barber":
		return []string{"Corte", "Barba", "Sobrancelha"}

	case "clinic":
		return []string{"Consulta", "Retorno", "Avaliação"}

	case "gym":
		return []string{
			"Avaliação física",
			"Aula experimental",
			"Personal",
		}

	case "petshop":
		return []string{"Banho", "Tosa", "Consulta"}

	default:
		return []string{"Atendimento"}
	}
}
