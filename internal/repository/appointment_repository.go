package repository

import (
	"flitta/internal/database"
	"flitta/internal/model"
	"fmt"
	"strings"
	"time"
)

func GetAppointments(clientID int) []model.Appointment {
	rows, err := database.DB.Query(`
	    SELECT name, service, date, time
		FROM appointments
		WHERE client_id = $1
		  AND status IN ('scheduled', 'confirmed')
	`, clientID)

	if err != nil {
		fmt.Println("ERRO QUERY:", err)
		return []model.Appointment{}
	}
	defer rows.Close()

	var list []model.Appointment
	for rows.Next() {
		var a model.Appointment
		rows.Scan(&a.Name, &a.Service, &a.Date, &a.Time)
		list = append(list, a)
	}

	fmt.Println("TOTAL ENCONTRADO:", len(list))
	return list
}

func GetTodayAppointments(clientID int) ([]model.Appointment, error) {

	today := time.Now().Format("2006-01-02")

	rows, err := database.DB.Query(`
		SELECT id, name, service, date, time
		FROM appointments
		WHERE client_id = $1
		  AND date = $2
		  AND status in ('scheduled', 'confirmed')
		ORDER BY time ASC
	`, clientID, today)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.Appointment{}

	for rows.Next() {
		var a model.Appointment
		err := rows.Scan(&a.ID, &a.Name, &a.Service, &a.Date, &a.Time)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}

	return list, nil
}

func CreateAppointment(clientID int, phone, name, service, date, time string) error {
	result, err := database.DB.Exec(`
		INSERT INTO appointments (
			client_id,
			customer_phone,
			name,
			service,
			date,
			time,
			price_snapshot,
			duration_snapshot
		)
		SELECT
			$1,
			$2,
			$3,
			s.name,
			$5,
			$6,
			COALESCE(s.price, 0),
			COALESCE(s.duration, 30)
		FROM services AS s
		WHERE s.client_id = $1
			AND LOWER(s.name) = LOWER($4)
		ORDER BY s.id
		LIMIT 1
	`, clientID, phone, name, service, date, time)

	if err != nil {
		if strings.Contains(err.Error(), "unique_schedule") {
			return fmt.Errorf("horário já ocupado")
		}

		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("serviço não encontrado")
	}

	return nil
}

// AppointmentTimeRow is one appointment row used to compute occupied 30-minute slots.
type AppointmentTimeRow struct {
	Time    string
	Service string
}

func ListAppointmentSlotsForDate(clientID int, date string) ([]AppointmentTimeRow, error) {
	rows, err := database.DB.Query(`
	    SELECT time, service
		FROM appointments
		WHERE client_id = $1
		   AND date = $2
		   AND status IN ('scheduled', 'confirmed')
	`, clientID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AppointmentTimeRow
	for rows.Next() {
		var r AppointmentTimeRow
		if err := rows.Scan(&r.Time, &r.Service); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func ListAppointmentSlotsForDateExcept(clientID int, date string, exceptID int) ([]AppointmentTimeRow, error) {
	rows, err := database.DB.Query(`
		SELECT time, service
		FROM appointments
		WHERE client_id = $1
		  AND date = $2
		  AND id <> $3
		  AND status IN ('scheduled', 'confirmed')
	`, clientID, date, exceptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AppointmentTimeRow
	for rows.Next() {
		var r AppointmentTimeRow
		if err := rows.Scan(&r.Time, &r.Service); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func GetAppointmentForReschedule(
	id int,
	customerPhone string,
) (date string, clientID int, service string, err error) {
	err = database.DB.QueryRow(`
	    SELECT date, client_id, service
		FROM appointments
		WHERE id = $1
		  AND customer_phone = $2
		  AND status IN ('scheduled', 'confirmed')
	`, id, customerPhone).Scan(&date, &clientID, &service)

	return
}

func UpdateAppointmentTimeByID(id int, newTime string) error {
	result, err := database.DB.Exec(`
	    UPDATE appointments
		SET time = $1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	      AND status IN ('scheduled', 'confirmed')
	`, newTime, id)

	if err != nil {
		if strings.Contains(err.Error(), "unique_schedule") {
			return fmt.Errorf("horário já ocupado")
		}

		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agendamento não encontrado ou indisponível")
	}

	return nil
}

func CancelCustomerAppointment(
	id int,
	clientID int,
	customerPhone string,
) (rowsAffected int64, err error) {
	result, err := database.DB.Exec(`
	    UPDATE appointments
		SET status = 'cancelled',
		    cancelled_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND client_id = $2
		  AND customer_phone = $3
		  AND status IN ('scheduled', 'confirmed')
		  `, id, clientID, customerPhone)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func UpdateAppointmentOutcome(
	id int,
	clientID int,
	status string,
) (rowsAffected int64, err error) {
	if status != "completed" && status != "no_show" {
		return 0, fmt.Errorf("status de atendimento inválido")
	}

	result, err := database.DB.Exec(`
		UPDATE appointments
		SET status = $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
			AND client_id = $3
			AND status IN ('scheduled', 'confirmed')
			AND (date::date + time::time) <= (NOW() AT TIME ZONE 'America/Sao_Paulo')
	`, status, id, clientID)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func GetWorkingHours(clientID int) (start, end string, interval int, err error) {
	err = database.DB.QueryRow(`
		SELECT start_time, end_time, interval_minutes
		FROM working_hours
		WHERE client_id = $1
	`, clientID).Scan(&start, &end, &interval)
	return
}

func GetAppointmentsByClient(clientID int) ([]model.Appointment, error) {
	rows, err := database.DB.Query(`
		SELECT id, client_id, name, service, date, time, customer_phone, status
		FROM appointments
		WHERE client_id = $1
		  AND date::date >= CURRENT_DATE
		  AND status IN ('scheduled', 'confirmed')
		ORDER BY date ASC, time ASC
	`, clientID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []model.Appointment

	for rows.Next() {
		var a model.Appointment

		err := rows.Scan(
			&a.ID,
			&a.ClientID,
			&a.Name,
			&a.Service,
			&a.Date,
			&a.Time,
			&a.CustomerPhone,
			&a.Status,
		)
		if err != nil {
			return nil, err
		}

		appointments = append(appointments, a)
	}

	return appointments, rows.Err()

}

func GetPendingReminderAppointments(clientID int, hoursBefore int) ([]model.Appointment, error) {
	rows, err := database.DB.Query(`
		SELECT id, client_id, name, service, date, time, customer_phone, reminder_sent
		FROM appointments
		WHERE client_id = $1
		  AND status IN ('scheduled', 'confirmed')
		  AND COALESCE(reminder_sent, false) = false
		  AND (date::date + time::time) >= NOW()
		  AND (date::date + time::time) <= NOW() + ($2::text || ' hours')::interval
		ORDER BY date ASC, time ASC
		`, clientID, hoursBefore)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []model.Appointment

	for rows.Next() {
		var a model.Appointment

		err := rows.Scan(
			&a.ID,
			&a.ClientID,
			&a.Name,
			&a.Service,
			&a.Date,
			&a.Time,
			&a.CustomerPhone,
			&a.ReminderSent,
		)
		if err != nil {
			return nil, err
		}

		reminders = append(reminders, a)
	}

	return reminders, rows.Err()
}

func MarkReminderAsSent(appointmentID int, clientID int) error {
	result, err := database.DB.Exec(`
		UPDATE appointments
		SET reminder_sent = true,
			reminder_sent_at = NOW(),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
			AND client_id = $2
			AND status IN ('scheduled', 'confirmed')
	`, appointmentID, clientID)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agendamento não encontrado ou indisponível")
	}

	return nil
}

func GetAppointmentByID(appointmentID int, clientID int) (model.Appointment, error) {
	var a model.Appointment

	err := database.DB.QueryRow(`
		SELECT id, client_id, name, service, date, time, customer_phone, reminder_sent
		FROM appointments
		WHERE id = $1
		  AND client_id = $2
		  AND status IN ('scheduled', 'confirmed')
	`, appointmentID, clientID).Scan(
		&a.ID,
		&a.ClientID,
		&a.Name,
		&a.Service,
		&a.Date,
		&a.Time,
		&a.CustomerPhone,
		&a.ReminderSent,
	)

	return a, err
}

type PendingReminderWithCompany struct {
	model.Appointment
	CompanyName string
}

func GetAllPendingReminderAppointments(hoursBefore int) ([]PendingReminderWithCompany, error) {
	rows, err := database.DB.Query(`
		SELECT 
			a.id,
			a.client_id,
			a.name,
			a.service,
			a.date,
			a.time,
			a.customer_phone,
			a.reminder_sent,
			c.name
		FROM appointments a
		INNER JOIN clients c ON c.id = a.client_id
		WHERE COALESCE(a.reminder_sent, false) = false
		  AND a.status IN ('scheduled', 'confirmed')
		  AND (a.date::date + a.time::time) >= NOW()
		  AND (a.date::date + a.time::time) <= NOW() + ($1::text || ' hours')::interval
		ORDER BY a.date ASC, a.time ASC
	`, hoursBefore)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []PendingReminderWithCompany

	for rows.Next() {
		var reminder PendingReminderWithCompany

		err := rows.Scan(
			&reminder.ID,
			&reminder.ClientID,
			&reminder.Name,
			&reminder.Service,
			&reminder.Date,
			&reminder.Time,
			&reminder.CustomerPhone,
			&reminder.ReminderSent,
			&reminder.CompanyName,
		)
		if err != nil {
			return nil, err
		}

		reminders = append(reminders, reminder)
	}

	return reminders, rows.Err()
}
