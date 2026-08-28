package service

import (
	"database/sql"
	"errors"
	"flitta/internal/config"
	"flitta/internal/repository"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidToken = errors.New("token inválido")

var ErrInvalidCredentials = errors.New(
	"email ou senha inválidos",
)

type authClaims struct {
	ClientID int `json:"client_id"`
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func AuthenticateUser(
	email string,
	password string,
) (string, error) {
	email = strings.ToLower(
		strings.TrimSpace(email),
	)

	clientID, hash, err :=
		repository.GetActiveUserCredentialsByEmail(email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidCredentials
		}

		return "", fmt.Errorf(
			"consultar usuário no login: %w",
			err,
		)
	}

	if !CheckPassword(password, hash) {
		return "", ErrInvalidCredentials
	}

	token, err := GenerateToken(clientID)
	if err != nil {
		return "", fmt.Errorf(
			"gerar token no login: %w",
			err,
		)
	}

	return token, nil
}

func GenerateToken(clientID int) (string, error) {
	now := time.Now()

	claims := authClaims{
		ClientID: clientID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(config.GetJWTSecret())
}

func ValidateToken(tokenString string) (int, error) {
	claims := &authClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return config.GetJWTSecret(), nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
		jwt.WithExpirationRequired(),
	)

	if err != nil || !token.Valid {
		return 0, ErrInvalidToken
	}

	if claims.ClientID <= 0 {
		return 0, ErrInvalidToken
	}

	return claims.ClientID, nil
}
