package handler

import (
	"encoding/json"
	"flitta/internal/database"
	"flitta/internal/middleware"
	"flitta/internal/model"
	"flitta/internal/repository"
	"flitta/internal/service"
	"flitta/internal/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	clientID, ok := r.Context().Value(middleware.ClientIDKey).(int)
	if !ok {
		http.Error(w, "cliente não identificado", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name     string  `json:"name"`
		Duration int     `json:"duration"`
		Price    float64 `json:"price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "nome do serviço é obrigatório", http.StatusBadRequest)
		return
	}

	if req.Duration <= 0 {
		req.Duration = 30
	}

	if req.Price < 0 {
		http.Error(w, "preço do serviço não pode ser negativo", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`
		INSERT INTO services (client_id, name, duration, price)
		VALUES ($1, $2, $3, $4)
	`, clientID, req.Name, req.Duration, req.Price)

	if err != nil {
		http.Error(w, "erro ao criar serviço", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "serviço criado com sucesso",
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

func isValidHour(value string) bool {
	_, err := time.Parse("15:04", value)
	return err == nil
}

func parseHour(value string) (time.Time, error) {
	return time.Parse("15:04", value)
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
		http.Error(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}

	if req.Start == "" || req.End == "" || req.Interval <= 0 {
		http.Error(w, "Preencha início, fim e intervalo", http.StatusBadRequest)
		return
	}

	if !isValidHour(req.Start) || !isValidHour(req.End) {
		http.Error(w, "Use horários no formato HH:MM, exemplo: 09:00", http.StatusBadRequest)
		return
	}

	startTime, err := parseHour(req.Start)
	if err != nil {
		http.Error(w, "Horário de início inválido", http.StatusBadRequest)
		return
	}

	endTime, err := parseHour(req.End)
	if err != nil {
		http.Error(w, "Horário de fim inválido", http.StatusBadRequest)
		return
	}

	if !startTime.Before(endTime) {
		http.Error(w, "O horário de início deve ser menor que o horário de fim", http.StatusBadRequest)
		return
	}

	if req.Interval < 15 {
		http.Error(w, "O intervalo mínimo deve ser de 15 minutos", http.StatusBadRequest)
		return
	}

	if req.Interval > 240 {
		http.Error(w, "O intervalo não pode ser maior que 240 minutos", http.StatusBadRequest)
		return
	}

	err = repository.SetWorkingHours(clientID, req.Start, req.End, req.Interval)
	if err != nil {
		fmt.Println("ERRO REAL:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Horários atualizados",
	})
}

func GetWorkingHourHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.Context().Value(middleware.ClientIDKey).(int)

	start, end, interval, err := repository.GetWorkingHours(clientID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"configured": false,
			"start":      "",
			"end":        "",
			"interval":   0,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"configured": true,
		"start":      start,
		"end":        end,
		"interval":   interval,
	})
}

func GetCompanySettingsHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.Context().Value(middleware.ClientIDKey).(int)

	client, err := repository.GetClientSettings(clientID)
	if err != nil {
		http.Error(w, "empresa não encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}

func UpdateCompanySettingsHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.Context().Value(middleware.ClientIDKey).(int)

	var req struct {
		Name         string `json:"name"`
		Phone        string `json:"phone"`
		BusinessType string `json:"business_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Phone = strings.TrimSpace(req.Phone)
	req.BusinessType = strings.TrimSpace(req.BusinessType)

	if req.Name == "" {
		http.Error(w, "nome da empresa é obrigatório", http.StatusBadRequest)
		return
	}

	if req.Phone == "" {
		http.Error(w, "telefone da empresa é obrigatório", http.StatusBadRequest)
		return
	}

	phone := utils.NormalizeCustomerPhone(req.Phone)
	businessType := utils.NormalizeBusinessType(req.BusinessType)

	err := repository.UpdateClientSettings(clientID, req.Name, phone, businessType)
	if err != nil {
		http.Error(w, "erro ao atualizar dados da empresa", http.StatusInternalServerError)
		return
	}

	client, err := repository.GetClientSettings(clientID)
	if err != nil {
		http.Error(w, "erro ao buscar dados atualizados", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
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

type ReminderResponse struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Service       string `json:"service"`
	Date          string `json:"date"`
	Time          string `json:"time"`
	CustomerPhone string `json:"customer_phone"`
	Message       string `json:"message"`
}

func formatReminderDate(date string) string {
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		return date
	}

	return parts[2] + "/" + parts[1] + "/" + parts[0]
}

func buildReminderMessage(companyName string, appointment model.Appointment) string {
	return fmt.Sprintf(
		"Olá, %s! 😊\nPassando para lembrar do seu horário em %s.\n\n📌 Serviço: %s\n📅 Data: %s\n🕒 Horário: %s\n\nSe precisar remarcar ou cancelar, responda esta mensagem.",
		appointment.Name,
		companyName,
		appointment.Service,
		formatReminderDate(appointment.Date),
		appointment.Time,
	)
}

func GetPendingRemindersHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.Context().Value(middleware.ClientIDKey).(int)

	hoursBefore := 24

	if rawHours := r.URL.Query().Get("hours"); rawHours != "" {
		parsedHours, err := strconv.Atoi(rawHours)
		if err == nil && parsedHours > 0 {
			hoursBefore = parsedHours
		}
	}

	client, err := repository.GetClientSettings(clientID)
	if err != nil {
		http.Error(w, "empresa não encontrada", http.StatusNotFound)
		return
	}

	appointments, err := repository.GetPendingReminderAppointments(clientID, hoursBefore)
	if err != nil {
		http.Error(w, "erro ao buscar lembretes pendentes", http.StatusInternalServerError)
		return
	}

	var reminders []ReminderResponse

	for _, appointment := range appointments {
		reminders = append(reminders, ReminderResponse{
			ID:            appointment.ID,
			Name:          appointment.Name,
			Service:       appointment.Service,
			Date:          appointment.Date,
			Time:          appointment.Time,
			CustomerPhone: appointment.CustomerPhone,
			Message:       buildReminderMessage(client.Name, appointment),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reminders)
}

func MarkReminderSentHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.Context().Value(middleware.ClientIDKey).(int)

	var req struct {
		AppointmentID int `json:"appointment_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}

	if req.AppointmentID <= 0 {
		http.Error(w, "appointment_id é obrigatório", http.StatusBadRequest)
		return
	}

	err := repository.MarkReminderAsSent(req.AppointmentID, clientID)
	if err != nil {
		http.Error(w, "erro ao marcar lembrete como enviado", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "lembrete marcado como enviado",
	})
}

func SendReminderHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.Context().Value(middleware.ClientIDKey).(int)

	var req struct {
		AppointmentID int `json:"appointment_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}

	if req.AppointmentID <= 0 {
		http.Error(w, "appointment_id é obrigatório", http.StatusBadRequest)
		return
	}

	client, err := repository.GetClientSettings(clientID)
	if err != nil {
		http.Error(w, "empresa não encontrado", http.StatusNotFound)
		return
	}

	appointment, err := repository.GetAppointmentByID(req.AppointmentID, clientID)
	if err != nil {
		http.Error(w, "agendamento não encontrado", http.StatusNotFound)
		return
	}

	if appointment.ReminderSent {
		http.Error(w, "lembrete já foi enviado", http.StatusBadRequest)
		return
	}

	message := buildReminderMessage(client.Name, appointment)

	result, err := service.SendWhatsAppMessage(appointment.CustomerPhone, message)
	if err != nil {
		http.Error(w, "erro ao enviar lembrete: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = repository.MarkReminderAsSent(req.AppointmentID, clientID)
	if err != nil {
		http.Error(w, "lembrete enviado, mas erro ao atualizar status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "lembrete enviado com sucesso",
		"provider":   result.Provider,
		"to":         result.To,
		"message_id": result.MessageID,
		"simulated":  result.Simulated,
	})
}
