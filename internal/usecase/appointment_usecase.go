package usecase

import (
	"flitta/internal/database"
	"flitta/internal/model"
	"flitta/internal/repository"
	"flitta/internal/utils"
	"fmt"
	"math"
	"strings"
	"time"
)

type AppointmentUsecase struct{}

func NewAppointmentUsecase() *AppointmentUsecase {
	return &AppointmentUsecase{}
}

func bookedSlotMapFromRows(clientID int, rows []repository.AppointmentTimeRow) map[string]bool {
	booked := map[string]bool{}
	layout := "15:04"
	for _, row := range rows {
		duration := repository.GetServiceDuration(clientID, row.Service)
		start, _ := time.Parse(layout, row.Time)
		slots := int(math.Ceil(float64(duration) / 30.0))
		for i := 0; i < slots; i++ {
			slot := start.Add(time.Duration(i*30) * time.Minute)
			booked[slot.Format("15:04")] = true
		}
	}
	return booked
}

func (u *AppointmentUsecase) CreateAppointment(
	clientID int,
	customerPhone,
	name,
	serviceName,
	date,
	timeStr string,
) error {

	phone := utils.NormalizeCustomerPhone(customerPhone)

	duration := repository.GetServiceDuration(clientID, serviceName)

	booked, _ := u.getBookedTimes(clientID, date)

	layout := "15:04"
	start, err := time.Parse(layout, timeStr)
	if err != nil {
		return fmt.Errorf("horário inválido")
	}

	slots := int(math.Ceil(float64(duration) / 30.0))

	for i := 0; i < slots; i++ {
		slot := start.Add(time.Duration(i*30) * time.Minute)

		if booked[slot.Format("15:04")] {
			return fmt.Errorf("horário em conflito")
		}
	}

	return repository.CreateAppointment(
		clientID,
		phone,
		name,
		serviceName,
		date,
		timeStr,
	)
}

func (u *AppointmentUsecase) getBookedTimes(clientID int, date string) (map[string]bool, error) {
	rows, err := repository.ListAppointmentSlotsForDate(clientID, date)
	if err != nil {
		return map[string]bool{}, nil
	}
	return bookedSlotMapFromRows(clientID, rows), nil
}

func (u *AppointmentUsecase) GetBookedTimes(clientID int, date string) map[string]bool {
	m, _ := u.getBookedTimes(clientID, date)
	return m
}

func (u *AppointmentUsecase) CanRescheduleAppointmentByID(
	id int,
	customerPhone string,
	newTime string,
) error {

	phone := utils.NormalizeCustomerPhone(customerPhone)

	date, clientID, svc, err := repository.GetAppointmentForReschedule(id, phone)
	if err != nil {
		return fmt.Errorf("agendamento não encontrado")
	}

	duration := repository.GetServiceDuration(clientID, svc)

	layout := "15:04"

	start, err := time.Parse(layout, newTime)
	if err != nil {
		return fmt.Errorf("horário inválido")
	}

	workingStart, workingEnd, _, err := repository.GetWorkingHours(clientID)
	if err != nil {
		return fmt.Errorf("horário de funcionamento não configurado")
	}

	startWork, err := time.Parse(layout, workingStart)
	if err != nil {
		return fmt.Errorf("horário de abertura inválido")
	}

	endWork, err := time.Parse(layout, workingEnd)
	if err != nil {
		return fmt.Errorf("horário de fechamento inválido")
	}

	if start.Before(startWork) || start.Add(time.Duration(duration)*time.Minute).After(endWork) {
		return fmt.Errorf("horário fora do funcionamento")
	}

	validStart := false
	for _, slot := range generateTimeSlots(clientID) {
		if strings.TrimSpace(slot) == strings.TrimSpace(newTime) {
			validStart = true
			break
		}
	}

	if !validStart {
		return fmt.Errorf("horário inválido")
	}

	rows, err := repository.ListAppointmentSlotsForDateExcept(clientID, date, id)
	if err != nil {
		return err
	}

	booked := bookedSlotMapFromRows(clientID, rows)

	slots := int(math.Ceil(float64(duration) / 30.0))

	for i := 0; i < slots; i++ {
		slot := start.Add(time.Duration(i*30) * time.Minute)

		if booked[slot.Format("15:04")] {
			return fmt.Errorf("horário já ocupado")
		}
	}

	return nil
}

func (u *AppointmentUsecase) GetAvailableSlotsForRescheduleByID(
	id int,
	customerPhone string,
) ([]string, string, error) {

	phone := utils.NormalizeCustomerPhone(customerPhone)

	date, clientID, svc, err := repository.GetAppointmentForReschedule(id, phone)
	if err != nil {
		return nil, "", fmt.Errorf("agendamento não encontrado")
	}

	duration := repository.GetServiceDuration(clientID, svc)

	rows, err := repository.ListAppointmentSlotsForDateExcept(clientID, date, id)
	if err != nil {
		return nil, "", err
	}

	booked := bookedSlotMapFromRows(clientID, rows)

	_, workingEnd, _, err := repository.GetWorkingHours(clientID)
	if err != nil {
		return nil, "", fmt.Errorf("horário de funcionamento não configurado")
	}

	layout := "15:04"

	endWork, err := time.Parse(layout, workingEnd)
	if err != nil {
		return nil, "", fmt.Errorf("horário de fechamento inválido")
	}

	slots := generateTimeSlots(clientID)

	var available []string

	slotsNeeded := int(math.Ceil(float64(duration) / 30.0))

	for _, s := range slots {
		start, err := time.Parse(layout, s)
		if err != nil {
			continue
		}

		if start.Add(time.Duration(duration) * time.Minute).After(endWork) {
			continue
		}

		hasConflict := false

		for i := 0; i < slotsNeeded; i++ {
			slot := start.Add(time.Duration(i*30) * time.Minute)

			if booked[slot.Format("15:04")] {
				hasConflict = true
				break
			}
		}

		if !hasConflict {
			available = append(available, s)
		}
	}

	return available, date, nil
}

func (u *AppointmentUsecase) RescheduleAppointmentByID(
	id int,
	customerPhone string,
	newTime string,
) error {

	if err := u.CanRescheduleAppointmentByID(id, customerPhone, newTime); err != nil {
		return err
	}

	return repository.UpdateAppointmentTimeByID(id, newTime)
}

func (u *AppointmentUsecase) CancelAppointment(
	id int,
	clientID int,
	customerPhone string,
) error {

	phone := utils.NormalizeCustomerPhone(customerPhone)

	n, err := repository.DeleteCustomerAppointment(id, clientID, phone)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("agendamento não encontrado")
	}

	return nil
}

func (u *AppointmentUsecase) GetAvailableSlots(clientID int, date string) ([]string, error) {

	slots := generateTimeSlots(clientID)
	booked := u.GetBookedTimes(clientID, date)

	var available []string

	for _, s := range slots {
		if !booked[s] {
			available = append(available, s)
		}
	}

	return available, nil
}

func generateTimeSlots(clientID int) []string {
	start, end, interval, err := repository.GetWorkingHours(clientID)
	if err != nil {
		return []string{}
	}

	layout := "15:04"
	startTime, _ := time.Parse(layout, start)
	endTime, _ := time.Parse(layout, end)

	var slots []string

	for t := startTime; t.Before(endTime); t = t.Add(time.Duration(interval) * time.Minute) {
		slots = append(slots, t.Format("15:04"))
	}

	return slots
}

func (u *AppointmentUsecase) GetAppointmentsByCustomerPhone(
	clientID int,
	customerPhone string,
) ([]model.Appointment, error) {

	phone := utils.NormalizeCustomerPhone(customerPhone)

	rows, err := database.DB.Query(`
		SELECT id, name, service, date, time
		FROM appointments
		WHERE client_id = $1
		  AND customer_phone = $2
		  AND date::date >= CURRENT_DATE
		ORDER BY date ASC, time ASC
	`, clientID, phone)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Appointment

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
