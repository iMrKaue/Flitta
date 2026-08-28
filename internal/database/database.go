package database

import (
	"context"
	"database/sql"
	"errors"
	"flitta/internal/config"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

type databaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func envOr(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value != "" {
		return value
	}

	return fallback
}

func requiredEnv(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("%s é obrigatório", key)
	}

	return value, nil
}

func loadDatabaseConfig() (databaseConfig, error) {
	if config.IsDevelopment() {
		return databaseConfig{
			Host:     envOr("POSTGRES_HOST", "localhost"),
			Port:     envOr("POSTGRES_PORT", "5435"),
			User:     envOr("POSTGRES_USER", "postgres"),
			Password: envOr("POSTGRES_PASSWORD", "postgres"),
			Name:     envOr("POSTGRES_DB", "flitta"),
			SSLMode:  envOr("POSTGRES_SSLMODE", "disable"),
		}, nil
	}

	host, err := requiredEnv("POSTGRES_HOST")
	if err != nil {
		return databaseConfig{}, err
	}

	port, err := requiredEnv("POSTGRES_PORT")
	if err != nil {
		return databaseConfig{}, err
	}

	user, err := requiredEnv("POSTGRES_USER")
	if err != nil {
		return databaseConfig{}, err
	}

	password, err := requiredEnv("POSTGRES_PASSWORD")
	if err != nil {
		return databaseConfig{}, err
	}

	name, err := requiredEnv("POSTGRES_DB")
	if err != nil {
		return databaseConfig{}, err
	}

	sslMode, err := requiredEnv("POSTGRES_SSLMODE")
	if err != nil {
		return databaseConfig{}, err
	}

	if strings.EqualFold(sslMode, "disable") {
		return databaseConfig{}, errors.New(
			"POSTGRES_SSLMODE=disable não é permitido fora de development",
		)
	}

	return databaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     name,
		SSLMode:  sslMode,
	}, nil
}

func ConnectDB() {
	cfg, err := loadDatabaseConfig()
	if err != nil {
		log.Fatal("Configuração inválida do PostgreSQL: ", err)
	}

	connectionURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, cfg.Port),
		Path:   cfg.Name,
	}

	query := connectionURL.Query()
	query.Set("sslmode", cfg.SSLMode)
	connectionURL.RawQuery = query.Encode()

	DB, err = sql.Open("postgres", connectionURL.String())
	if err != nil {
		log.Fatal("Erro ao preparar conexão com PostgreSQL: ", err)
	}

	pingCtx, cancelPing := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	if err := DB.PingContext(pingCtx); err != nil {
		cancelPing()
		log.Fatal("Erro ao conectar no PostgreSQL: ", err)
	}

	cancelPing()

	migrationCtx, cancelMigrations := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)
	defer cancelMigrations()

	if err := runMigrations(migrationCtx); err != nil {
		log.Fatal("Erro ao executar migrations: ", err)
	}

	log.Println("PostgreSQL conectado ✅")
}
