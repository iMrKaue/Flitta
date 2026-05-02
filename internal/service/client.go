package service

import (
	"flitta/internal/database"
	"flitta/internal/repository"
	"flitta/internal/utils"
	"fmt"
	"strings"
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

func RegisterClient(name, phone, email, password, businessType string) (string, error) {

	phone = utils.NormalizeCustomerPhone(phone)
	businessType = NormalizeBusinessType(businessType)

	hashed, err := HashPassword(password)
	if err != nil {
		return "", err
	}

	clientID, err := repository.CreateClient(name, phone, email, hashed, businessType)
	if err != nil {
		if strings.Contains(err.Error(), "clients_phone_key") {
			return "", fmt.Errorf("telefone já cadastrado")
		}
		return "", err
	}

	token, err := GenerateToken(clientID)
	if err != nil {
		return "", err
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