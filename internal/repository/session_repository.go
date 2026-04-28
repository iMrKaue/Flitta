package repository

import (
	"flitta/internal/database"
	"flitta/internal/model"
	"fmt"
)

func GetSession(phone string) (model.Session, error) {
	var s model.Session

	err := database.DB.QueryRow(`
		SELECT phone, client_id, state, name, service, date, time,
		COALESCE(selected_appointment_id, 0),
		COALESCE(suggested_time, '')
		FROM user_sessions
		WHERE phone = $1
	`, phone).Scan(
		&s.Phone,
		&s.ClientID,
		&s.State,
		&s.Name,
		&s.Service,
		&s.Date,
		&s.Time,
		&s.SelectedAppointmentID,
		&s.SuggestedTime,
	)

	return s, err
}

func SaveSession(s model.Session) error {
	_, err := database.DB.Exec(`
	INSERT INTO user_sessions (phone, client_id, state, name, service, date, time, selected_appointment_id, suggested_time)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	ON CONFLICT (phone)
	DO UPDATE SET
		state = EXCLUDED.state,
		name = EXCLUDED.name,
		service = EXCLUDED.service,
		date = EXCLUDED.date,
		time = EXCLUDED.time,
		selected_appointment_id = EXCLUDED.selected_appointment_id,
		suggested_time = EXCLUDED.suggested_time,
		updated_at = CURRENT_TIMESTAMP
`,
		s.Phone,
		s.ClientID,
		s.State,
		s.Name,
		s.Service,
		s.Date,
		s.Time,
		s.SelectedAppointmentID,
		s.SuggestedTime,
	)

	return err
}

func DeleteSession(phone string) {
	database.DB.Exec("DELETE FROM user_sessions WHERE phone = $1", phone)
}

func GetServices(clientID int) []model.SalonService {
	rows, err := database.DB.Query(`
		SELECT id, name FROM services WHERE client_id = $1
	`, clientID)

	if err != nil {
		fmt.Println("Erro ao buscar serviços:", err)
		return []model.SalonService{}
	}
	defer rows.Close()

	list := make([]model.SalonService, 0)
	for rows.Next() {
		var s model.SalonService
		rows.Scan(&s.ID, &s.Name)
		list = append(list, s)
	}

	return list
}

func CreateService(clientID int, name string) error {
	_, err := database.DB.Exec(`
		INSERT INTO services (client_id, name)
		VALUES ($1, $2)
	`, clientID, name)

	return err
}

func DeleteService(clientID int, name string) error {
	_, err := database.DB.Exec(`
		DELETE FROM services WHERE client_id = $1 AND name = $2
	`, clientID, name)

	return err
}

func SetWorkingHours(clientID int, start, end string, interval int) error {
	res, err := database.DB.Exec(`
		UPDATE working_hours
		SET start_time = $2, end_time = $3, interval_minutes = $4
		WHERE client_id = $1
	`, clientID, start, end, interval)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err = database.DB.Exec(`
		INSERT INTO working_hours (client_id, start_time, end_time, interval_minutes)
		VALUES ($1, $2, $3, $4)
	`, clientID, start, end, interval)
	return err
}

func GetServiceDuration(clientID int, serviceName string) int {
	var duration *int

	err := database.DB.QueryRow(`
		SELECT COALESCE(duration, 30)
		FROM services
		WHERE client_id = $1 AND LOWER(name) = LOWER($2)
	`, clientID, serviceName).Scan(&duration)

	if err != nil {
		fmt.Println("Erro ao buscar duração:", err)
		return 30
	}

	return *duration
}
