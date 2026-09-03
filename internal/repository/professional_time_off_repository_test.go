package repository

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateProfessionalTimeOffSuccess(t *testing.T) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectQuery(
		`SELECT EXISTS`,
	).
		WithArgs(
			9,
			5,
			"2026-09-15T14:00",
			"2026-09-15T16:00",
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(false),
		)

	rows := sqlmock.NewRows([]string{
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
		"2026-09-15T14:00",
		"2026-09-15T16:00",
		"Consulta",
	)

	mock.ExpectQuery(
		`INSERT INTO professional_time_off`,
	).
		WithArgs(
			9,
			5,
			"2026-09-15T14:00",
			"2026-09-15T16:00",
			"Consulta",
		).
		WillReturnRows(rows)

	timeOff, err := CreateProfessionalTimeOff(
		9,
		5,
		"2026-09-15T14:00",
		"2026-09-15T16:00",
		"Consulta",
	)

	if err != nil {
		t.Fatalf(
			"esperava sucesso, recebeu erro: %v",
			err,
		)
	}

	if timeOff.ID != 1 {
		t.Errorf(
			"esperava ID 1, recebeu %d",
			timeOff.ID,
		)
	}

	if timeOff.ClientID != 9 {
		t.Errorf(
			"esperava client_id 9, recebeu %d",
			timeOff.ClientID,
		)
	}

	if timeOff.ProfessionalID != 5 {
		t.Errorf(
			"esperava professional_id 5, recebeu %d",
			timeOff.ProfessionalID,
		)
	}

	if timeOff.StartAt != "2026-09-15T14:00" {
		t.Errorf(
			"start_at inesperado: %s",
			timeOff.StartAt,
		)
	}

	if timeOff.EndAt != "2026-09-15T16:00" {
		t.Errorf(
			"end_at inesperado: %s",
			timeOff.EndAt,
		)
	}

	if timeOff.Reason != "Consulta" {
		t.Errorf(
			"esperava reason Consulta, recebeu %q",
			timeOff.Reason,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestCreateProfessionalTimeOffInvalidPeriod(t *testing.T) {
	tests := []struct {
		name    string
		startAt string
		endAt   string
	}{
		{
			name:    "invalid start format",
			startAt: "15/09/2026 14:00",
			endAt:   "2026-09-15T16:00",
		},
		{
			name:    "invalid end format",
			startAt: "2026-09-15T14:00",
			endAt:   "15/09/2026 16:00",
		},
		{
			name:    "start equals end",
			startAt: "2026-09-15T14:00",
			endAt:   "2026-09-15T14:00",
		},
		{
			name:    "start after end",
			startAt: "2026-09-15T16:00",
			endAt:   "2026-09-15T14:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := setupProfessionalRepositoryMock(t)

			_, err := CreateProfessionalTimeOff(
				9,
				5,
				tt.startAt,
				tt.endAt,
				"Consulta",
			)

			if !errors.Is(
				err,
				ErrProfessionalTimeOffInvalid,
			) {
				t.Fatalf(
					"esperava ErrProfessionalTimeOffInvalid, recebeu %v",
					err,
				)
			}

			assertProfessionalRepositoryExpectations(
				t,
				mock,
			)
		})
	}
}

func TestCreateProfessionalTimeOffProfessionalNotFound(t *testing.T) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectQuery(
		`SELECT EXISTS`,
	).
		WithArgs(
			10,
			5,
			"2026-09-15T14:00",
			"2026-09-15T16:00",
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(false),
		)

	mock.ExpectQuery(
		`INSERT INTO professional_time_off`,
	).
		WithArgs(
			10,
			5,
			"2026-09-15T14:00",
			"2026-09-15T16:00",
			"Consulta",
		).
		WillReturnError(sql.ErrNoRows)

	_, err := CreateProfessionalTimeOff(
		10,
		5,
		"2026-09-15T14:00",
		"2026-09-15T16:00",
		"Consulta",
	)

	if !errors.Is(
		err,
		ErrProfessionalNotFound,
	) {
		t.Fatalf(
			"esperava ErrProfessionalNotFound, recebeu %v",
			err,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestCreateProfessionalTimeOffTrimsReason(t *testing.T) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectQuery(
		`SELECT EXISTS`,
	).
		WithArgs(
			9,
			5,
			"2026-09-15T14:00",
			"2026-09-15T16:00",
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(false),
		)

	rows := sqlmock.NewRows([]string{
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
		"2026-09-15T14:00",
		"2026-09-15T16:00",
		"Consulta",
	)

	mock.ExpectQuery(
		`INSERT INTO professional_time_off`,
	).
		WithArgs(
			9,
			5,
			"2026-09-15T14:00",
			"2026-09-15T16:00",
			"Consulta",
		).
		WillReturnRows(rows)

	timeOff, err := CreateProfessionalTimeOff(
		9,
		5,
		"2026-09-15T14:00",
		"2026-09-15T16:00",
		"   Consulta   ",
	)

	if err != nil {
		t.Fatalf(
			"esperava sucesso, recebeu erro: %v",
			err,
		)
	}

	if timeOff.Reason != "Consulta" {
		t.Errorf(
			"esperava reason normalizado, recebeu %q",
			timeOff.Reason,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}
func TestListProfessionalTimeOffSuccess(t *testing.T) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectQuery(
		`SELECT 1\s+FROM professionals`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(1),
		)

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"professional_id",
		"start_at",
		"end_at",
		"reason",
	}).
		AddRow(
			1,
			9,
			5,
			"2026-09-15T14:00",
			"2026-09-15T16:00",
			"Consulta",
		).
		AddRow(
			2,
			9,
			5,
			"2026-09-20T00:00",
			"2026-09-27T23:59",
			"Férias",
		)

	mock.ExpectQuery(
		`SELECT\s+id,\s+client_id,\s+professional_id,`,
	).
		WithArgs(
			9,
			5,
		).
		WillReturnRows(rows)

	timeOffs, err := ListProfessionalTimeOff(
		9,
		5,
	)

	if err != nil {
		t.Fatalf(
			"esperava sucesso, recebeu erro: %v",
			err,
		)
	}

	if len(timeOffs) != 2 {
		t.Fatalf(
			"esperava 2 indisponibilidades, recebeu %d",
			len(timeOffs),
		)
	}

	if timeOffs[0].ID != 1 {
		t.Errorf(
			"esperava primeiro ID 1, recebeu %d",
			timeOffs[0].ID,
		)
	}

	if timeOffs[0].Reason != "Consulta" {
		t.Errorf(
			"esperava primeiro motivo Consulta, recebeu %q",
			timeOffs[0].Reason,
		)
	}

	if timeOffs[1].ID != 2 {
		t.Errorf(
			"esperava segundo ID 2, recebeu %d",
			timeOffs[1].ID,
		)
	}

	if timeOffs[1].Reason != "Férias" {
		t.Errorf(
			"esperava segundo motivo Férias, recebeu %q",
			timeOffs[1].Reason,
		)
	}

	if timeOffs[1].StartAt != "2026-09-20T00:00" {
		t.Errorf(
			"start_at inesperado: %s",
			timeOffs[1].StartAt,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestListProfessionalTimeOffEmpty(t *testing.T) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectQuery(
		`SELECT 1\s+FROM professionals`,
	).
		WithArgs(
			5,
			9,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(1),
		)

	mock.ExpectQuery(
		`SELECT\s+id,\s+client_id,\s+professional_id,`,
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

	timeOffs, err := ListProfessionalTimeOff(
		9,
		5,
	)

	if err != nil {
		t.Fatalf(
			"esperava sucesso, recebeu erro: %v",
			err,
		)
	}

	if timeOffs == nil {
		t.Fatal(
			"esperava slice vazia, recebeu nil",
		)
	}

	if len(timeOffs) != 0 {
		t.Fatalf(
			"esperava 0 indisponibilidades, recebeu %d",
			len(timeOffs),
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestListProfessionalTimeOffProfessionalNotFound(t *testing.T) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectQuery(
		`SELECT 1\s+FROM professionals`,
	).
		WithArgs(
			5,
			10,
		).
		WillReturnError(sql.ErrNoRows)

	timeOffs, err := ListProfessionalTimeOff(
		10,
		5,
	)

	if !errors.Is(
		err,
		ErrProfessionalNotFound,
	) {
		t.Fatalf(
			"esperava ErrProfessionalNotFound, recebeu %v",
			err,
		)
	}

	if timeOffs != nil {
		t.Fatalf(
			"esperava nil, recebeu %#v",
			timeOffs,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestDeleteProfessionalTimeOffSuccess(t *testing.T) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectExec(
		`DELETE FROM professional_time_off`,
	).
		WithArgs(
			3,
			9,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := DeleteProfessionalTimeOff(
		9,
		3,
	)

	if err != nil {
		t.Fatalf(
			"esperava sucesso, recebeu erro: %v",
			err,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestDeleteProfessionalTimeOffNotFound(t *testing.T) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectExec(
		`DELETE FROM professional_time_off`,
	).
		WithArgs(
			3,
			10,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := DeleteProfessionalTimeOff(
		10,
		3,
	)

	if !errors.Is(
		err,
		ErrProfessionalTimeOffNotFound,
	) {
		t.Fatalf(
			"esperava ErrProfessionalTimeOffNotFound, recebeu %v",
			err,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestCreateProfessionalTimeOffOverlap(t *testing.T) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectQuery(
		`SELECT EXISTS`,
	).
		WithArgs(
			9,
			5,
			"2026-09-15T15:00",
			"2026-09-15T17:00",
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(true),
		)

	_, err := CreateProfessionalTimeOff(
		9,
		5,
		"2026-09-15T15:00",
		"2026-09-15T17:00",
		"Bloqueio",
	)

	if !errors.Is(
		err,
		ErrProfessionalTimeOffOverlap,
	) {
		t.Fatalf(
			"esperava ErrProfessionalTimeOffOverlap, recebeu %v",
			err,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestCreateProfessionalTimeOffAdjacentPeriod(t *testing.T) {
	mock := setupProfessionalRepositoryMock(t)

	mock.ExpectQuery(
		`SELECT EXISTS`,
	).
		WithArgs(
			9,
			5,
			"2026-09-15T16:00",
			"2026-09-15T18:00",
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(false),
		)

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"professional_id",
		"start_at",
		"end_at",
		"reason",
	}).AddRow(
		2,
		9,
		5,
		"2026-09-15T16:00",
		"2026-09-15T18:00",
		"Bloqueio",
	)

	mock.ExpectQuery(
		`INSERT INTO professional_time_off`,
	).
		WithArgs(
			9,
			5,
			"2026-09-15T16:00",
			"2026-09-15T18:00",
			"Bloqueio",
		).
		WillReturnRows(rows)

	timeOff, err := CreateProfessionalTimeOff(
		9,
		5,
		"2026-09-15T16:00",
		"2026-09-15T18:00",
		"Bloqueio",
	)

	if err != nil {
		t.Fatalf(
			"esperava sucesso para períodos adjacentes, recebeu %v",
			err,
		)
	}

	if timeOff.ID != 2 {
		t.Errorf(
			"esperava ID 2, recebeu %d",
			timeOff.ID,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}
