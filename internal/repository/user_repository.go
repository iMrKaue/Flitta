package repository

import "flitta/internal/database"

func GetActiveUserCredentialsByEmail(
	email string,
) (int, string, error) {
	var clientID int
	var passwordHash string

	err := database.DB.QueryRow(`
		SELECT client_id, password
		FROM users
		WHERE LOWER(BTRIM(email)) = $1
			AND active = TRUE
		`, email).Scan(
		&clientID,
		&passwordHash,
	)

	return clientID, passwordHash, err
}
