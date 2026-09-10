package usecase

import (
	"flitta/internal/database"
	"flitta/internal/model"
	"flitta/internal/repository"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAppointmentIntervalsOverlap(t *testing.T) {
	tests := []struct {
		name      string
		startA    string
		durationA int
		startB    string
		durationB int
		want      bool
	}{
		{
			name:      "partial overlap",
			startA:    "10:00",
			durationA: 90,
			startB:    "11:00",
			durationB: 60,
			want:      true,
		},
		{
			name:      "adjacent intervals",
			startA:    "10:00",
			durationA: 90,
			startB:    "11:30",
			durationB: 30,
			want:      false,
		},
		{
			name:      "short appointment overlaps",
			startA:    "10:00",
			durationA: 45,
			startB:    "10:30",
			durationB: 30,
			want:      true,
		},
		{
			name:      "new interval contains existing",
			startA:    "09:00",
			durationA: 180,
			startB:    "10:00",
			durationB: 30,
			want:      true,
		},
		{
			name:      "separate intervals",
			startA:    "09:00",
			durationA: 60,
			startB:    "11:00",
			durationB: 60,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := appointmentIntervalsOverlap(
				tt.startA,
				tt.durationA,
				tt.startB,
				tt.durationB,
			)

			if err != nil {
				t.Fatalf(
					"não esperava erro: %v",
					err,
				)
			}

			if got != tt.want {
				t.Fatalf(
					"esperava overlap=%v, recebeu %v",
					tt.want,
					got,
				)
			}
		})
	}
}

func TestHasProfessionalAppointmentConflict(t *testing.T) {
	appointments := []repository.AppointmentIntervalRow{
		{
			ID:              101,
			ProfessionalID:  5,
			Time:            "10:00",
			DurationMinutes: 90,
		},
		{
			ID:              102,
			ProfessionalID:  5,
			Time:            "14:00",
			DurationMinutes: 30,
		},
	}

	tests := []struct {
		name     string
		start    string
		duration int
		want     bool
	}{
		{
			name:     "overlaps first appointment",
			start:    "11:00",
			duration: 60,
			want:     true,
		},
		{
			name:     "starts when first appointment ends",
			start:    "11:30",
			duration: 30,
			want:     false,
		},
		{
			name:     "contains second appointment",
			start:    "13:30",
			duration: 90,
			want:     true,
		},
		{
			name:     "free between appointments",
			start:    "12:00",
			duration: 60,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hasProfessionalAppointmentConflict(
				tt.start,
				tt.duration,
				appointments,
			)

			if err != nil {
				t.Fatalf(
					"não esperava erro: %v",
					err,
				)
			}

			if got != tt.want {
				t.Fatalf(
					"esperava conflito=%v, recebeu %v",
					tt.want,
					got,
				)
			}
		})
	}
}

func TestCheckProfessionalAppointmentConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"erro ao criar mock do banco: %v",
			err,
		)
	}
	defer db.Close()

	previousDB := database.DB
	database.DB = db

	t.Cleanup(func() {
		database.DB = previousDB
	})

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
				),
		)

	usecase := NewAppointmentUsecase()

	conflict, err :=
		usecase.checkProfessionalAppointmentConflict(
			9,
			5,
			"2026-09-20",
			"11:00",
			60,
			0,
		)

	if err != nil {
		t.Fatalf(
			"não esperava erro: %v",
			err,
		)
	}

	if !conflict {
		t.Fatal(
			"esperava conflito de agendamento",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas do banco não atendidas: %v",
			err,
		)
	}
}

func TestAppointmentFitsProfessionalWorkingHours(t *testing.T) {
	workingHours := []model.ProfessionalWorkingHour{
		{
			Weekday: 1,
			Start:   "09:00",
			End:     "12:00",
		},
		{
			Weekday: 1,
			Start:   "13:00",
			End:     "18:00",
		},
	}

	tests := []struct {
		name     string
		date     string
		start    string
		duration int
		want     bool
	}{
		{
			name:     "fits morning range",
			date:     "2026-09-21",
			start:    "10:30",
			duration: 90,
			want:     true,
		},
		{
			name:     "crosses lunch break",
			date:     "2026-09-21",
			start:    "11:00",
			duration: 90,
			want:     false,
		},
		{
			name:     "fits afternoon range",
			date:     "2026-09-21",
			start:    "13:00",
			duration: 60,
			want:     true,
		},
		{
			name:     "starts during lunch break",
			date:     "2026-09-21",
			start:    "12:00",
			duration: 30,
			want:     false,
		},
		{
			name:     "different weekday",
			date:     "2026-09-22",
			start:    "10:00",
			duration: 30,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err :=
				appointmentFitsProfessionalWorkingHours(
					tt.date,
					tt.start,
					tt.duration,
					workingHours,
				)

			if err != nil {
				t.Fatalf(
					"não esperava erro: %v",
					err,
				)
			}

			if got != tt.want {
				t.Fatalf(
					"esperava fits=%v, recebeu %v",
					tt.want,
					got,
				)
			}
		})
	}
}

func TestAppointmentOverlapsProfessionalTimeOff(t *testing.T) {
	timeOffs := []model.ProfessionalTimeOff{
		{
			ID:             1,
			ProfessionalID: 5,
			StartAt:        "2026-09-21T14:00",
			EndAt:          "2026-09-21T16:00",
			Reason:         "Consulta",
		},
		{
			ID:             2,
			ProfessionalID: 5,
			StartAt:        "2026-09-25T00:00",
			EndAt:          "2026-09-28T00:00",
			Reason:         "Férias",
		},
	}

	tests := []struct {
		name     string
		date     string
		start    string
		duration int
		want     bool
	}{
		{
			name:     "overlaps consultation",
			date:     "2026-09-21",
			start:    "15:00",
			duration: 30,
			want:     true,
		},
		{
			name:     "ends when time off starts",
			date:     "2026-09-21",
			start:    "13:00",
			duration: 60,
			want:     false,
		},
		{
			name:     "starts when time off ends",
			date:     "2026-09-21",
			start:    "16:00",
			duration: 60,
			want:     false,
		},
		{
			name:     "crosses into consultation",
			date:     "2026-09-21",
			start:    "13:30",
			duration: 60,
			want:     true,
		},
		{
			name:     "inside vacation",
			date:     "2026-09-26",
			start:    "10:00",
			duration: 60,
			want:     true,
		},
		{
			name:     "after vacation",
			date:     "2026-09-28",
			start:    "10:00",
			duration: 60,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err :=
				appointmentOverlapsProfessionalTimeOff(
					tt.date,
					tt.start,
					tt.duration,
					timeOffs,
				)

			if err != nil {
				t.Fatalf(
					"não esperava erro: %v",
					err,
				)
			}

			if got != tt.want {
				t.Fatalf(
					"esperava overlap=%v, recebeu %v",
					tt.want,
					got,
				)
			}
		})
	}
}

func TestAppointmentFitsBusinessHours(t *testing.T) {
	tests := []struct {
		name     string
		start    string
		duration int
		want     bool
	}{
		{
			name:     "fits business hours",
			start:    "10:00",
			duration: 60,
			want:     true,
		},
		{
			name:     "starts exactly when business opens",
			start:    "09:00",
			duration: 60,
			want:     true,
		},
		{
			name:     "ends exactly when business closes",
			start:    "17:00",
			duration: 60,
			want:     true,
		},
		{
			name:     "starts before business opens",
			start:    "08:30",
			duration: 60,
			want:     false,
		},
		{
			name:     "ends after business closes",
			start:    "17:30",
			duration: 60,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := appointmentFitsBusinessHours(
				tt.start,
				tt.duration,
				"09:00",
				"18:00",
			)

			if err != nil {
				t.Fatalf(
					"não esperava erro: %v",
					err,
				)
			}

			if got != tt.want {
				t.Fatalf(
					"esperava fits=%v, recebeu %v",
					tt.want,
					got,
				)
			}
		})
	}
}

func TestCheckProfessionalAvailabilitySuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"erro ao criar mock do banco: %v",
			err,
		)
	}
	defer db.Close()

	previousDB := database.DB
	database.DB = db

	t.Cleanup(func() {
		database.DB = previousDB
	})

	// Empresa: 09:00–18:00.
	mock.ExpectQuery(
		`(?s)SELECT.*start_time.*end_time.*interval_minutes.*FROM working_hours.*client_id = \$1`,
	).
		WithArgs(9).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"start_time",
				"end_time",
				"interval_minutes",
			}).AddRow(
				"09:00",
				"18:00",
				30,
			),
		)

	// Profissional pertence ao tenant.
	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*id = \$1.*client_id = \$2`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	// Segunda-feira: 09:00–12:00 e 13:00–18:00.
	mock.ExpectQuery(
		`(?s)FROM professional_working_hours.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			9,
			5,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"weekday",
				"start",
				"end",
			}).
				AddRow(
					1,
					1,
					"09:00",
					"12:00",
				).
				AddRow(
					2,
					1,
					"13:00",
					"18:00",
				),
		)

	// Validação de tenant feita pelo repository de time_off.
	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*id = \$1.*client_id = \$2`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	// Nenhuma indisponibilidade.
	mock.ExpectQuery(
		`(?s)FROM professional_time_off.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			9,
			5,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"client_id",
				"professional_id",
				"start_at",
				"end_at",
				"reason",
			}),
		)

	// Nenhum appointment conflitante.
	mock.ExpectQuery(
		`(?s)SELECT.*id,.*professional_id,.*time,.*duration_snapshot.*FROM appointments.*client_id = \$1.*professional_id = \$2.*date = \$3`,
	).
		WithArgs(
			9,
			5,
			"2026-09-21",
			0,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"professional_id",
				"time",
				"duration_snapshot",
			}),
		)

	usecase := NewAppointmentUsecase()

	available, err :=
		usecase.checkProfessionalAvailability(
			9,
			5,
			"2026-09-21",
			"10:00",
			60,
			0,
		)

	if err != nil {
		t.Fatalf(
			"não esperava erro: %v",
			err,
		)
	}

	if !available {
		t.Fatal(
			"esperava profissional disponível",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas do banco não atendidas: %v",
			err,
		)
	}
}

func TestCheckProfessionalAvailabilityOutsideBusinessHours(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"erro ao criar mock do banco: %v",
			err,
		)
	}
	defer db.Close()

	previousDB := database.DB
	database.DB = db

	t.Cleanup(func() {
		database.DB = previousDB
	})

	mock.ExpectQuery(
		`(?s)SELECT.*start_time.*end_time.*interval_minutes.*FROM working_hours.*client_id = \$1`,
	).
		WithArgs(9).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"start_time",
				"end_time",
				"interval_minutes",
			}).AddRow(
				"09:00",
				"18:00",
				30,
			),
		)

	usecase := NewAppointmentUsecase()

	available, err :=
		usecase.checkProfessionalAvailability(
			9,
			5,
			"2026-09-21",
			"08:30",
			60,
			0,
		)

	if err != nil {
		t.Fatalf(
			"não esperava erro: %v",
			err,
		)
	}

	if available {
		t.Fatal(
			"esperava profissional indisponível fora do horário da empresa",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas do banco não atendidas: %v",
			err,
		)
	}
}

func TestCheckProfessionalAvailabilityOutsideProfessionalHours(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"erro ao criar mock do banco: %v",
			err,
		)
	}
	defer db.Close()

	previousDB := database.DB
	database.DB = db

	t.Cleanup(func() {
		database.DB = previousDB
	})

	// Empresa aberta das 09:00 às 18:00.
	mock.ExpectQuery(
		`(?s)SELECT.*start_time.*end_time.*interval_minutes.*FROM working_hours.*client_id = \$1`,
	).
		WithArgs(9).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"start_time",
				"end_time",
				"interval_minutes",
			}).AddRow(
				"09:00",
				"18:00",
				30,
			),
		)

	// Profissional pertence ao estabelecimento.
	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*id = \$1.*client_id = \$2`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	// Profissional trabalha 09:00–12:00 e 13:00–18:00.
	mock.ExpectQuery(
		`(?s)FROM professional_working_hours.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			9,
			5,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"weekday",
				"start",
				"end",
			}).
				AddRow(
					1,
					1,
					"09:00",
					"12:00",
				).
				AddRow(
					2,
					1,
					"13:00",
					"18:00",
				),
		)

	usecase := NewAppointmentUsecase()

	available, err :=
		usecase.checkProfessionalAvailability(
			9,
			5,
			"2026-09-21",
			"12:00",
			30,
			0,
		)

	if err != nil {
		t.Fatalf(
			"não esperava erro: %v",
			err,
		)
	}

	if available {
		t.Fatal(
			"esperava profissional indisponível fora da própria jornada",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas do banco não atendidas: %v",
			err,
		)
	}
}

func TestCheckProfessionalAvailabilityBlockedByTimeOff(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"erro ao criar mock do banco: %v",
			err,
		)
	}
	defer db.Close()

	previousDB := database.DB
	database.DB = db

	t.Cleanup(func() {
		database.DB = previousDB
	})

	// Empresa: 09:00–18:00.
	mock.ExpectQuery(
		`(?s)SELECT.*start_time.*end_time.*interval_minutes.*FROM working_hours.*client_id = \$1`,
	).
		WithArgs(9).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"start_time",
				"end_time",
				"interval_minutes",
			}).AddRow(
				"09:00",
				"18:00",
				30,
			),
		)

	// Profissional pertence ao tenant.
	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*id = \$1.*client_id = \$2`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	// Profissional trabalha nesse horário.
	mock.ExpectQuery(
		`(?s)FROM professional_working_hours.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			9,
			5,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"weekday",
				"start",
				"end",
			}).
				AddRow(
					1,
					1,
					"09:00",
					"12:00",
				).
				AddRow(
					2,
					1,
					"13:00",
					"18:00",
				),
		)

	// Validação do profissional no repository de time_off.
	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*id = \$1.*client_id = \$2`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	// Consulta das 14:00 às 16:00.
	mock.ExpectQuery(
		`(?s)FROM professional_time_off.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			9,
			5,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"client_id",
				"professional_id",
				"start_at",
				"end_at",
				"reason",
			}).AddRow(
				1,
				9,
				5,
				"2026-09-21T14:00",
				"2026-09-21T16:00",
				"Consulta",
			),
		)

	usecase := NewAppointmentUsecase()

	available, err :=
		usecase.checkProfessionalAvailability(
			9,
			5,
			"2026-09-21",
			"15:00",
			30,
			0,
		)

	if err != nil {
		t.Fatalf(
			"não esperava erro: %v",
			err,
		)
	}

	if available {
		t.Fatal(
			"esperava profissional indisponível por time_off",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas do banco não atendidas: %v",
			err,
		)
	}
}

func TestCheckProfessionalAvailabilityBlockedByAppointment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"erro ao criar mock do banco: %v",
			err,
		)
	}
	defer db.Close()

	previousDB := database.DB
	database.DB = db

	t.Cleanup(func() {
		database.DB = previousDB
	})

	// Empresa: 09:00–18:00.
	mock.ExpectQuery(
		`(?s)SELECT.*start_time.*end_time.*interval_minutes.*FROM working_hours.*client_id = \$1`,
	).
		WithArgs(9).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"start_time",
				"end_time",
				"interval_minutes",
			}).AddRow(
				"09:00",
				"18:00",
				30,
			),
		)

	// Profissional pertence ao tenant.
	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*id = \$1.*client_id = \$2`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	// Profissional trabalha nesse horário.
	mock.ExpectQuery(
		`(?s)FROM professional_working_hours.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			9,
			5,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"weekday",
				"start",
				"end",
			}).
				AddRow(
					1,
					1,
					"09:00",
					"12:00",
				).
				AddRow(
					2,
					1,
					"13:00",
					"18:00",
				),
		)

	// Validação do profissional no repository de time_off.
	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*id = \$1.*client_id = \$2`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	// Nenhum time_off.
	mock.ExpectQuery(
		`(?s)FROM professional_time_off.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			9,
			5,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"client_id",
				"professional_id",
				"start_at",
				"end_at",
				"reason",
			}),
		)

	// Appointment já existente: 10:00–11:30.
	mock.ExpectQuery(
		`(?s)SELECT.*id,.*professional_id,.*time,.*duration_snapshot.*FROM appointments.*client_id = \$1.*professional_id = \$2.*date = \$3.*status\s+IN\s+\('scheduled',\s*'confirmed'\).*id <> \$4`,
	).
		WithArgs(
			9,
			5,
			"2026-09-21",
			0,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"professional_id",
				"time",
				"duration_snapshot",
			}).AddRow(
				101,
				5,
				"10:00",
				90,
			),
		)

	usecase := NewAppointmentUsecase()

	available, err :=
		usecase.checkProfessionalAvailability(
			9,
			5,
			"2026-09-21",
			"11:00",
			60,
			0,
		)

	if err != nil {
		t.Fatalf(
			"não esperava erro: %v",
			err,
		)
	}

	if available {
		t.Fatal(
			"esperava profissional indisponível por conflito com outro appointment",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas do banco não atendidas: %v",
			err,
		)
	}
}

func TestCheckProfessionalAvailabilityBlockedByLegacyAppointment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"erro ao criar mock do banco: %v",
			err,
		)
	}
	defer db.Close()

	previousDB := database.DB
	database.DB = db

	t.Cleanup(func() {
		database.DB = previousDB
	})

	// Empresa: 09:00–18:00.
	mock.ExpectQuery(
		`(?s)SELECT.*start_time.*end_time.*interval_minutes.*FROM working_hours.*client_id = \$1`,
	).
		WithArgs(9).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"start_time",
				"end_time",
				"interval_minutes",
			}).AddRow(
				"09:00",
				"18:00",
				30,
			),
		)

	// Profissional pertence ao estabelecimento.
	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*id = \$1.*client_id = \$2`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	// Jornada do profissional.
	mock.ExpectQuery(
		`(?s)FROM professional_working_hours.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			9,
			5,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"weekday",
				"start",
				"end",
			}).
				AddRow(
					1,
					1,
					"09:00",
					"12:00",
				).
				AddRow(
					2,
					1,
					"13:00",
					"18:00",
				),
		)

	// Validação do profissional para time_off.
	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*id = \$1.*client_id = \$2`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	// Nenhum time_off.
	mock.ExpectQuery(
		`(?s)FROM professional_time_off.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			9,
			5,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"client_id",
				"professional_id",
				"start_at",
				"end_at",
				"reason",
			}),
		)

	// Appointment legado sem professional_id:
	// 10:00–11:00.
	mock.ExpectQuery(
		`(?s)SELECT.*COALESCE\(professional_id,\s*0\).*FROM appointments.*client_id = \$1.*professional_id = \$2.*professional_id IS NULL.*date = \$3.*status\s+IN\s+\('scheduled',\s*'confirmed'\).*id <> \$4`,
	).
		WithArgs(
			9,
			5,
			"2026-09-21",
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

	usecase := NewAppointmentUsecase()

	available, err :=
		usecase.checkProfessionalAvailability(
			9,
			5,
			"2026-09-21",
			"10:30",
			30,
			0,
		)

	if err != nil {
		t.Fatalf(
			"não esperava erro: %v",
			err,
		)
	}

	if available {
		t.Fatal(
			"esperava profissional indisponível por conflito com appointment legado",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas do banco não atendidas: %v",
			err,
		)
	}
}

func TestGenerateTimeSlotsBetween(t *testing.T) {
	slots, err := generateTimeSlotsBetween(
		"09:00",
		"11:00",
		30,
	)

	if err != nil {
		t.Fatalf(
			"não esperava erro: %v",
			err,
		)
	}

	expected := []string{
		"09:00",
		"09:30",
		"10:00",
		"10:30",
	}

	if len(slots) != len(expected) {
		t.Fatalf(
			"esperava %d horários, recebeu %d",
			len(expected),
			len(slots),
		)
	}

	for i := range expected {
		if slots[i] != expected[i] {
			t.Fatalf(
				"posição %d: esperava %s, recebeu %s",
				i,
				expected[i],
				slots[i],
			)
		}
	}
}

func TestGenerateTimeSlotsBetweenInvalidInterval(t *testing.T) {
	_, err := generateTimeSlotsBetween(
		"09:00",
		"18:00",
		0,
	)

	if err == nil {
		t.Fatal(
			"esperava erro para intervalo inválido",
		)
	}
}

func TestGetAvailableSlotsForProfessionalService(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"erro ao criar mock do banco: %v",
			err,
		)
	}
	defer db.Close()

	previousDB := database.DB
	database.DB = db

	t.Cleanup(func() {
		database.DB = previousDB
	})

	// 1. Profissional existe e está ativo.
	mock.ExpectQuery(
		`(?s)SELECT active.*FROM professionals.*id = \$1.*client_id = \$2`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"active"},
			).AddRow(true),
		)

	// 2. Serviço Corte está atribuído ao profissional e dura 90 minutos.
	mock.ExpectQuery(
		`(?s)SELECT.*s.id,.*s.name,.*s.duration,.*s.price.*FROM professional_services ps.*INNER JOIN services s.*ps.client_id = \$1.*ps.professional_id = \$2.*LOWER\(BTRIM\(s.name\)\).*LOWER\(BTRIM\(\$3\)\)`,
	).
		WithArgs(
			9,
			5,
			"Corte",
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"name",
				"duration",
				"price",
			}).AddRow(
				23,
				"Corte",
				90,
				75.00,
			),
		)

	// 3. Empresa: 09:00–18:00, grade de 30 minutos.
	mock.ExpectQuery(
		`(?s)SELECT.*start_time.*end_time.*interval_minutes.*FROM working_hours.*client_id = \$1`,
	).
		WithArgs(9).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"start_time",
				"end_time",
				"interval_minutes",
			}).AddRow(
				"09:00",
				"18:00",
				30,
			),
		)

	// 4. Validação tenant do profissional para a jornada.
	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*id = \$1.*client_id = \$2`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	// Segunda: 09:00–12:00 e 13:00–18:00.
	mock.ExpectQuery(
		`(?s)FROM professional_working_hours.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			9,
			5,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"weekday",
				"start",
				"end",
			}).
				AddRow(
					1,
					1,
					"09:00",
					"12:00",
				).
				AddRow(
					2,
					1,
					"13:00",
					"18:00",
				),
		)

	// 5. Validação tenant para time_off.
	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*id = \$1.*client_id = \$2`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	// Time off: 15:00–16:00.
	mock.ExpectQuery(
		`(?s)FROM professional_time_off.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			9,
			5,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"client_id",
				"professional_id",
				"start_at",
				"end_at",
				"reason",
			}).AddRow(
				1,
				9,
				5,
				"2026-09-21T15:00",
				"2026-09-21T16:00",
				"Consulta",
			),
		)

	// 6. Appointment existente: 10:00–11:00.
	mock.ExpectQuery(
		`(?s)SELECT.*COALESCE\(professional_id,\s*0\).*FROM appointments.*client_id = \$1.*professional_id = \$2.*professional_id IS NULL.*date = \$3.*status\s+IN\s+\('scheduled',\s*'confirmed'\).*id <> \$4`,
	).
		WithArgs(
			9,
			5,
			"2026-09-21",
			0,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"professional_id",
				"time",
				"duration_snapshot",
			}).AddRow(
				101,
				5,
				"10:00",
				60,
			),
		)

	usecase := NewAppointmentUsecase()

	slots, err :=
		usecase.GetAvailableSlotsForProfessionalService(
			9,
			5,
			"2026-09-21",
			"Corte",
		)

	if err != nil {
		t.Fatalf(
			"não esperava erro: %v",
			err,
		)
	}

	expected := []string{
		"13:00",
		"13:30",
		"16:00",
		"16:30",
	}

	if len(slots) != len(expected) {
		t.Fatalf(
			"esperava %d horários, recebeu %d: %v",
			len(expected),
			len(slots),
			slots,
		)
	}

	for i := range expected {
		if slots[i] != expected[i] {
			t.Fatalf(
				"posição %d: esperava %s, recebeu %s",
				i,
				expected[i],
				slots[i],
			)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas do banco não atendidas: %v",
			err,
		)
	}
}
