package service

import (
	"flitta/internal/model"
	"flitta/internal/repository"
	"fmt"
	"log"
	"strings"
	"time"
)

type ReminderScheduler struct {
	CheckInterval time.Duration
	HoursBefore   int
}

func NewReminderScheduler(checkInterval time.Duration, hoursBefore int) *ReminderScheduler {
	return &ReminderScheduler{
		CheckInterval: checkInterval,
		HoursBefore:   hoursBefore,
	}
}

func (s *ReminderScheduler) Start() {
	log.Printf("Scheduler de lembretes iniciado: verificação a cada %s, janela de %d horas", s.CheckInterval, s.HoursBefore)

	go func() {
		s.ProcessPendingReminders()

		ticker := time.NewTicker(s.CheckInterval)
		defer ticker.Stop()

		for range ticker.C {
			s.ProcessPendingReminders()
		}
	}()
}

func (s *ReminderScheduler) ProcessPendingReminders() {
	reminders, err := repository.GetAllPendingReminderAppointments(s.HoursBefore)
	if err != nil {
		log.Println("Erro ao buscar lembretes automáticos pendentes:", err)
		return
	}

	if len(reminders) == 0 {
		log.Println("Scheduler: nenhum lembrete pendente encontrado")
		return
	}

	for _, reminder := range reminders {
		message := BuildScheduledReminderMessage(reminder.CompanyName, reminder.Appointment)

		_, err := SendAppointmentReminder(
			reminder.CustomerPhone,
			message,
			formatScheduledReminderDate(reminder.Date),
			reminder.Time,
		)
		if err != nil {
			log.Printf("Erro ao enviar lembrete automático do agendamento %d: %v", reminder.ID, err)
			continue
		}

		err = repository.MarkReminderAsSent(reminder.ID, reminder.ClientID)
		if err != nil {
			log.Printf("Lembrete do agendamento %d enviado, mas não foi possível atualizar status: %v", reminder.ID, err)
			continue
		}

		log.Printf("Lembrete automático processado com sucesso: appointment_id=%d client_id=%d", reminder.ID, reminder.ClientID)
	}
}

func BuildScheduledReminderMessage(companyName string, appointment model.Appointment) string {
	return fmt.Sprintf(
		"Olá, %s! 😊\nPassando para lembrar do seu horário em %s.\n\n📌 Serviço: %s\n📅 Data: %s\n🕒 Horário: %s\n\nSe precisar remarcar ou cancelar, responda esta mensagem.",
		appointment.Name,
		companyName,
		appointment.Service,
		formatScheduledReminderDate(appointment.Date),
		appointment.Time,
	)
}

func formatScheduledReminderDate(date string) string {
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		return date
	}

	return parts[2] + "/" + parts[1] + "/" + parts[0]
}
