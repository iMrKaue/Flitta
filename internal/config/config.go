package config

import (
	"log"
	"os"
)

func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		log.Fatal("JWT_SECRET não definido no ambiente")
	}

	return []byte(secret)
}
