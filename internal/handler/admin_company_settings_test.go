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

func companyRequestWithClientID(
	method string,
	path string,
	body string,
	clientID int,
) *http.Request {
	request := httptest.NewRequest(
		method,
		path,
		strings.NewReader(body),
	)

	ctx := context.WithValue(
		request.Context(),
		middleware.ClientIDKey,
		clientID,
	)

	return request.WithContext(ctx)
}

func TestGetCompanySettingsReturnsBusinessProfile(
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
		`(?s)SELECT.*id.*name.*phone.*business_type.*COALESCE\(description, ''\).*COALESCE\(address, ''\).*COALESCE\(city, ''\).*COALESCE\(instagram, ''\).*COALESCE\(welcome_message, ''\).*FROM clients.*WHERE id = \$1`,
	).
		WithArgs(42).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"name",
					"phone",
					"business_type",
					"description",
					"address",
					"city",
					"instagram",
					"welcome_message",
				},
			).AddRow(
				42,
				"Flitta Studio",
				"+5511999999999",
				"barber",
				"Barbearia especializada em cortes modernos.",
				"Rua Flitta, 100",
				"São Paulo",
				"@flittastudio",
				"Olá! Bem-vindo ao Flitta Studio.",
			),
		)

	request := companyRequestWithClientID(
		http.MethodGet,
		"/admin/company/get",
		"",
		42,
	)

	recorder := httptest.NewRecorder()

	GetCompanySettingsHandler(
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

	var response struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		Address        string `json:"address"`
		City           string `json:"city"`
		Instagram      string `json:"instagram"`
		WelcomeMessage string `json:"welcome_message"`
	}

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	); err != nil {
		t.Fatalf(
			"resposta JSON inválida: %v",
			err,
		)
	}

	if response.Name != "Flitta Studio" {
		t.Fatalf(
			"nome inesperado: %q",
			response.Name,
		)
	}

	if response.City != "São Paulo" {
		t.Fatalf(
			"cidade inesperada: %q",
			response.City,
		)
	}

	if response.WelcomeMessage !=
		"Olá! Bem-vindo ao Flitta Studio." {
		t.Fatalf(
			"welcome_message inesperada: %q",
			response.WelcomeMessage,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}

func TestUpdateCompanySettingsUpdatesBusinessProfile(
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

	mock.ExpectExec(
		`(?s)UPDATE clients.*name = \$1.*phone = \$2.*business_type = \$3.*description = NULLIF\(BTRIM\(\$4\), ''\).*address = NULLIF\(BTRIM\(\$5\), ''\).*city = NULLIF\(BTRIM\(\$6\), ''\).*instagram = NULLIF\(BTRIM\(\$7\), ''\).*welcome_message = NULLIF\(BTRIM\(\$8\), ''\).*WHERE id = \$9`,
	).
		WithArgs(
			"Flitta Studio",
			sqlmock.AnyArg(),
			"barber",
			"Barbearia moderna",
			"Rua Flitta, 100",
			"São Paulo",
			"@flittastudio",
			"Olá! Como podemos ajudar?",
			42,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	mock.ExpectQuery(
		`(?s)SELECT.*id.*name.*phone.*business_type.*COALESCE\(description, ''\).*FROM clients.*WHERE id = \$1`,
	).
		WithArgs(42).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"name",
					"phone",
					"business_type",
					"description",
					"address",
					"city",
					"instagram",
					"welcome_message",
				},
			).AddRow(
				42,
				"Flitta Studio",
				"+5511999999999",
				"barber",
				"Barbearia moderna",
				"Rua Flitta, 100",
				"São Paulo",
				"@flittastudio",
				"Olá! Como podemos ajudar?",
			),
		)

	body := `{
		"name": " Flitta Studio ",
		"phone": "+5511999999999",
		"business_type": "barber",
		"description": " Barbearia moderna ",
		"address": " Rua Flitta, 100 ",
		"city": " São Paulo ",
		"instagram": " @flittastudio ",
		"welcome_message": " Olá! Como podemos ajudar? "
	}`

	request := companyRequestWithClientID(
		http.MethodPut,
		"/admin/company/update",
		body,
		42,
	)

	recorder := httptest.NewRecorder()

	UpdateCompanySettingsHandler(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status esperado %d, recebido %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response map[string]interface{}

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	); err != nil {
		t.Fatalf(
			"resposta JSON inválida: %v",
			err,
		)
	}

	if response["description"] != "Barbearia moderna" {
		t.Fatalf(
			"description inesperada: %v",
			response["description"],
		)
	}

	if response["instagram"] != "@flittastudio" {
		t.Fatalf(
			"instagram inesperado: %v",
			response["instagram"],
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expectativas SQL não atendidas: %v",
			err,
		)
	}
}

func TestUpdateCompanySettingsRejectsEmptyName(
	t *testing.T,
) {
	request := companyRequestWithClientID(
		http.MethodPut,
		"/admin/company/update",
		`{
			"name": "   ",
			"phone": "+5511999999999",
			"business_type": "barber"
		}`,
		42,
	)

	recorder := httptest.NewRecorder()

	UpdateCompanySettingsHandler(
		recorder,
		request,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}
