package handler

import (
	"encoding/json"
	"flitta/internal/middleware"
	"flitta/internal/repository"
	"net/http"
)

func GetIntelligenceSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	clientID, ok := r.Context().Value(middleware.ClientIDKey).(int)
	if !ok {
		http.Error(w, "cliente não identificado", http.StatusUnauthorized)
		return
	}

	summary, err := repository.GetIntelligenceSummary(clientID)
	if err != nil {
		http.Error(w, "erro ao buscar indicadores", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func GetServicePerformanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	clientID, ok := r.Context().Value(middleware.ClientIDKey).(int)
	if !ok {
		http.Error(w, "cliente não identificado", http.StatusUnauthorized)
		return
	}

	services, err := repository.GetServicePerformance(clientID)
	if err != nil {
		http.Error(w, "erro ao buscar indicadores por serviço", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}
