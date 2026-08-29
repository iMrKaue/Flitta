package middleware

import (
	"flitta/internal/database"
	"flitta/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAuthMiddlewareAllowsActiveClient(t *testing.T) {
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

	mock.ExpectQuery(
		`SELECT status, pilot_expires_at FROM clients WHERE id = \$1`,
	).
		WithArgs(42).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"status",
					"pilot_expires_at",
				},
			).AddRow(
				"active",
				nil,
			),
		)

	token, err := service.GenerateToken(42)
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	called := false

	next := func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		called = true

		clientID, ok :=
			r.Context().Value(ClientIDKey).(int)

		if !ok || clientID != 42 {
			t.Fatalf(
				"client_id inesperado: %v",
				clientID,
			)
		}

		w.WriteHeader(http.StatusOK)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/admin/test",
		nil,
	)

	request.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	recorder := httptest.NewRecorder()

	AuthMiddleware(next)(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if !called {
		t.Fatal("handler protegido não foi executado")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}

func TestAuthMiddlewareBlocksSuspendedClient(t *testing.T) {
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

	mock.ExpectQuery(
		`SELECT status, pilot_expires_at FROM clients WHERE id = \$1`,
	).
		WithArgs(42).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"status",
					"pilot_expires_at",
				},
			).AddRow(
				"suspended",
				nil,
			),
		)

	token, err := service.GenerateToken(42)
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	called := false

	next := func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		called = true
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/admin/test",
		nil,
	)

	request.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	recorder := httptest.NewRecorder()

	AuthMiddleware(next)(
		recorder,
		request,
	)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusForbidden,
			recorder.Code,
		)
	}

	if called {
		t.Fatal(
			"handler protegido foi executado para cliente suspenso",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}
