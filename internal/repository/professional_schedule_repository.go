package repository

import (
	"database/sql"
	"errors"
	"flitta/internal/database"
	"flitta/internal/model"
	"sort"
	"time"
)

var (
	ErrProfessionalWorkingHoursInvalid = errors.New(
		"horário do profissional inválido",
	)

	ErrProfessionalWorkingHoursOverlap = errors.New(
		"existem horários sobrepostos para o profissional",
	)
)

type professionalTimeRange struct {
	start time.Time
	end   time.Time
}

func validateProfessionalWorkingHours(
	hours []model.ProfessionalWorkingHour,
) error {
	rangesByWeekday := make(
		map[int][]professionalTimeRange,
	)

	for _, hour := range hours {
		if hour.Weekday < 1 || hour.Weekday > 7 {
			return ErrProfessionalWorkingHoursInvalid
		}

		start, err := time.Parse(
			"15:04",
			hour.Start,
		)
		if err != nil {
			return ErrProfessionalWorkingHoursInvalid
		}

		end, err := time.Parse(
			"15:04",
			hour.End,
		)
		if err != nil {
			return ErrProfessionalWorkingHoursInvalid
		}

		if !start.Before(end) {
			return ErrProfessionalWorkingHoursInvalid
		}

		rangesByWeekday[hour.Weekday] = append(
			rangesByWeekday[hour.Weekday],
			professionalTimeRange{
				start: start,
				end:   end,
			},
		)
	}

	for _, ranges := range rangesByWeekday {
		sort.Slice(
			ranges,
			func(i, j int) bool {
				return ranges[i].start.Before(
					ranges[j].start,
				)
			},
		)

		for i := 1; i < len(ranges); i++ {
			previous := ranges[i-1]
			current := ranges[i]

			if current.start.Before(previous.end) {
				return ErrProfessionalWorkingHoursOverlap
			}
		}
	}

	return nil
}

func ReplaceProfessionalWorkingHours(
	clientID int,
	professionalID int,
	hours []model.ProfessionalWorkingHour,
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

	if err := validateProfessionalWorkingHours(
		hours,
	); err != nil {
		return err
	}

	_, err = tx.Exec(`
DELETE FROM professional_working_hours
WHERE client_id = $1
  AND professional_id = $2
`,
		clientID,
		professionalID,
	)

	if err != nil {
		return err
	}

	for _, hour := range hours {
		_, err := tx.Exec(`
INSERT INTO professional_working_hours (
client_id,
professional_id,
weekday,
start_time,
end_time
)
VALUES (
$1,
$2,
$3,
CAST($4 AS TIME),
CAST($5 AS TIME)
)
`,
			clientID,
			professionalID,
			hour.Weekday,
			hour.Start,
			hour.End,
		)

		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func ListProfessionalWorkingHours(
	clientID int,
	professionalID int,
) ([]model.ProfessionalWorkingHour, error) {
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

	rows, err := database.DB.Query(`
SELECT
id,
weekday,
TO_CHAR(start_time, 'HH24:MI'),
TO_CHAR(end_time, 'HH24:MI')
FROM professional_working_hours
WHERE client_id = $1
  AND professional_id = $2
ORDER BY
weekday ASC,
start_time ASC,
end_time ASC,
id ASC
`,
		clientID,
		professionalID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	hours := []model.ProfessionalWorkingHour{}

	for rows.Next() {
		var hour model.ProfessionalWorkingHour

		if err := rows.Scan(
			&hour.ID,
			&hour.Weekday,
			&hour.Start,
			&hour.End,
		); err != nil {
			return nil, err
		}

		hours = append(
			hours,
			hour,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hours, nil
}
