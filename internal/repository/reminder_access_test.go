package repository

import (
	"flitta/internal/database"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetAllPendingRemindersFiltersBlockedClients(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro ao criar sqlmock: %v", err)
	}
	defer db.Close()

	previousDB := database.DB
	database.DB = db
	defer func() {
		database.DB = previousDB
	}()

	mock.ExpectQuery(
		`(?s)c\.status = 'active'.*c\.pilot_expires_at IS NULL.*c\.pilot_expires_at > NOW\(\)`,
	).
		WithArgs(24).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"client_id",
					"name",
					"service",
					"date",
					"time",
					"customer_phone",
					"reminder_sent",
					"company_name",
				},
			),
		)

	reminders, err :=
		GetAllPendingReminderAppointments(24)

	if err != nil {
		t.Fatalf(
			"erro inesperado: %v",
			err,
		)
	}

	if len(reminders) != 0 {
		t.Fatalf(
			"esperado 0 lembretes, recebido %d",
			len(reminders),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}
