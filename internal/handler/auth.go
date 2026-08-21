package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flitta/internal/database"
	"flitta/internal/repository"
	"flitta/internal/service"
	"flitta/internal/utils"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type RegisterRequest struct {
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	BusinessType string `json:"business_type"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 🔹 cria cliente via service (arquitetura correta)
	token, err := service.RegisterClient(
		req.Name,
		req.Phone,
		req.Email,
		req.Password,
		req.BusinessType,
	)

	if err != nil {
		fmt.Println("ERRO REGISTER:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 🔹 pega o client_id recém criado (precisamos para defaults)
	phone := utils.NormalizeCustomerPhone(req.Phone)

	clientID, err := service.GetClientByPhone(phone)
	if err != nil {
		http.Error(w, "Erro ao buscar cliente", http.StatusInternalServerError)
		return
	}

	// cria serviços padrão conforme o tipo de negócio
	defaultServices := getDefaultServicesByBusinessType(req.BusinessType)

	for _, s := range defaultServices {
		if err := repository.CreateService(clientID, s); err != nil {
			fmt.Println("ERRO service default:", s, err)
		}
	}

	// 🔹 define horário padrão
	if err := repository.SetWorkingHours(clientID, "09:00", "18:00", 60); err != nil {
		fmt.Println("ERRO working hours default:", err)
	}

	// 🔹 resposta final (já autenticado)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Usuário criado com sucesso",
		"token":   token,
	})
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

	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" || req.Password == "" {
		http.Error(
			w,
			"Email e senha são obrigatórios",
			http.StatusBadRequest,
		)
		return
	}

	var id int
	var hash string

	err := database.DB.QueryRow(`
		SELECT id, password
		FROM clients
		WHERE email = $1
	`, req.Email).Scan(&id, &hash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(
				w,
				"Email ou senha inválidos",
				http.StatusUnauthorized,
			)
			return
		}

		log.Printf("erro ao consultar cliente no login: %v", err)

		http.Error(
			w,
			"Erro interno",
			http.StatusInternalServerError,
		)
		return
	}

	if !service.CheckPassword(req.Password, hash) {
		http.Error(
			w,
			"Email ou senha inválidos",
			http.StatusUnauthorized,
		)
		return
	}

	token, err := service.GenerateToken(id)
	if err != nil {
		log.Printf("erro ao gerar token no login: %v", err)

		http.Error(
			w,
			"Erro interno",
			http.StatusInternalServerError,
		)
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

func getDefaultServicesByBusinessType(businessType string) []string {
	businessType = service.NormalizeBusinessType(businessType)

	switch businessType {
	case "beauty":
		return []string{"Corte", "Escova", "Progressiva"}
	case "barber":
		return []string{"Corte", "Barba", "Sobrancelha"}
	case "clinic":
		return []string{"Consulta", "Retorno", "Avaliação"}
	case "gym":
		return []string{"Avaliação física", "Aula experimental", "Personal"}
	case "petshop":
		return []string{"Banho", "Tosa", "Consulta"}
	default:
		return []string{"Atendimento"}
	}
}
