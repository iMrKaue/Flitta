package repository

import "flitta/internal/database"

func CreateClient(name, phone, email, password string) (int, error) {
	var id int

	err := database.DB.QueryRow(`
		INSERT INTO clients (name, phone, email, password)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, name, phone, email, password).Scan(&id)

	return id, err
}
