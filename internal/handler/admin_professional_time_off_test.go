package handler

import (
	"database/sql"
	"encoding/json"
	"flitta/internal/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateProfessionalTimeOffHandlerSuccess(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

	mock.ExpectQuery(
		`(?s)SELECT EXISTS.*FROM professional_time_off.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			7,
			10,
			"2026-09-15T14:00",
			"2026-09-15T16:00",
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(false),
		)

	mock.ExpectQuery(
		`(?s)INSERT INTO professional_time_off.*RETURNING`,
	).
		WithArgs(
			7,
			10,
			"2026-09-15T14:00",
			"2026-09-15T16:00",
			"Consulta",
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
				7,
				10,
				"2026-09-15T14:00",
				"2026-09-15T16:00",
				"Consulta",
			),
		)

	request := httptest.NewRequest(
		http.MethodPost,
		"/admin/professionals/time-off/create",
		strings.NewReader(`{
			"professional_id": 10,
			"start_at": "2026-09-15T14:00",
			"end_at": "2026-09-15T16:00",
			"reason": "Consulta"
		}`),
	)

	request = professionalScheduleRequestWithClient(
		request,
		7,
	)

	recorder := httptest.NewRecorder()

	CreateProfessionalTimeOffHandler(
		recorder,
		request,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"esperado HTTP 201, recebido %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response model.ProfessionalTimeOff

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"erro ao decodificar resposta: %v",
			err,
		)
	}

	if response.ID != 1 {
		t.Errorf(
			"esperado ID 1, recebido %d",
			response.ID,
		)
	}

	if response.ProfessionalID != 10 {
		t.Errorf(
			"esperado professional_id 10, recebido %d",
			response.ProfessionalID,
		)
	}

	if response.Reason != "Consulta" {
		t.Errorf(
			"esperado motivo Consulta, recebido %q",
			response.Reason,
		)
	}

	assertProfessionalScheduleHandlerExpectations(
		t,
		mock,
	)
}

func TestCreateProfessionalTimeOffHandlerRejectsOverlap(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

	mock.ExpectQuery(
		`(?s)SELECT EXISTS.*FROM professional_time_off.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			7,
			10,
			"2026-09-15T15:00",
			"2026-09-15T17:00",
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(true),
		)

	request := httptest.NewRequest(
		http.MethodPost,
		"/admin/professionals/time-off/create",
		strings.NewReader(`{
			"professional_id": 10,
			"start_at": "2026-09-15T15:00",
			"end_at": "2026-09-15T17:00",
			"reason": "Bloqueio"
		}`),
	)

	request = professionalScheduleRequestWithClient(
		request,
		7,
	)

	recorder := httptest.NewRecorder()

	CreateProfessionalTimeOffHandler(
		recorder,
		request,
	)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"esperado HTTP 409, recebido %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	assertProfessionalScheduleHandlerExpectations(
		t,
		mock,
	)
}

func TestCreateProfessionalTimeOffHandlerProfessionalNotFound(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

	mock.ExpectQuery(
		`(?s)SELECT EXISTS.*FROM professional_time_off.*client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			10,
			5,
			"2026-09-15T14:00",
			"2026-09-15T16:00",
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(false),
		)

	mock.ExpectQuery(
		`(?s)INSERT INTO professional_time_off.*RETURNING`,
	).
		WithArgs(
			10,
			5,
			"2026-09-15T14:00",
			"2026-09-15T16:00",
			"Consulta",
		).
		WillReturnError(sql.ErrNoRows)

	request := httptest.NewRequest(
		http.MethodPost,
		"/admin/professionals/time-off/create",
		strings.NewReader(`{
			"professional_id": 5,
			"start_at": "2026-09-15T14:00",
			"end_at": "2026-09-15T16:00",
			"reason": "Consulta"
		}`),
	)

	request = professionalScheduleRequestWithClient(
		request,
		10,
	)

	recorder := httptest.NewRecorder()

	CreateProfessionalTimeOffHandler(
		recorder,
		request,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"esperado HTTP 404, recebido %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	assertProfessionalScheduleHandlerExpectations(
		t,
		mock,
	)
}

func TestCreateProfessionalTimeOffHandlerInvalidPeriod(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

	request := httptest.NewRequest(
		http.MethodPost,
		"/admin/professionals/time-off/create",
		strings.NewReader(`{
			"professional_id": 10,
			"start_at": "2026-09-15T18:00",
			"end_at": "2026-09-15T14:00",
			"reason": "Consulta"
		}`),
	)

	request = professionalScheduleRequestWithClient(
		request,
		7,
	)

	recorder := httptest.NewRecorder()

	CreateProfessionalTimeOffHandler(
		recorder,
		request,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"esperado HTTP 400, recebido %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	assertProfessionalScheduleHandlerExpectations(
		t,
		mock,
	)
}

func TestListProfessionalTimeOffHandlerSuccess(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

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

	mock.ExpectQuery(
		`(?s)SELECT.*FROM professional_time_off.*WHERE client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			7,
			10,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"client_id",
				"professional_id",
				"start_at",
				"end_at",
				"reason",
			}).
				AddRow(
					1,
					7,
					10,
					"2026-09-15T14:00",
					"2026-09-15T16:00",
					"Consulta",
				).
				AddRow(
					2,
					7,
					10,
					"2026-09-20T00:00",
					"2026-09-28T00:00",
					"Férias",
				),
		)

	request := httptest.NewRequest(
		http.MethodGet,
		"/admin/professionals/time-off/list?professional_id=10",
		nil,
	)

	request = professionalScheduleRequestWithClient(
		request,
		7,
	)

	recorder := httptest.NewRecorder()

	ListProfessionalTimeOffHandler(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"esperado HTTP 200, recebido %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response struct {
		ProfessionalID int                         `json:"professional_id"`
		TimeOffs       []model.ProfessionalTimeOff `json:"time_offs"`
	}

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"erro ao decodificar resposta: %v",
			err,
		)
	}

	if response.ProfessionalID != 10 {
		t.Errorf(
			"esperado professional_id 10, recebido %d",
			response.ProfessionalID,
		)
	}

	if len(response.TimeOffs) != 2 {
		t.Fatalf(
			"esperadas 2 indisponibilidades, recebidas %d",
			len(response.TimeOffs),
		)
	}

	if response.TimeOffs[0].Reason != "Consulta" {
		t.Errorf(
			"esperado primeiro motivo Consulta, recebido %q",
			response.TimeOffs[0].Reason,
		)
	}

	if response.TimeOffs[1].Reason != "Férias" {
		t.Errorf(
			"esperado segundo motivo Férias, recebido %q",
			response.TimeOffs[1].Reason,
		)
	}

	assertProfessionalScheduleHandlerExpectations(
		t,
		mock,
	)
}

func TestListProfessionalTimeOffHandlerEmpty(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

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

	mock.ExpectQuery(
		`(?s)SELECT.*FROM professional_time_off.*WHERE client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			7,
			10,
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

	request := httptest.NewRequest(
		http.MethodGet,
		"/admin/professionals/time-off/list?professional_id=10",
		nil,
	)

	request = professionalScheduleRequestWithClient(
		request,
		7,
	)

	recorder := httptest.NewRecorder()

	ListProfessionalTimeOffHandler(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"esperado HTTP 200, recebido %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response struct {
		TimeOffs []model.ProfessionalTimeOff `json:"time_offs"`
	}

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"erro ao decodificar resposta: %v",
			err,
		)
	}

	if response.TimeOffs == nil {
		t.Fatal("esperava lista vazia, recebeu nil")
	}

	if len(response.TimeOffs) != 0 {
		t.Fatalf(
			"esperadas 0 indisponibilidades, recebidas %d",
			len(response.TimeOffs),
		)
	}

	assertProfessionalScheduleHandlerExpectations(
		t,
		mock,
	)
}

func TestListProfessionalTimeOffHandlerProfessionalNotFound(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*WHERE id = \$1.*client_id = \$2`,
	).
		WithArgs(
			5,
			10,
		).
		WillReturnError(sql.ErrNoRows)

	request := httptest.NewRequest(
		http.MethodGet,
		"/admin/professionals/time-off/list?professional_id=5",
		nil,
	)

	request = professionalScheduleRequestWithClient(
		request,
		10,
	)

	recorder := httptest.NewRecorder()

	ListProfessionalTimeOffHandler(
		recorder,
		request,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"esperado HTTP 404, recebido %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	assertProfessionalScheduleHandlerExpectations(
		t,
		mock,
	)
}

func TestDeleteProfessionalTimeOffHandlerSuccess(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

	mock.ExpectExec(
		`DELETE FROM professional_time_off`,
	).
		WithArgs(
			3,
			7,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/admin/professionals/time-off/delete?time_off_id=3",
		nil,
	)

	request = professionalScheduleRequestWithClient(
		request,
		7,
	)

	recorder := httptest.NewRecorder()

	DeleteProfessionalTimeOffHandler(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"esperado HTTP 200, recebido %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	assertProfessionalScheduleHandlerExpectations(
		t,
		mock,
	)
}

func TestDeleteProfessionalTimeOffHandlerNotFound(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

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

	request := httptest.NewRequest(
		http.MethodDelete,
		"/admin/professionals/time-off/delete?time_off_id=3",
		nil,
	)

	request = professionalScheduleRequestWithClient(
		request,
		10,
	)

	recorder := httptest.NewRecorder()

	DeleteProfessionalTimeOffHandler(
		recorder,
		request,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"esperado HTTP 404, recebido %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	assertProfessionalScheduleHandlerExpectations(
		t,
		mock,
	)
}

func TestDeleteProfessionalTimeOffHandlerRequiresID(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/admin/professionals/time-off/delete",
		nil,
	)

	request = professionalScheduleRequestWithClient(
		request,
		7,
	)

	recorder := httptest.NewRecorder()

	DeleteProfessionalTimeOffHandler(
		recorder,
		request,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"esperado HTTP 400, recebido %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	assertProfessionalScheduleHandlerExpectations(
		t,
		mock,
	)
}
