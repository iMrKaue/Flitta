package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flitta/internal/database"
	"flitta/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestLoginHandlerAuthenticatesActiveUser(t *testing.T) {
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

	hash, err := service.HashPassword("Teste1234")
	if err != nil {
		t.Fatalf("erro ao gerar hash: %v", err)
	}

	mock.ExpectQuery(
		`SELECT client_id, password FROM users WHERE LOWER\(BTRIM\(email\)\) = \$1 AND active = TRUE`,
	).
		WithArgs("owner@flitta.local").
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"client_id", "password"},
			).AddRow(
				42,
				hash,
			),
		)

	body := `{
		"email": "  OWNER@FLITTA.LOCAL  ",
		"password": "Teste1234"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	LoginHandler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status esperado %d, recebido %d: %s",
			http.StatusOK,
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

func TestLoginHandlerRejectsInactiveOrUnknownUser(t *testing.T) {
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
		`SELECT client_id, password FROM users WHERE LOWER\(BTRIM\(email\)\) = \$1 AND active = TRUE`,
	).
		WithArgs("inactive@flitta.local").
		WillReturnError(sql.ErrNoRows)

	body := `{
		"email": "inactive@flitta.local",
		"password": "Teste1234"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	LoginHandler(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}

func TestLoginHandlerRejectsInvalidPassword(t *testing.T) {
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

	hash, err := service.HashPassword("SenhaCorreta123")
	if err != nil {
		t.Fatalf("erro ao gerar hash: %v", err)
	}

	mock.ExpectQuery(
		`SELECT client_id, password FROM users WHERE LOWER\(BTRIM\(email\)\) = \$1 AND active = TRUE`,
	).
		WithArgs("owner@flitta.local").
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"client_id", "password"},
			).AddRow(
				42,
				hash,
			),
		)

	body := `{
		"email": "owner@flitta.local",
		"password": "SenhaErrada123"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	LoginHandler(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}

func TestLoginHandlerHandlesDatabaseError(t *testing.T) {
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
		`SELECT client_id, password FROM users WHERE LOWER\(BTRIM\(email\)\) = \$1 AND active = TRUE`,
	).
		WithArgs("owner@flitta.local").
		WillReturnError(
			errors.New("database unavailable"),
		)

	body := `{
		"email": "owner@flitta.local",
		"password": "Teste1234"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	LoginHandler(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}
