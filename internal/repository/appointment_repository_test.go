package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListProfessionalAppointmentIntervalsForDateSuccess(
	t *testing.T,
) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectQuery(
		`(?s)SELECT.*id,.*professional_id,.*time,.*duration_snapshot.*FROM appointments.*client_id = \$1.*professional_id = \$2.*date = \$3.*status\s+IN\s+\('scheduled',\s*'confirmed'\).*id <> \$4`,
	).
		WithArgs(
			9,
			5,
			"2026-09-20",
			0,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"professional_id",
				"time",
				"duration_snapshot",
			}).
				AddRow(
					101,
					5,
					"10:00",
					90,
				).
				AddRow(
					102,
					5,
					"14:00",
					30,
				),
		)

	appointments, err :=
		ListProfessionalAppointmentIntervalsForDate(
			9,
			5,
			"2026-09-20",
			0,
		)

	if err != nil {
		t.Fatalf(
			"esperava sucesso, recebeu erro: %v",
			err,
		)
	}

	if len(appointments) != 2 {
		t.Fatalf(
			"esperava 2 agendamentos, recebeu %d",
			len(appointments),
		)
	}

	if appointments[0].ID != 101 {
		t.Errorf(
			"esperava primeiro ID 101, recebeu %d",
			appointments[0].ID,
		)
	}

	if appointments[0].ProfessionalID != 5 {
		t.Errorf(
			"esperava professional_id 5, recebeu %d",
			appointments[0].ProfessionalID,
		)
	}

	if appointments[0].Time != "10:00" {
		t.Errorf(
			"esperava horário 10:00, recebeu %s",
			appointments[0].Time,
		)
	}

	if appointments[0].DurationMinutes != 90 {
		t.Errorf(
			"esperava duração 90, recebeu %d",
			appointments[0].DurationMinutes,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestListProfessionalAppointmentIntervalsForDateExceptID(
	t *testing.T,
) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectQuery(
		`(?s)SELECT.*id,.*professional_id,.*time,.*duration_snapshot.*FROM appointments.*client_id = \$1.*professional_id = \$2.*date = \$3.*status\s+IN\s+\('scheduled',\s*'confirmed'\).*id <> \$4`,
	).
		WithArgs(
			9,
			5,
			"2026-09-20",
			101,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"professional_id",
				"time",
				"duration_snapshot",
			}).
				AddRow(
					102,
					5,
					"14:00",
					30,
				),
		)

	appointments, err :=
		ListProfessionalAppointmentIntervalsForDate(
			9,
			5,
			"2026-09-20",
			101,
		)

	if err != nil {
		t.Fatalf(
			"esperava sucesso, recebeu erro: %v",
			err,
		)
	}

	if len(appointments) != 1 {
		t.Fatalf(
			"esperava 1 agendamento, recebeu %d",
			len(appointments),
		)
	}

	if appointments[0].ID != 102 {
		t.Errorf(
			"esperava ID 102, recebeu %d",
			appointments[0].ID,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestListProfessionalAppointmentIntervalsForDateIncludesLegacyAppointment(
	t *testing.T,
) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectQuery(
		`(?s)SELECT.*COALESCE\(professional_id,\s*0\).*FROM appointments.*client_id = \$1.*professional_id = \$2.*professional_id IS NULL.*date = \$3.*status\s+IN\s+\('scheduled',\s*'confirmed'\).*id <> \$4`,
	).
		WithArgs(
			9,
			5,
			"2026-09-20",
			0,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"professional_id",
				"time",
				"duration_snapshot",
			}).AddRow(
				200,
				0,
				"10:00",
				60,
			),
		)

	appointments, err :=
		ListProfessionalAppointmentIntervalsForDate(
			9,
			5,
			"2026-09-20",
			0,
		)

	if err != nil {
		t.Fatalf(
			"esperava sucesso, recebeu erro: %v",
			err,
		)
	}

	if len(appointments) != 1 {
		t.Fatalf(
			"esperava 1 agendamento legado, recebeu %d",
			len(appointments),
		)
	}

	if appointments[0].ID != 200 {
		t.Fatalf(
			"esperava ID 200, recebeu %d",
			appointments[0].ID,
		)
	}

	if appointments[0].ProfessionalID != 0 {
		t.Fatalf(
			"esperava professional_id 0 para legado, recebeu %d",
			appointments[0].ProfessionalID,
		)
	}

	if appointments[0].DurationMinutes != 60 {
		t.Fatalf(
			"esperava duração 60, recebeu %d",
			appointments[0].DurationMinutes,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}
