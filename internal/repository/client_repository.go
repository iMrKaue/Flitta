package repository

import "flitta/internal/database"

type ClientSettings struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	BusinessType string `json:"business_type"`
}

func CreateClient(name, phone, email, password, businessType string) (int, error) {
	var id int

	err := database.DB.QueryRow(`
		INSERT INTO clients (name, phone, email, password, business_type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, name, phone, email, password, businessType).Scan(&id)

	return id, err
}

func GetBusinessTypeByClientID(clientID int) string {
	var businessType string

	err := database.DB.QueryRow(`
		SELECT business_type
		FROM clients
		WHERE id = $1
	`, clientID).Scan(&businessType)

	if err != nil || businessType == "" {
		return "other"
	}

	return businessType
}

func GetClientSettings(clientID int) (ClientSettings, error) {
	var client ClientSettings

	err := database.DB.QueryRow(`
		SELECT id, name, phone, email, business_type
		FROM clients
		WHERE id = $1
	`, clientID).Scan(
		&client.ID,
		&client.Name,
		&client.Phone,
		&client.Email,
		&client.BusinessType,
	)

	return client, err
}

func UpdateClientSettings(clientID int, name, phone, businessType string) error {
	_, err := database.DB.Exec(`
		UPDATE clients
		SET name = $1,
		    phone = $2,
		    business_type = $3
		WHERE id = $4
	`, name, phone, businessType, clientID)

	return err
}
