package handler

import (
	"context"
	"encoding/json"
	"flitta/internal/database"
	"flitta/internal/middleware"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func setupProfessionalScheduleHandlerMock(
	t *testing.T,
) sqlmock.Sqlmock {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"erro ao criar sqlmock: %v",
			err,
		)
	}

	previousDB := database.DB
	database.DB = db

	t.Cleanup(func() {
		database.DB = previousDB
		db.Close()
	})

	return mock
}

func professionalScheduleRequestWithClient(
	request *http.Request,
	clientID int,
) *http.Request {
	ctx := context.WithValue(
		request.Context(),
		middleware.ClientIDKey,
		clientID,
	)

	return request.WithContext(ctx)
}

func assertProfessionalScheduleHandlerExpectations(
	t *testing.T,
	mock sqlmock.Sqlmock,
) {
	t.Helper()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}

func TestReplaceProfessionalWorkingHoursHandlerSuccess(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

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
			sqlmock.NewResult(0, 0),
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

	request := httptest.NewRequest(
		http.MethodPut,
		"/admin/professionals/hours",
		strings.NewReader(`{
"professional_id": 10,
"hours": [
{
"weekday": 1,
"start": "09:00",
"end": "12:00"
},
{
"weekday": 1,
"start": "13:00",
"end": "18:00"
}
]
}`),
	)

	request = professionalScheduleRequestWithClient(
		request,
		7,
	)

	recorder := httptest.NewRecorder()

	ReplaceProfessionalWorkingHoursHandler(
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

func TestReplaceProfessionalWorkingHoursHandlerRejectsOverlap(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

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

	request := httptest.NewRequest(
		http.MethodPut,
		"/admin/professionals/hours",
		strings.NewReader(`{
"professional_id": 10,
"hours": [
{
"weekday": 1,
"start": "09:00",
"end": "12:00"
},
{
"weekday": 1,
"start": "11:00",
"end": "14:00"
}
]
}`),
	)

	request = professionalScheduleRequestWithClient(
		request,
		7,
	)

	recorder := httptest.NewRecorder()

	ReplaceProfessionalWorkingHoursHandler(
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

func TestReplaceProfessionalWorkingHoursHandlerCannotAccessOtherTenantProfessional(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

	mock.ExpectBegin()

	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*WHERE id = \$1.*client_id = \$2`,
	).
		WithArgs(
			10,
			8,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			),
		)

	mock.ExpectRollback()

	request := httptest.NewRequest(
		http.MethodPut,
		"/admin/professionals/hours",
		strings.NewReader(`{
"professional_id": 10,
"hours": [
{
"weekday": 1,
"start": "09:00",
"end": "18:00"
}
]
}`),
	)

	request = professionalScheduleRequestWithClient(
		request,
		8,
	)

	recorder := httptest.NewRecorder()

	ReplaceProfessionalWorkingHoursHandler(
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

func TestReplaceProfessionalWorkingHoursHandlerRequiresHoursField(
	t *testing.T,
) {
	request := httptest.NewRequest(
		http.MethodPut,
		"/admin/professionals/hours",
		strings.NewReader(`{
"professional_id": 10
}`),
	)

	request = professionalScheduleRequestWithClient(
		request,
		7,
	)

	recorder := httptest.NewRecorder()

	ReplaceProfessionalWorkingHoursHandler(
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
}

func TestGetProfessionalWorkingHoursHandlerSuccess(
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
		)

	mock.ExpectQuery(
		`(?s)FROM professional_working_hours.*WHERE client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			7,
			10,
		).
		WillReturnRows(rows)

	request := httptest.NewRequest(
		http.MethodGet,
		"/admin/professionals/hours?professional_id=10",
		nil,
	)

	request = professionalScheduleRequestWithClient(
		request,
		7,
	)

	recorder := httptest.NewRecorder()

	GetProfessionalWorkingHoursHandler(
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
		ProfessionalID int `json:"professional_id"`

		Hours []struct {
			ID      int    `json:"id"`
			Weekday int    `json:"weekday"`
			Start   string `json:"start"`
			End     string `json:"end"`
		} `json:"hours"`
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
		t.Fatalf(
			"professional_id esperado 10, recebido %d",
			response.ProfessionalID,
		)
	}

	if len(response.Hours) != 2 {
		t.Fatalf(
			"esperadas 2 faixas, recebidas %d",
			len(response.Hours),
		)
	}

	if response.Hours[0].Weekday != 1 ||
		response.Hours[0].Start != "09:00" ||
		response.Hours[0].End != "12:00" {
		t.Fatalf(
			"primeira faixa inesperada: %+v",
			response.Hours[0],
		)
	}

	assertProfessionalScheduleHandlerExpectations(
		t,
		mock,
	)
}

func TestGetProfessionalWorkingHoursHandlerCannotAccessOtherTenantProfessional(
	t *testing.T,
) {
	mock := setupProfessionalScheduleHandlerMock(t)

	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*WHERE id = \$1.*client_id = \$2`,
	).
		WithArgs(
			10,
			8,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			),
		)

	request := httptest.NewRequest(
		http.MethodGet,
		"/admin/professionals/hours?professional_id=10",
		nil,
	)

	request = professionalScheduleRequestWithClient(
		request,
		8,
	)

	recorder := httptest.NewRecorder()

	GetProfessionalWorkingHoursHandler(
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

func TestGetProfessionalWorkingHoursHandlerRequiresProfessionalID(
	t *testing.T,
) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/admin/professionals/hours",
		nil,
	)

	request = professionalScheduleRequestWithClient(
		request,
		7,
	)

	recorder := httptest.NewRecorder()

	GetProfessionalWorkingHoursHandler(
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
}

func TestProfessionalWorkingHoursHandlersMethodNotAllowed(
	t *testing.T,
) {
	putRecorder := httptest.NewRecorder()

	putRequest := httptest.NewRequest(
		http.MethodPost,
		"/admin/professionals/hours",
		nil,
	)

	ReplaceProfessionalWorkingHoursHandler(
		putRecorder,
		putRequest,
	)

	if putRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"PUT handler deveria retornar 405, recebeu %d",
			putRecorder.Code,
		)
	}

	getRecorder := httptest.NewRecorder()

	getRequest := httptest.NewRequest(
		http.MethodPost,
		"/admin/professionals/hours",
		nil,
	)

	GetProfessionalWorkingHoursHandler(
		getRecorder,
		getRequest,
	)

	if getRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"GET handler deveria retornar 405, recebeu %d",
			getRecorder.Code,
		)
	}
}
