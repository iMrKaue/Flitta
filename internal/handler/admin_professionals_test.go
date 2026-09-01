package handler

import (
	"context"
	"flitta/internal/database"
	"flitta/internal/middleware"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func setupProfessionalHandlerMock(
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

func professionalRequestWithClientID(
	req *http.Request,
	clientID int,
) *http.Request {
	ctx := context.WithValue(
		req.Context(),
		middleware.ClientIDKey,
		clientID,
	)

	return req.WithContext(ctx)
}

func TestCreateProfessionalHandlerSuccess(
	t *testing.T,
) {
	mock := setupProfessionalHandlerMock(t)

	mock.ExpectQuery(
		`(?s)INSERT INTO professionals.*RETURNING.*id.*client_id.*name.*active`,
	).
		WithArgs(
			7,
			"Carlos",
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"client_id",
					"name",
					"active",
				},
			).AddRow(
				10,
				7,
				"Carlos",
				true,
			),
		)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/professionals/create",
		strings.NewReader(
			`{"name":"  Carlos  "}`,
		),
	)

	req = professionalRequestWithClientID(
		req,
		7,
	)

	rec := httptest.NewRecorder()

	CreateProfessionalHandler(
		rec,
		req,
	)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"status esperado %d, recebido %d: %s",
			http.StatusCreated,
			rec.Code,
			rec.Body.String(),
		)
	}

	if !strings.Contains(
		rec.Body.String(),
		`"name":"Carlos"`,
	) {
		t.Fatalf(
			"resposta inesperada: %s",
			rec.Body.String(),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}

func TestCreateProfessionalHandlerDuplicateReturnsConflict(
	t *testing.T,
) {
	mock := setupProfessionalHandlerMock(t)

	mock.ExpectQuery(
		`(?s)INSERT INTO professionals.*RETURNING`,
	).
		WithArgs(
			7,
			"Carlos",
		).
		WillReturnError(
			&pq.Error{
				Code:       "23505",
				Constraint: "professionals_client_name_normalized_unique",
			},
		)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/professionals/create",
		strings.NewReader(
			`{"name":"Carlos"}`,
		),
	)

	req = professionalRequestWithClientID(
		req,
		7,
	)

	rec := httptest.NewRecorder()

	CreateProfessionalHandler(
		rec,
		req,
	)

	if rec.Code != http.StatusConflict {
		t.Fatalf(
			"status esperado %d, recebido %d: %s",
			http.StatusConflict,
			rec.Code,
			rec.Body.String(),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}

func TestSetProfessionalActiveUsesAuthenticatedClientID(
	t *testing.T,
) {
	mock := setupProfessionalHandlerMock(t)

	mock.ExpectExec(
		`(?s)UPDATE professionals.*WHERE id = \$2.*client_id = \$3`,
	).
		WithArgs(
			false,
			44,
			7,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/professionals/active",
		strings.NewReader(
			`{
				"professional_id":44,
				"active":false,
				"client_id":999
			}`,
		),
	)

	req = professionalRequestWithClientID(
		req,
		7,
	)

	rec := httptest.NewRecorder()

	SetProfessionalActiveHandler(
		rec,
		req,
	)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"status esperado %d, recebido %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}

func TestReplaceProfessionalServicesCannotAccessOtherTenantProfessional(
	t *testing.T,
) {
	mock := setupProfessionalHandlerMock(t)

	mock.ExpectBegin()

	mock.ExpectQuery(
		`(?s)SELECT 1.*FROM professionals.*WHERE id = \$1.*client_id = \$2`,
	).
		WithArgs(
			44,
			7,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			),
		)

	mock.ExpectRollback()

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/professionals/services",
		strings.NewReader(
			`{
				"professional_id":44,
				"service_ids":[1,2],
				"client_id":999
			}`,
		),
	)

	req = professionalRequestWithClientID(
		req,
		7,
	)

	rec := httptest.NewRecorder()

	ReplaceProfessionalServicesHandler(
		rec,
		req,
	)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"status esperado %d, recebido %d: %s",
			http.StatusNotFound,
			rec.Code,
			rec.Body.String(),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}

func TestListProfessionalsHandler(
	t *testing.T,
) {
	mock := setupProfessionalHandlerMock(t)

	mock.ExpectQuery(
		`(?s)FROM professionals p.*WHERE p\.client_id = \$1`,
	).
		WithArgs(7).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"client_id",
					"name",
					"active",
					"service_id",
					"service_name",
					"duration",
					"price",
				},
			).AddRow(
				10,
				7,
				"Ana",
				true,
				1,
				"Corte",
				30,
				50.0,
			),
		)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/professionals/list",
		nil,
	)

	req = professionalRequestWithClientID(
		req,
		7,
	)

	rec := httptest.NewRecorder()

	ListProfessionalsHandler(
		rec,
		req,
	)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"status esperado %d, recebido %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}

	if !strings.Contains(
		rec.Body.String(),
		`"name":"Ana"`,
	) {
		t.Fatalf(
			"resposta inesperada: %s",
			rec.Body.String(),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}
