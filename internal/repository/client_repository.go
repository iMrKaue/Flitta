package repository

import (
	"flitta/internal/database"
	"fmt"
)

type ClientSettings struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	BusinessType string `json:"business_type"`
}

func CreateClientWithDefaults(
	businessName string,
	responsibleName string,
	phone string,
	email string,
	password string,
	businessType string,
	defaultServices []string,
) (int, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return 0, fmt.Errorf(
			"iniciar transação de cadastro: %w",
			err,
		)
	}

	defer tx.Rollback()

	var clientID int

	err = tx.QueryRow(`
		INSERT INTO clients (
			name,
			phone,
			business_type
		)
		VALUES ($1, $2, $3)
		RETURNING id
	`,
		businessName,
		phone,
		businessType,
	).Scan(&clientID)

	if err != nil {
		return 0, fmt.Errorf(
			"criar cliente: %w",
			err,
		)
	}

	if _, err := tx.Exec(`
			INSERT INTO users (
					client_id,
					name,
					email,
					password,
					role,
					active
			)
			VALUES ($1, $2, $3, $4, $5, $6)
	`,
		clientID,
		responsibleName,
		email,
		password,
		"owner",
		true,
	); err != nil {
		return 0, fmt.Errorf(
			"criar usuário proprietário: %w",
			err,
		)
	}

	for _, serviceName := range defaultServices {
		if _, err := tx.Exec(`
			INSERT INTO services (
				client_id,
				name
			)
			VALUES ($1, $2)
		`, clientID, serviceName); err != nil {
			return 0, fmt.Errorf(
				"criar serviço padrão %q: %w",
				serviceName,
				err,
			)
		}
	}

	if _, err := tx.Exec(`
		INSERT INTO working_hours (
			client_id,
			start_time,
			end_time,
			interval_minutes
		)
		VALUES ($1, $2, $3, $4)
	`,
		clientID,
		"09:00",
		"18:00",
		60,
	); err != nil {
		return 0, fmt.Errorf(
			"criar horário padrão: %w",
			err,
		)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf(
			"confirmar cadastro: %w",
			err,
		)
	}

	return clientID, nil
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
		SELECT id, name, phone, business_type
		FROM clients
		WHERE id = $1
	`, clientID).Scan(
		&client.ID,
		&client.Name,
		&client.Phone,
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
