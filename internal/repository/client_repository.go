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
