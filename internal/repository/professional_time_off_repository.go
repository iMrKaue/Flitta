package repository

import (
	"database/sql"
	"errors"
	"flitta/internal/database"
	"flitta/internal/model"
	"strings"
	"time"
)

var ErrProfessionalTimeOffInvalid = errors.New(
	"período de indisponibilidade inválido",
)

var ErrProfessionalTimeOffNotFound = errors.New(
	"indisponibilidade de profissional não encontrada",
)

var ErrProfessionalTimeOffOverlap = errors.New(
	"já existe uma indisponibilidade nesse período",
)

const professionalTimeOffLayout = "2006-01-02T15:04"

func CreateProfessionalTimeOff(
	clientID int,
	professionalID int,
	startAt string,
	endAt string,
	reason string,
) (model.ProfessionalTimeOff, error) {
	var timeOff model.ProfessionalTimeOff

	start, err := time.Parse(
		professionalTimeOffLayout,
		startAt,
	)
	if err != nil {
		return timeOff, ErrProfessionalTimeOffInvalid
	}

	end, err := time.Parse(
		professionalTimeOffLayout,
		endAt,
	)
	if err != nil {
		return timeOff, ErrProfessionalTimeOffInvalid
	}

	if !start.Before(end) {
		return timeOff, ErrProfessionalTimeOffInvalid
	}

	reason = strings.TrimSpace(reason)

	var overlap bool

	err = database.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM professional_time_off
			WHERE client_id = $1
			  AND professional_id = $2
			  AND start_at < CAST($4 AS TIMESTAMP)
			  AND end_at > CAST($3 AS TIMESTAMP)
		)
	`,
		clientID,
		professionalID,
		startAt,
		endAt,
	).Scan(&overlap)

	if err != nil {
		return timeOff, err
	}

	if overlap {
		return timeOff, ErrProfessionalTimeOffOverlap
	}

	err = database.DB.QueryRow(`
		INSERT INTO professional_time_off (
			client_id,
			professional_id,
			start_at,
			end_at,
			reason
		)
		SELECT
			$1,
			p.id,
			CAST($3 AS TIMESTAMP),
			CAST($4 AS TIMESTAMP),
			NULLIF($5, '')
		FROM professionals p
		WHERE p.id = $2
	  	AND p.client_id = $1
		RETURNING
			id,
			client_id,
			professional_id,
			TO_CHAR(start_at, 'YYYY-MM-DD"T"HH24:MI'),
			TO_CHAR(end_at, 'YYYY-MM-DD"T"HH24:MI'),
			COALESCE(reason, '')
	`,
		clientID,
		professionalID,
		startAt,
		endAt,
		reason,
	).Scan(
		&timeOff.ID,
		&timeOff.ClientID,
		&timeOff.ProfessionalID,
		&timeOff.StartAt,
		&timeOff.EndAt,
		&timeOff.Reason,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return timeOff, ErrProfessionalNotFound
	}

	if err != nil {
		return timeOff, err
	}

	return timeOff, nil
}

func ListProfessionalTimeOff(
	clientID int,
	professionalID int,
) ([]model.ProfessionalTimeOff, error) {
	var exists int

	err := database.DB.QueryRow(`
		SELECT 1
		FROM professionals
		WHERE id = $1
			AND client_id = $2
	`,
		professionalID,
		clientID,
	).Scan(&exists)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProfessionalNotFound
	}

	if err != nil {
		return nil, err
	}

	timeOffs := make([]model.ProfessionalTimeOff, 0)

	rows, err := database.DB.Query(`
		SELECT
			id,
			client_id,
			professional_id,
			TO_CHAR(start_at, 'YYYY-MM-DD"T"HH24:MI'),
			TO_CHAR(end_at, 'YYYY-MM-DD"T"HH24:MI'),
			COALESCE(reason, '')
		FROM professional_time_off
		WHERE client_id = $1
			AND professional_id = $2
		ORDER BY start_at, end_at, id
	`,
		clientID,
		professionalID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var timeOff model.ProfessionalTimeOff

		err := rows.Scan(
			&timeOff.ID,
			&timeOff.ClientID,
			&timeOff.ProfessionalID,
			&timeOff.StartAt,
			&timeOff.EndAt,
			&timeOff.Reason,
		)
		if err != nil {
			return nil, err
		}

		timeOffs = append(
			timeOffs,
			timeOff,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return timeOffs, nil
}

func DeleteProfessionalTimeOff(
	clientID int,
	timeOffId int,
) error {
	result, err := database.DB.Exec(`
		DELETE FROM professional_time_off
		WHERE id = $1
			AND client_id = $2
	`,
		timeOffId,
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
		return ErrProfessionalTimeOffNotFound
	}

	return nil
}
