package database

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"flitta/migrations"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
)

const migrationLockID int64 = 71920260423

type migrationFile struct {
	Name     string
	SQL      string
	Checksum string
}

func runMigrations(ctx context.Context) error {
	conn, err := DB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("obter conexão para migrations: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(
		ctx,
		`SELECT pg_advisory_lock($1)`,
		migrationLockID,
	); err != nil {
		return fmt.Errorf("obter lock de migrations: %w", err)
	}

	defer func() {
		_, _ = conn.ExecContext(
			context.Background(),
			`SELECT pg_advisory_unlock($1)`,
			migrationLockID,
		)
	}()

	if err := ensureSchemaMigrationsTable(ctx, conn); err != nil {
		return err
	}

	files, err := loadMigrationFiles()
	if err != nil {
		return err
	}

	for _, migration := range files {
		appliedChecksum, err := getAppliedMigration(
			ctx,
			conn,
			migration.Name,
		)
		if err != nil {
			return err
		}

		if appliedChecksum != "" {
			if appliedChecksum != migration.Checksum {
				return fmt.Errorf(
					"migration %s foi alterada após ser aplicada",
					migration.Name,
				)
			}

			continue
		}

		if err := applyMigration(ctx, conn, migration); err != nil {
			return err
		}
	}

	return nil
}

func ensureSchemaMigrationsTable(
	ctx context.Context,
	conn *sql.Conn,
) error {
	_, err := conn.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			checksum CHAR(64) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE
				NOT NULL
				DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf(
			"criar tabela schema_migrations: %w",
			err,
		)
	}

	return nil
}

func loadMigrationFiles() ([]migrationFile, error) {
	entries, err := fs.ReadDir(migrations.Files, ".")
	if err != nil {
		return nil, fmt.Errorf("listar migrations: %w", err)
	}

	files := make([]migrationFile, 0)

	for _, entry := range entries {
		if entry.IsDir() ||
			!strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := migrations.Files.ReadFile(entry.Name())
		if err != nil {
			return nil, fmt.Errorf(
				"ler migration %s: %w",
				entry.Name(),
				err,
			)
		}

		content = normalizeMigrationContent(content)

		sum := sha256.Sum256(content)

		files = append(files, migrationFile{
			Name:     entry.Name(),
			SQL:      string(content),
			Checksum: hex.EncodeToString(sum[:]),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return files, nil
}

func normalizeMigrationContent(content []byte) []byte {
	content = bytes.ReplaceAll(
		content,
		[]byte("\r\n"),
		[]byte("\n"),
	)

	content = bytes.ReplaceAll(
		content,
		[]byte("\r"),
		[]byte("\n"),
	)

	return content
}

func getAppliedMigration(
	ctx context.Context,
	conn *sql.Conn,
	version string,
) (string, error) {
	var checksum string

	err := conn.QueryRowContext(ctx, `
		SELECT checksum
		FROM schema_migrations
		WHERE version = $1
	`, version).Scan(&checksum)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}

		return "", fmt.Errorf(
			"consultar migration %s: %w",
			version,
			err,
		)
	}

	return checksum, nil
}

func applyMigration(
	ctx context.Context,
	conn *sql.Conn,
	migration migrationFile,
) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(
			"iniciar migration %s: %w",
			migration.Name,
			err,
		)
	}

	defer tx.Rollback()

	sqlBody, err := stripTransactionWrapper(migration.SQL)
	if err != nil {
		return fmt.Errorf(
			"preparar migration %s: %w",
			migration.Name,
			err,
		)
	}

	if _, err := tx.ExecContext(ctx, sqlBody); err != nil {
		return fmt.Errorf(
			"aplicar migration %s: %w",
			migration.Name,
			err,
		)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO schema_migrations (
			version,
			checksum
		)
		VALUES ($1, $2)
	`, migration.Name, migration.Checksum); err != nil {
		return fmt.Errorf(
			"registrar migration %s: %w",
			migration.Name,
			err,
		)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"confirmar migration %s: %w",
			migration.Name,
			err,
		)
	}

	log.Printf("Migration aplicada: %s", migration.Name)

	return nil
}

func stripTransactionWrapper(content string) (string, error) {
	content = strings.TrimSpace(content)

	if !strings.HasPrefix(content, "BEGIN;") {
		return "", errors.New(
			"migration não começa com BEGIN;",
		)
	}

	if !strings.HasSuffix(content, "COMMIT;") {
		return "", errors.New(
			"migration não termina com COMMIT;",
		)
	}

	content = strings.TrimSpace(
		strings.TrimPrefix(content, "BEGIN;"),
	)

	content = strings.TrimSpace(
		strings.TrimSuffix(content, "COMMIT;"),
	)

	if content == "" {
		return "", errors.New("migration vazia")
	}

	return content, nil
}
