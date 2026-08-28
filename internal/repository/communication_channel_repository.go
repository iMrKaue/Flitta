package repository

import "flitta/internal/database"

func GetClientIDByCommunicationChannel(
	provider string,
	channelType string,
	externalAddress string,
) (int, error) {
	var clientID int

	err := database.DB.QueryRow(`
		SELECT client_id
		FROM communication_channels
		WHERE provider = $1
			AND channel_type = $2
			AND external_address = $3
			AND active = TRUE
	`, provider, channelType, externalAddress).Scan(&clientID)

	if err != nil {
		return 0, err
	}

	return clientID, nil
}
