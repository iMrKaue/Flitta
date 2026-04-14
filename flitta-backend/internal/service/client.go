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

func RegisterClient(name, phone, email, password string) (string, error) {

	phone = utils.NormalizeCustomerPhone(phone)

	hashed, err := HashPassword(password)
	if err != nil {
		return "", err
	}

	clientID, err := repository.CreateClient(name, phone, email, hashed)
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
