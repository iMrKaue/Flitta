package repository

import (
	"database/sql"
	"errors"
	"flitta/internal/database"
	"flitta/internal/model"
	"strings"
)

var (
	ErrProfessionalNotFound = errors.New(
		"profissional não encontrado",
	)

	ErrProfessionalNameRequired = errors.New(
		"nome do profissional é obrigatório",
	)

	ErrProfessionalServiceNotFound = errors.New(
		"serviço não pertence ao estabelecimento",
	)
)

func CreateProfessional(
	clientID int,
	name string,
) (model.Professional, error) {
	var professional model.Professional

	name = strings.TrimSpace(name)

	if name == "" {
		return professional, ErrProfessionalNameRequired
	}

	err := database.DB.QueryRow(`
		INSERT INTO professionals (
			client_id,
			name
		)
		VALUES ($1, $2)
		RETURNING
			id,
			client_id,
			name,
			active
	`,
		clientID,
		name,
	).Scan(
		&professional.ID,
		&professional.ClientID,
		&professional.Name,
		&professional.Active,
	)

	if err != nil {
		return professional, err
	}

	professional.Services = []model.SalonService{}

	return professional, nil
}

func ListProfessionals(
	clientID int,
) ([]model.Professional, error) {
	rows, err := database.DB.Query(`
		SELECT
			p.id,
			p.client_id,
			p.name,
			p.active,
			COALESCE(s.id, 0),
			COALESCE(s.name, ''),
			COALESCE(s.duration, 30),
			COALESCE(s.price, 0)::DOUBLE PRECISION
		FROM professionals p

		LEFT JOIN professional_services ps
			ON ps.professional_id = p.id
			AND ps.client_id = p.client_id

		LEFT JOIN services s
			ON s.id = ps.service_id
			AND s.client_id = ps.client_id

		WHERE p.client_id = $1

		ORDER BY
			p.name ASC,
			s.name ASC,
			s.id ASC
	`, clientID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	professionals := []model.Professional{}

	indexByID := make(map[int]int)

	for rows.Next() {
		var professional model.Professional

		var service model.SalonService

		if err := rows.Scan(
			&professional.ID,
			&professional.ClientID,
			&professional.Name,
			&professional.Active,
			&service.ID,
			&service.Name,
			&service.Duration,
			&service.Price,
		); err != nil {
			return nil, err
		}

		index, exists := indexByID[professional.ID]

		if !exists {
			professional.Services = []model.SalonService{}

			professionals = append(
				professionals,
				professional,
			)

			index = len(professionals) - 1

			indexByID[professional.ID] = index
		}

		if service.ID > 0 {
			professionals[index].Services = append(
				professionals[index].Services,
				service,
			)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return professionals, nil
}

func SetProfessionalActive(
	clientID int,
	professionalID int,
	active bool,
) error {
	result, err := database.DB.Exec(`
		UPDATE professionals
		SET
			active = $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		  AND client_id = $3
	`,
		active,
		professionalID,
		clientID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrProfessionalNotFound
	}

	return nil
}

func ReplaceProfessionalServices(
	clientID int,
	professionalID int,
	serviceIDs []int,
) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	var exists int

	err = tx.QueryRow(`
		SELECT 1
		FROM professionals
		WHERE id = $1
		  AND client_id = $2
	`,
		professionalID,
		clientID,
	).Scan(&exists)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrProfessionalNotFound
	}

	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		DELETE FROM professional_services
		WHERE client_id = $1
		  AND professional_id = $2
	`,
		clientID,
		professionalID,
	)

	if err != nil {
		return err
	}

	seen := make(map[int]bool)

	for _, serviceID := range serviceIDs {
		if serviceID <= 0 {
			return ErrProfessionalServiceNotFound
		}

		if seen[serviceID] {
			continue
		}

		seen[serviceID] = true

		result, err := tx.Exec(`
			INSERT INTO professional_services (
				client_id,
				professional_id,
				service_id
			)
			SELECT
				$1,
				$2,
				s.id
			FROM services s
			WHERE s.id = $3
			  AND s.client_id = $1
		`,
			clientID,
			professionalID,
			serviceID,
		)

		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected != 1 {
			return ErrProfessionalServiceNotFound
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
