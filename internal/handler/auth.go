package handler

import (
	"encoding/json"
	"flitta/internal/database"
	"flitta/internal/repository"
	"flitta/internal/service"
	"flitta/internal/utils"
	"fmt"
	"net/http"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
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

	// 🔹 cria serviços padrão
	defaultServices := []string{"Corte", "Escova", "Progressiva"}

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
	var req LoginRequest

	json.NewDecoder(r.Body).Decode(&req)

	var id int
	var hash string

	err := database.DB.QueryRow(`
		SELECT id, password FROM clients WHERE email = $1
	`, req.Email).Scan(&id, &hash)

	if err != nil {
		http.Error(w, "Usuário não encontrado", 401)
		return
	}

	if !service.CheckPassword(req.Password, hash) {
		http.Error(w, "Senha inválida", 401)
		return
	}

	token, _ := service.GenerateToken(id)

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
