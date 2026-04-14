package handler

import (
	"encoding/json"
	"flitta/internal/middleware"
	"flitta/internal/repository"
	"flitta/internal/service"
	"fmt"
	"net/http"
)

type CreateServiceRequest struct {
	ClientID int    `json:"client_id"`
	Name     string `json:"name"`
}

type WorkingHoursRequest struct {
	ClientID int    `json:"client_id"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Interval int    `json:"interval"`
}

func CreateServiceHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.Context().Value(middleware.ClientIDKey).(int)

	var req struct {
		Name string `json:"name"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	err := repository.CreateService(clientID, req.Name)
	if err != nil {
		http.Error(w, "Erro ao criar serviço", 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Criado com sucesso",
	})
}

func GetServicesHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.Context().Value(middleware.ClientIDKey).(int)

	services := repository.GetServices(clientID)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(services)
}

func DeleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.Context().Value(middleware.ClientIDKey).(int)
	name := r.URL.Query().Get("name")

	err := repository.DeleteService(clientID, name)
	if err != nil {
		http.Error(w, "Error deleting service", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Deleted successfully",
	})
}

func SetWorkingHoursHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.Context().Value(middleware.ClientIDKey).(int)

	var req struct {
		Start    string `json:"start"`
		End      string `json:"end"`
		Interval int    `json:"interval"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Erro ao ler dados", 400)
		return
	}

	err = repository.SetWorkingHours(clientID, req.Start, req.End, req.Interval)
	if err != nil {
		fmt.Println("ERRO REAL:", err)
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"messsage": "Horários atualizados",
	})
}

func DashboardToday(w http.ResponseWriter, r *http.Request) {

	phone := r.URL.Query().Get("phone")

	clientID, err := service.GetClientByPhone(phone)
	fmt.Println("PHONE RECEBIDO:", phone)
	if err != nil {
		http.Error(w, "Cliente não encontrado", 404)
		return
	}

	list, err := repository.GetTodayAppointments(clientID)
	if err != nil {
		http.Error(w, "Erro ao buscar agenda", 500)
		return
	}

	json.NewEncoder(w).Encode(list)
}
