package config

import (
	"log"
	"os"

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
