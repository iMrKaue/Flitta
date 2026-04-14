package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func ConnectDB() {
	host := envOr("POSTGRES_HOST", "localhost")
	port := envOr("POSTGRES_PORT", "5435")
	user := envOr("POSTGRES_USER", "postgres")
	password := envOr("POSTGRES_PASSWORD", "postgres")
	dbname := envOr("POSTGRES_DB", "flitta")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Erro ao conectar no banco:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Erro ao pingar banco:", err)
	}

	runMigrations()

	log.Println("PostgreSQL conectado ✅")
}

func runMigrations() {
	stmts := []string{
		`ALTER TABLE appointments ADD COLUMN IF NOT EXISTS customer_phone VARCHAR(64)`,
		`ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS selected_appointment_id INTEGER NOT NULL DEFAULT 0`,
	}

	for _, q := range stmts {
		if _, err := DB.Exec(q); err != nil {
			log.Printf("migration: %v", err)
		}
	}
}
