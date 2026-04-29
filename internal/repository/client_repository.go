package repository

import "flitta/internal/database"

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
