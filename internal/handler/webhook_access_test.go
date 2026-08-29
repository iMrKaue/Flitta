package handler

import (
	"flitta/internal/database"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestWebhookBlockedTwilioClientReturnsEmptyTwiML(
	t *testing.T,
) {
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
		`SELECT client_id FROM communication_channels`,
	).
		WithArgs(
			"twilio",
			"whatsapp",
			"whatsapp:+5511888888888",
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"client_id"},
			).AddRow(42),
		)

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

	form := url.Values{}
	form.Set(
		"From",
		"whatsapp:+5511999999999",
	)
	form.Set(
		"To",
		"whatsapp:+5511888888888",
	)
	form.Set("Body", "Oi")

	request := httptest.NewRequest(
		http.MethodPost,
		"/webhook",
		strings.NewReader(form.Encode()),
	)

	request.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	recorder := httptest.NewRecorder()

	WebhookHandler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	body := recorder.Body.String()

	if strings.Contains(body, "<Message>") {
		t.Fatalf(
			"TwiML não deveria conter mensagem: %s",
			body,
		)
	}

	if !strings.Contains(body, "<Response") {
		t.Fatalf(
			"TwiML inválido: %s",
			body,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}

func TestWebhookBlockedLocalClientReturnsForbidden(
	t *testing.T,
) {
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
		`SELECT client_id FROM communication_channels`,
	).
		WithArgs(
			"twilio",
			"whatsapp",
			"whatsapp:+5511888888888",
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"client_id"},
			).AddRow(42),
		)

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

	body := `{
		"from": "+5511999999999",
		"to": "+5511888888888",
		"text": "Oi"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/webhook",
		strings.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	WebhookHandler(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusForbidden,
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
