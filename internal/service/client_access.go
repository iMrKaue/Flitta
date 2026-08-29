package service

import (
	"database/sql"
	"errors"
	"flitta/internal/repository"
	"fmt"
	"time"
)

var ErrClientAccessBlocked = errors.New(
	"acesso ao estabelecimento bloqueado",
)

func ValidateClientAccess(clientID int) error {
	if clientID <= 0 {
		return ErrClientAccessBlocked
	}

	status, pilotExpiresAt, err :=
		repository.GetClientAccessState(clientID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrClientAccessBlocked
		}

		return fmt.Errorf(
			"consultar estado de acesso do estabelecimento: %w",
			err,
		)
	}

	if status != "active" {
		return ErrClientAccessBlocked
	}

	if pilotExpiresAt != nil &&
		!pilotExpiresAt.After(time.Now()) {
		return ErrClientAccessBlocked
	}

	return nil
}
