package repository

import (
	"errors"
	"flitta/internal/database"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func setupProfessionalRepositoryMock(
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

func assertProfessionalRepositoryExpectations(
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

func TestCreateProfessional(
	t *testing.T,
) {
	mock := setupProfessionalRepositoryMock(t)

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

	professional, err := CreateProfessional(
		7,
		"  Carlos  ",
	)

	if err != nil {
		t.Fatalf(
			"erro inesperado: %v",
			err,
		)
	}

	if professional.ID != 10 {
		t.Fatalf(
			"ID esperado 10, recebido %d",
			professional.ID,
		)
	}

	if professional.ClientID != 7 {
		t.Fatalf(
			"client_id esperado 7, recebido %d",
			professional.ClientID,
		)
	}

	if professional.Name != "Carlos" {
		t.Fatalf(
			"nome esperado Carlos, recebido %q",
			professional.Name,
		)
	}

	if !professional.Active {
		t.Fatal(
			"profissional deveria estar ativo",
		)
	}

	if professional.Services == nil {
		t.Fatal(
			"services deveria ser array vazio, não nil",
		)
	}

	if len(professional.Services) != 0 {
		t.Fatalf(
			"esperado 0 serviços, recebido %d",
			len(professional.Services),
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestListProfessionalsWithServices(
	t *testing.T,
) {
	mock := setupProfessionalRepositoryMock(t)

	rows := sqlmock.NewRows(
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
	).
		AddRow(
			10,
			7,
			"Ana",
			true,
			1,
			"Corte",
			30,
			50.0,
		).
		AddRow(
			10,
			7,
			"Ana",
			true,
			2,
			"Barba",
			45,
			35.0,
		).
		AddRow(
			11,
			7,
			"Bruno",
			false,
			0,
			"",
			30,
			0.0,
		)

	mock.ExpectQuery(
		`(?s)FROM professionals p.*LEFT JOIN professional_services ps.*WHERE p\.client_id = \$1`,
	).
		WithArgs(7).
		WillReturnRows(rows)

	professionals, err :=
		ListProfessionals(7)

	if err != nil {
		t.Fatalf(
			"erro inesperado: %v",
			err,
		)
	}

	if len(professionals) != 2 {
		t.Fatalf(
			"esperados 2 profissionais, recebidos %d",
			len(professionals),
		)
	}

	ana := professionals[0]

	if ana.ID != 10 ||
		ana.Name != "Ana" ||
		!ana.Active {
		t.Fatalf(
			"dados inesperados para Ana: %+v",
			ana,
		)
	}

	if len(ana.Services) != 2 {
		t.Fatalf(
			"Ana deveria possuir 2 serviços, recebeu %d",
			len(ana.Services),
		)
	}

	if ana.Services[0].Name != "Corte" {
		t.Fatalf(
			"primeiro serviço esperado Corte, recebido %q",
			ana.Services[0].Name,
		)
	}

	if ana.Services[1].Name != "Barba" {
		t.Fatalf(
			"segundo serviço esperado Barba, recebido %q",
			ana.Services[1].Name,
		)
	}

	bruno := professionals[1]

	if bruno.ID != 11 ||
		bruno.Name != "Bruno" ||
		bruno.Active {
		t.Fatalf(
			"dados inesperados para Bruno: %+v",
			bruno,
		)
	}

	if bruno.Services == nil {
		t.Fatal(
			"services de Bruno deveria ser array vazio",
		)
	}

	if len(bruno.Services) != 0 {
		t.Fatalf(
			"Bruno deveria possuir 0 serviços, recebeu %d",
			len(bruno.Services),
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}

func TestReplaceProfessionalServices(
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
		`(?s)DELETE FROM professional_services.*WHERE client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			7,
			10,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 2),
		)

	mock.ExpectExec(
		`(?s)INSERT INTO professional_services.*FROM services s.*s\.id = \$3.*s\.client_id = \$1`,
	).
		WithArgs(
			7,
			10,
			1,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	mock.ExpectExec(
		`(?s)INSERT INTO professional_services.*FROM services s.*s\.id = \$3.*s\.client_id = \$1`,
	).
		WithArgs(
			7,
			10,
			2,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	mock.ExpectCommit()

	err := ReplaceProfessionalServices(
		7,
		10,
		[]int{1, 2},
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

func TestReplaceProfessionalServicesProfessionalNotFound(
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

	err := ReplaceProfessionalServices(
		7,
		999,
		[]int{1},
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

func TestReplaceProfessionalServicesRollsBackWhenServiceIsInvalid(
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
		`(?s)DELETE FROM professional_services.*WHERE client_id = \$1.*professional_id = \$2`,
	).
		WithArgs(
			7,
			10,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 2),
		)

	mock.ExpectExec(
		`(?s)INSERT INTO professional_services.*FROM services s.*s\.id = \$3.*s\.client_id = \$1`,
	).
		WithArgs(
			7,
			10,
			1,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	mock.ExpectExec(
		`(?s)INSERT INTO professional_services.*FROM services s.*s\.id = \$3.*s\.client_id = \$1`,
	).
		WithArgs(
			7,
			10,
			999,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	mock.ExpectRollback()

	err := ReplaceProfessionalServices(
		7,
		10,
		[]int{
			1,
			999,
		},
	)

	if !errors.Is(
		err,
		ErrProfessionalServiceNotFound,
	) {
		t.Fatalf(
			"esperado ErrProfessionalServiceNotFound, recebido %v",
			err,
		)
	}

	assertProfessionalRepositoryExpectations(
		t,
		mock,
	)
}
