package handler

import (
	"encoding/json"
	"flitta/internal/database"
	"flitta/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRegisterHandlerCreatesBusinessAndOwner(t *testing.T) {
	t.Setenv(
		"JWT_SECRET",
		"flitta-test-secret-with-at-least-32-characters",
	)

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

	mock.ExpectBegin()

	mock.ExpectQuery(`INSERT INTO clients`).
		WithArgs(
			"Flitta Teste",
			sqlmock.AnyArg(),
			"owner@flitta.local",
			sqlmock.AnyArg(),
			"barber",
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(42),
		)

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(
			42,
			"Responsavel Teste",
			"owner@flitta.local",
			sqlmock.AnyArg(),
			"owner",
			true,
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(`INSERT INTO services`).
		WithArgs(42, "Corte").
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(`INSERT INTO services`).
		WithArgs(42, "Barba").
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(`INSERT INTO services`).
		WithArgs(42, "Sobrancelha").
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(`INSERT INTO working_hours`).
		WithArgs(
			42,
			"09:00",
			"18:00",
			60,
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectCommit()

	body := `{
		"responsible_name": "  Responsavel Teste  ",
		"business_name": "  Flitta Teste  ",
		"phone": "+5511999999010",
		"email": "  OWNER@FLITTA.LOCAL  ",
		"password": "Teste1234",
		"business_type": "barber"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	RegisterHandler(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"status esperado %d, recebido %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response map[string]string

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"resposta JSON inválida: %v",
			err,
		)
	}

	if response["token"] == "" {
		t.Fatal("token não foi retornado")
	}

	clientID, err := service.ValidateToken(
		response["token"],
	)
	if err != nil {
		t.Fatalf(
			"token retornado é inválido: %v",
			err,
		)
	}

	if clientID != 42 {
		t.Fatalf(
			"client_id esperado 42, recebido %d",
			clientID,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}

func TestRegisterHandlerRequiresResponsibleAndBusinessNames(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "responsible name ausente",
			body: `{
				"business_name": "Flitta Teste",
				"phone": "+5511999999010",
				"email": "owner@flitta.local",
				"password": "Teste1234",
				"business_type": "barber"
			}`,
		},
		{
			name: "business name ausente",
			body: `{
				"responsible_name": "Responsavel Teste",
				"phone": "+5511999999010",
				"email": "owner@flitta.local",
				"password": "Teste1234",
				"business_type": "barber"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodPost,
				"/auth/register",
				strings.NewReader(tt.body),
			)

			recorder := httptest.NewRecorder()

			RegisterHandler(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(
					"status esperado %d, recebido %d",
					http.StatusBadRequest,
					recorder.Code,
				)
			}
		})
	}
}

func TestRegisterHandlerRejectsLegacyNameField(t *testing.T) {
	body := `{
		"name": "Formato Antigo",
		"responsible_name": "Responsavel Teste",
		"business_name": "Flitta Teste",
		"phone": "+5511999999010",
		"email": "owner@flitta.local",
		"password": "Teste1234",
		"business_type": "barber"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	RegisterHandler(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}
