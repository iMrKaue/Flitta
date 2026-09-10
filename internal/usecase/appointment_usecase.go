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

func appointmentIntervalsOverlap(
	startA string,
	durationA int,
	startB string,
	durationB int,
) (bool, error) {
	layout := "15:04"

	aStart, err := time.Parse(layout, startA)
	if err != nil {
		return false, err
	}

	bStart, err := time.Parse(layout, startB)
	if err != nil {
		return false, err
	}

	if durationA <= 0 || durationB <= 0 {
		return false, fmt.Errorf("duração inválida")
	}

	aEnd := aStart.Add(
		time.Duration(durationA) * time.Minute,
	)

	bEnd := bStart.Add(
		time.Duration(durationB) * time.Minute,
	)

	return aStart.Before(bEnd) &&
		aEnd.After(bStart), nil
}

func hasProfessionalAppointmentConflict(
	start string,
	duration int,
	appointments []repository.AppointmentIntervalRow,
) (bool, error) {
	for _, appointment := range appointments {
		overlap, err := appointmentIntervalsOverlap(
			start,
			duration,
			appointment.Time,
			appointment.DurationMinutes,
		)
		if err != nil {
			return false, err
		}

		if overlap {
			return true, nil
		}
	}

	return false, nil
}

func appointmentFitsProfessionalWorkingHours(
	date string,
	start string,
	duration int,
	workingHours []model.ProfessionalWorkingHour,
) (bool, error) {
	if duration <= 0 {
		return false, fmt.Errorf("duração inválida")
	}

	appointmentStart, err := time.Parse(
		"2006-01-02 15:04",
		date+" "+start,
	)
	if err != nil {
		return false, err
	}

	appointmentEnd := appointmentStart.Add(
		time.Duration(duration) * time.Minute,
	)

	weekday := int(appointmentStart.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	for _, workingHour := range workingHours {
		if workingHour.Weekday != weekday {
			continue
		}

		workStart, err := time.Parse(
			"2006-01-02 15:04",
			date+" "+workingHour.Start,
		)
		if err != nil {
			return false, err
		}

		workEnd, err := time.Parse(
			"2006-01-02 15:04",
			date+" "+workingHour.End,
		)
		if err != nil {
			return false, err
		}

		startInside :=
			appointmentStart.Equal(workStart) ||
				appointmentStart.After(workStart)

		endInside :=
			appointmentEnd.Equal(workEnd) ||
				appointmentEnd.Before(workEnd)

		if startInside && endInside {
			return true, nil
		}
	}

	return false, nil
}

func appointmentFitsBusinessHours(
	start string,
	duration int,
	workingStart string,
	workingEnd string,
) (bool, error) {
	if duration <= 0 {
		return false, fmt.Errorf("duração inválida")
	}

	layout := "15:04"

	appointmentStart, err := time.Parse(layout, start)
	if err != nil {
		return false, err
	}

	businessStart, err := time.Parse(layout, workingStart)
	if err != nil {
		return false, err
	}

	businessEnd, err := time.Parse(layout, workingEnd)
	if err != nil {
		return false, err
	}

	appointmentEnd := appointmentStart.Add(
		time.Duration(duration) * time.Minute,
	)

	startInside :=
		appointmentStart.Equal(businessStart) ||
			appointmentStart.After(businessStart)

	endInside :=
		appointmentEnd.Equal(businessEnd) ||
			appointmentEnd.Before(businessEnd)

	return startInside && endInside, nil
}

func appointmentOverlapsProfessionalTimeOff(
	date string,
	start string,
	duration int,
	timeOffs []model.ProfessionalTimeOff,
) (bool, error) {
	if duration <= 0 {
		return false, fmt.Errorf("duração inválida")
	}

	layout := "2006-01-02T15:04"

	appointmentStart, err := time.Parse(
		layout,
		date+"T"+start,
	)
	if err != nil {
		return false, err
	}

	appointmentEnd := appointmentStart.Add(
		time.Duration(duration) * time.Minute,
	)

	for _, timeOff := range timeOffs {
		timeOffStart, err := time.Parse(
			layout,
			timeOff.StartAt,
		)
		if err != nil {
			return false, err
		}

		timeOffEnd, err := time.Parse(
			layout,
			timeOff.EndAt,
		)
		if err != nil {
			return false, err
		}

		if !timeOffStart.Before(timeOffEnd) {
			return false, fmt.Errorf(
				"período de indisponibilidade inválido",
			)
		}

		overlap :=
			appointmentStart.Before(timeOffEnd) &&
				appointmentEnd.After(timeOffStart)

		if overlap {
			return true, nil
		}
	}

	return false, nil
}

func (u *AppointmentUsecase) checkProfessionalAppointmentConflict(
	clientID int,
	professionalID int,
	date string,
	start string,
	duration int,
	exceptID int,
) (bool, error) {
	appointments, err :=
		repository.ListProfessionalAppointmentIntervalsForDate(
			clientID,
			professionalID,
			date,
			exceptID,
		)
	if err != nil {
		return false, err
	}

	return hasProfessionalAppointmentConflict(
		start,
		duration,
		appointments,
	)
}

func (u *AppointmentUsecase) checkProfessionalAvailability(
	clientID int,
	professionalID int,
	date string,
	start string,
	duration int,
	exceptID int,
) (bool, error) {
	if duration <= 0 {
		return false, fmt.Errorf("duração inválida")
	}

	if _, err := time.Parse(
		"2006-01-02 15:04",
		date+" "+start,
	); err != nil {
		return false, fmt.Errorf("data ou horário inválido")
	}

	// 1. Horário global do estabelecimento.
	businessStart, businessEnd, _, err :=
		repository.GetWorkingHours(clientID)
	if err != nil {
		return false, err
	}

	fitsBusiness, err := appointmentFitsBusinessHours(
		start,
		duration,
		businessStart,
		businessEnd,
	)
	if err != nil {
		return false, err
	}

	if !fitsBusiness {
		return false, nil
	}

	// 2. Jornada semanal profissional.
	workingHours, err :=
		repository.ListProfessionalWorkingHours(
			clientID,
			professionalID,
		)
	if err != nil {
		return false, err
	}

	fitsProfessional, err :=
		appointmentFitsProfessionalWorkingHours(
			date,
			start,
			duration,
			workingHours,
		)
	if err != nil {
		return false, err
	}

	if !fitsProfessional {
		return false, nil
	}

	// 3. Ausências, consultas, férias e folgas pontuais.
	timeOffs, err :=
		repository.ListProfessionalTimeOff(
			clientID,
			professionalID,
		)
	if err != nil {
		return false, err
	}

	hasTimeOff, err :=
		appointmentOverlapsProfessionalTimeOff(
			date,
			start,
			duration,
			timeOffs,
		)
	if err != nil {
		return false, err
	}

	if hasTimeOff {
		return false, nil
	}

	// 4. Outros appointments do profissional.
	hasConflict, err :=
		u.checkProfessionalAppointmentConflict(
			clientID,
			professionalID,
			date,
			start,
			duration,
			exceptID,
		)
	if err != nil {
		return false, err
	}

	if hasConflict {
		return false, nil
	}

	return true, nil
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

	n, err := repository.CancelCustomerAppointment(id, clientID, phone)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("agendamento não encontrado")
	}

	return nil
}

func (u *AppointmentUsecase) CompleteAppointment(
	id int,
	clientID int,
) error {
	rowsAffected, err := repository.UpdateAppointmentOutcome(
		id,
		clientID,
		"completed",
	)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agendamento não encontrado ou indisponível")
	}

	return nil
}

func (u *AppointmentUsecase) MarkAppointmentNoShow(
	id int,
	clientID int,
) error {
	rowsAffected, err := repository.UpdateAppointmentOutcome(
		id,
		clientID,
		"no_show",
	)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agendamento não encontrado ou indisponível")
	}

	return nil
}

func (u *AppointmentUsecase) GetAvailableSlotsForProfessionalService(
	clientID int,
	professionalID int,
	date string,
	serviceName string,
) ([]string, error) {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("data inválida")
	}

	// 1. Valida profissional + serviço e obtém a duração real.
	service, err := repository.GetProfessionalServiceForBooking(
		clientID,
		professionalID,
		serviceName,
	)
	if err != nil {
		return nil, err
	}

	// 2. Obtém a grade global de horários da empresa.
	businessStart, businessEnd, intervalMinutes, err :=
		repository.GetWorkingHours(clientID)
	if err != nil {
		return nil, err
	}

	candidates, err := generateTimeSlotsBetween(
		businessStart,
		businessEnd,
		intervalMinutes,
	)
	if err != nil {
		return nil, err
	}

	// 3. Carrega a jornada do profissional uma única vez.
	workingHours, err :=
		repository.ListProfessionalWorkingHours(
			clientID,
			professionalID,
		)
	if err != nil {
		return nil, err
	}

	// 4. Carrega as indisponibilidades uma única vez.
	timeOffs, err :=
		repository.ListProfessionalTimeOff(
			clientID,
			professionalID,
		)
	if err != nil {
		return nil, err
	}

	// 5. Carrega os appointments daquele dia uma única vez.
	appointments, err :=
		repository.ListProfessionalAppointmentIntervalsForDate(
			clientID,
			professionalID,
			date,
			0,
		)
	if err != nil {
		return nil, err
	}

	available := make([]string, 0)

	for _, candidate := range candidates {
		fitsBusiness, err := appointmentFitsBusinessHours(
			candidate,
			service.Duration,
			businessStart,
			businessEnd,
		)
		if err != nil {
			return nil, err
		}

		if !fitsBusiness {
			continue
		}

		fitsProfessional, err :=
			appointmentFitsProfessionalWorkingHours(
				date,
				candidate,
				service.Duration,
				workingHours,
			)
		if err != nil {
			return nil, err
		}

		if !fitsProfessional {
			continue
		}

		hasTimeOff, err :=
			appointmentOverlapsProfessionalTimeOff(
				date,
				candidate,
				service.Duration,
				timeOffs,
			)
		if err != nil {
			return nil, err
		}

		if hasTimeOff {
			continue
		}

		hasConflict, err :=
			hasProfessionalAppointmentConflict(
				candidate,
				service.Duration,
				appointments,
			)
		if err != nil {
			return nil, err
		}

		if hasConflict {
			continue
		}

		available = append(
			available,
			candidate,
		)
	}

	return available, nil
}

func (u *AppointmentUsecase) GetAvailableSlots(clientID int, date string) ([]string, error) {
	return u.GetAvailableSlotsForService(clientID, date, "")
}

func (u *AppointmentUsecase) GetAvailableSlotsForService(
	clientID int,
	date string,
	serviceName string,
) ([]string, error) {

	slots := generateTimeSlots(clientID)
	booked := u.GetBookedTimes(clientID, date)

	duration := 30
	if strings.TrimSpace(serviceName) != "" {
		duration = repository.GetServiceDuration(clientID, serviceName)
	}

	_, workingEnd, _, err := repository.GetWorkingHours(clientID)
	if err != nil {
		return []string{}, fmt.Errorf("horário de funcionamento não configurado")
	}

	layout := "15:04"

	endWork, err := time.Parse(layout, workingEnd)
	if err != nil {
		return []string{}, fmt.Errorf("horário de fechamento inválido")
	}

	slotsNeeded := int(math.Ceil(float64(duration) / 30.0))

	var available []string

	for _, s := range slots {
		start, err := time.Parse(layout, s)
		if err != nil {
			continue
		}

		serviceEnd := start.Add(time.Duration(duration) * time.Minute)

		if serviceEnd.After(endWork) {
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

	return available, nil
}

func generateTimeSlotsBetween(
	start string,
	end string,
	intervalMinutes int,
) ([]string, error) {
	if intervalMinutes <= 0 {
		return nil, fmt.Errorf("intervalo de agenda inválido")
	}

	layout := "15:04"

	startTime, err := time.Parse(layout, start)
	if err != nil {
		return nil, fmt.Errorf("horário inicial inválido")
	}

	endTime, err := time.Parse(layout, end)
	if err != nil {
		return nil, fmt.Errorf("horário final inválido")
	}

	if !startTime.Before(endTime) {
		return nil, fmt.Errorf("horário de funcionamento inválido")
	}

	slots := []string{}

	for current := startTime; current.Before(endTime); current = current.Add(
		time.Duration(intervalMinutes) * time.Minute,
	) {
		slots = append(
			slots,
			current.Format(layout),
		)
	}

	return slots, nil
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
		  AND date::date >= (NOW() AT TIME ZONE 'America/Sao_Paulo')::date
		  AND status IN ('scheduled', 'confirmed')
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
