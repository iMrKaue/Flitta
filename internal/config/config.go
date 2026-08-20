package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Aviso: .env não carregado")
	}
}

func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		log.Fatal("JWT_SECRET não definido no ambiente")
	}

	return []byte(secret)
}

func GetAppEnv() string {
	return strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
}

func IsDevelopment() bool {
	return GetAppEnv() == "development"
}

func GetTwilioAuthToken() string {
	return strings.TrimSpace(os.Getenv("TWILIO_AUTH_TOKEN"))
}

func GetTwilioWebhookURL() string {
	return strings.TrimSpace(os.Getenv("TWILIO_WEBHOOK_URL"))
}
