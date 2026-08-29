package service

import (
	"errors"
	"flitta/internal/database"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestValidateClientAccess(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		expiresAt   interface{}
		wantBlocked bool
	}{
		{
			name:        "active sem expiracao",
			status:      "active",
			expiresAt:   nil,
			wantBlocked: false,
		},
		{
			name:        "active com expiracao futura",
			status:      "active",
			expiresAt:   time.Now().Add(24 * time.Hour),
			wantBlocked: false,
		},
		{
			name:        "suspended",
			status:      "suspended",
			expiresAt:   nil,
			wantBlocked: true,
		},
		{
			name:        "active com piloto expirado",
			status:      "active",
			expiresAt:   time.Now().Add(-24 * time.Hour),
			wantBlocked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf(
					"erro ao criar sqlmock: %v",
					err,
				)
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
						tt.status,
						tt.expiresAt,
					),
				)

			err = ValidateClientAccess(42)

			if tt.wantBlocked {
				if !errors.Is(
					err,
					ErrClientAccessBlocked,
				) {
					t.Fatalf(
						"esperado bloqueio, recebido: %v",
						err,
					)
				}
			} else if err != nil {
				t.Fatalf(
					"acesso deveria ser permitido: %v",
					err,
				)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf(
					"expectativas SQL não atendidas: %v",
					err,
				)
			}
		})
	}
}
