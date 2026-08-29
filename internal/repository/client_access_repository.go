package repository

import (
	"database/sql"
	"flitta/internal/database"
	"time"
)

func GetClientAccessState(
	clientID int,
) (string, *time.Time, error) {
	var status string
	var expiresAt sql.NullTime

	err := database.DB.QueryRow(`
		SELECT status, pilot_expires_at
		FROM clients
		WHERE id = $1
	`, clientID).Scan(
		&status,
		&expiresAt,
	)

	if err != nil {
		return "", nil, err
	}

	if !expiresAt.Valid {
		return status, nil, nil
	}

	expiration := expiresAt.Time

	return status, &expiration, nil
}
