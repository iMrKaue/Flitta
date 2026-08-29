package handler

import (
	"encoding/json"
	"errors"
	"flitta/internal/service"
	"io"
	"log"
	"net/http"
	"strings"
)

type RegisterRequest struct {
	ResponsibleName string `json:"responsible_name"`
	BusinessName    string `json:"business_name"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	BusinessType    string `json:"business_type"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Método não permitido",
			http.StatusMethodNotAllowed,
		)
		return
	}

	var req RegisterRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(
			w,
			"Requisição inválida",
			http.StatusBadRequest,
		)
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		http.Error(
			w,
			"Requisição inválida",
			http.StatusBadRequest,
		)
		return
	}

	req.ResponsibleName = strings.TrimSpace(req.ResponsibleName)
	req.BusinessName = strings.TrimSpace(req.BusinessName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Email = strings.ToLower(
		strings.TrimSpace(req.Email),
	)
	req.BusinessType = strings.TrimSpace(
		req.BusinessType,
	)

	if req.ResponsibleName == "" ||
		req.BusinessName == "" ||
		req.Phone == "" ||
		req.Email == "" ||
		req.Password == "" ||
		req.BusinessType == "" {
		http.Error(
			w,
			"Todos os campos são obrigatórios",
			http.StatusBadRequest,
		)
		return
	}

	if len(req.Password) < 8 {
		http.Error(
			w,
			"A senha deve possuir pelo menos 8 caracteres",
			http.StatusBadRequest,
		)
		return
	}

	token, err := service.RegisterClient(
		req.BusinessName,
		req.ResponsibleName,
		req.Phone,
		req.Email,
		req.Password,
		req.BusinessType,
	)

	if err != nil {
		switch {
		case errors.Is(
			err,
			service.ErrPhoneAlreadyRegistered,
		):
			http.Error(
				w,
				"Telefone já cadastrado",
				http.StatusConflict,
			)

		case errors.Is(
			err,
			service.ErrEmailAlreadyRegistered,
		):
			http.Error(
				w,
				"Email já cadastrado",
				http.StatusConflict,
			)

		default:
			log.Printf(
				"erro ao cadastrar cliente: %v",
				err,
			)

			http.Error(
				w,
				"Erro interno",
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(
		map[string]string{
			"message": "Usuário criado com sucesso",
			"token":   token,
		},
	); err != nil {
		log.Printf(
			"erro ao escrever resposta de cadastro: %v",
			err,
		)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Método não permitido",
			http.StatusMethodNotAllowed,
		)
		return
	}

	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"Requisição inválida",
			http.StatusBadRequest,
		)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Email == "" || req.Password == "" {
		http.Error(
			w,
			"Email e senha são obrigatórios",
			http.StatusBadRequest,
		)
		return
	}

	token, err := service.AuthenticateUser(
		req.Email,
		req.Password,
	)

	if err != nil {
		switch {
		case errors.Is(
			err,
			service.ErrInvalidCredentials,
		):
			http.Error(
				w,
				"Email ou senha inválidos",
				http.StatusUnauthorized,
			)

		case errors.Is(
			err,
			service.ErrClientAccessBlocked,
		):
			http.Error(
				w,
				"Acesso ao estabelecimento suspenso ou expirado",
				http.StatusForbidden,
			)

		default:
			log.Printf(
				"Erro ao autenticar usuário: %v",
				err,
			)

			http.Error(
				w,
				"Erro interno",
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
