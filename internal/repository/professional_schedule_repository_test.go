package repository

import (
	"errors"
	"flitta/internal/model"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestReplaceProfessionalWorkingHours(
	t *testing.T,
) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectBegin()

	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*WHERE id = \$1.*client_id = \$2`,
	).
		WithArgs(
			10,
			7,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	mock.ExpectExec(
		`(?s)DELETE FROM professional_working_hours.*WHERE client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			7,
			10,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 2),
		)

	mock.ExpectExec(
		`(?s)INSERT INTO professional_working_hours.*VALUES`,
	).
		WithArgs(
			7,
			10,
			1,
			"09:00",
			"12:00",
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		`(?s)INSERT INTO professional_working_hours.*VALUES`,
	).
		WithArgs(
			7,
			10,
			1,
			"13:00",
			"18:00",
		).
		WillReturnResult(
			sqlmock.NewResult(2, 1),
		)

	mock.ExpectCommit()

	err := ReplaceProfessionalWorkingHours(
		7,
		10,
		[]model.ProfessionalWorkingHour{
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
		},
	)

	if err != nil {
		t.Fatalf(
			"erro inesperado: %v",
			err,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestReplaceProfessionalWorkingHoursCanClearSchedule(
	t *testing.T,
) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectBegin()

	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*WHERE id = \$1.*client_id = \$2`,
	).
		WithArgs(
			10,
			7,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	mock.ExpectExec(
		`(?s)DELETE FROM professional_working_hours.*WHERE client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			7,
			10,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 5),
		)

	mock.ExpectCommit()

	err := ReplaceProfessionalWorkingHours(
		7,
		10,
		[]model.ProfessionalWorkingHour{},
	)

	if err != nil {
		t.Fatalf(
			"erro inesperado ao limpar agenda: %v",
			err,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestReplaceProfessionalWorkingHoursProfessionalNotFound(
	t *testing.T,
) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectBegin()

	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*WHERE id = \$1.*client_id = \$2`,
	).
		WithArgs(
			999,
			7,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			),
		)

	mock.ExpectRollback()

	err := ReplaceProfessionalWorkingHours(
		7,
		999,
		[]model.ProfessionalWorkingHour{
			{
				Weekday: 1,
				Start:   "09:00",
				End:     "18:00",
			},
		},
	)

	if !errors.Is(
		err,
		ErrProfessionalNotFound,
	) {
		t.Fatalf(
			"esperado ErrProfessionalNotFound, recebido %v",
			err,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestReplaceProfessionalWorkingHoursRejectsOverlap(
	t *testing.T,
) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectBegin()

	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*WHERE id = \$1.*client_id = \$2`,
	).
		WithArgs(
			10,
			7,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	mock.ExpectRollback()

	err := ReplaceProfessionalWorkingHours(
		7,
		10,
		[]model.ProfessionalWorkingHour{
			{
				Weekday: 1,
				Start:   "09:00",
				End:     "12:00",
			},
			{
				Weekday: 1,
				Start:   "11:00",
				End:     "14:00",
			},
		},
	)

	if !errors.Is(
		err,
		ErrProfessionalWorkingHoursOverlap,
	) {
		t.Fatalf(
			"esperado ErrProfessionalWorkingHoursOverlap, recebido %v",
			err,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestValidateProfessionalWorkingHoursRejectsInvalidValues(
	t *testing.T,
) {
	tests := []struct {
		name  string
		hours []model.ProfessionalWorkingHour
	}{
		{
			name: "weekday zero",
			hours: []model.ProfessionalWorkingHour{
				{
					Weekday: 0,
					Start:   "09:00",
					End:     "18:00",
				},
			},
		},
		{
			name: "weekday eight",
			hours: []model.ProfessionalWorkingHour{
				{
					Weekday: 8,
					Start:   "09:00",
					End:     "18:00",
				},
			},
		},
		{
			name: "invalid start",
			hours: []model.ProfessionalWorkingHour{
				{
					Weekday: 1,
					Start:   "99:00",
					End:     "18:00",
				},
			},
		},
		{
			name: "end before start",
			hours: []model.ProfessionalWorkingHour{
				{
					Weekday: 1,
					Start:   "18:00",
					End:     "09:00",
				},
			},
		},
		{
			name: "same start and end",
			hours: []model.ProfessionalWorkingHour{
				{
					Weekday: 1,
					Start:   "09:00",
					End:     "09:00",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				err := validateProfessionalWorkingHours(
					tt.hours,
				)

				if !errors.Is(
					err,
					ErrProfessionalWorkingHoursInvalid,
				) {
					t.Fatalf(
						"esperado ErrProfessionalWorkingHoursInvalid, recebido %v",
						err,
					)
				}
			},
		)
	}
}

func TestValidateProfessionalWorkingHoursAllowsAdjacentRanges(
	t *testing.T,
) {
	err := validateProfessionalWorkingHours(
		[]model.ProfessionalWorkingHour{
			{
				Weekday: 1,
				Start:   "09:00",
				End:     "12:00",
			},
			{
				Weekday: 1,
				Start:   "12:00",
				End:     "18:00",
			},
		},
	)

	if err != nil {
		t.Fatalf(
			"faixas adjacentes deveriam ser válidas: %v",
			err,
		)
	}
}

func TestListProfessionalWorkingHours(
	t *testing.T,
) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*WHERE id = \$1.*client_id = \$2`,
	).
		WithArgs(
			10,
			7,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(1),
		)

	rows := sqlmock.NewRows(
		[]string{
			"id",
			"weekday",
			"start_time",
			"end_time",
		},
	).
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
		).
		AddRow(
			3,
			2,
			"10:00",
			"17:00",
		)

	mock.ExpectQuery(
		`(?s)FROM professional_working_hours.*WHERE client_id = \$1.*professional_id = \$2.*ORDER BY.*weekday.*start_time`,
	).
		WithArgs(
			7,
			10,
		).
		WillReturnRows(rows)

	hours, err :=
		ListProfessionalWorkingHours(
			7,
			10,
		)

	if err != nil {
		t.Fatalf(
			"erro inesperado: %v",
			err,
		)
	}

	if hours == nil {
		t.Fatal(
			"hours deveria ser array, não nil",
		)
	}

	if len(hours) != 3 {
		t.Fatalf(
			"esperadas 3 faixas, recebidas %d",
			len(hours),
		)
	}

	if hours[0].ID != 1 ||
		hours[0].Weekday != 1 ||
		hours[0].Start != "09:00" ||
		hours[0].End != "12:00" {
		t.Fatalf(
			"primeira faixa inesperada: %+v",
			hours[0],
		)
	}

	if hours[2].Weekday != 2 ||
		hours[2].Start != "10:00" ||
		hours[2].End != "17:00" {
		t.Fatalf(
			"terceira faixa inesperada: %+v",
			hours[2],
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestListProfessionalWorkingHoursProfessionalNotFound(
	t *testing.T,
) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*WHERE id = \$1.*client_id = \$2`,
	).
		WithArgs(
			999,
			7,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			),
		)

	hours, err :=
		ListProfessionalWorkingHours(
			7,
			999,
		)

	if hours != nil {
		t.Fatalf(
			"esperado nil para profissional inexistente, recebido %+v",
			hours,
		)
	}

	if !errors.Is(
		err,
		ErrProfessionalNotFound,
	) {
		t.Fatalf(
			"esperado ErrProfessionalNotFound, recebido %v",
			err,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}
