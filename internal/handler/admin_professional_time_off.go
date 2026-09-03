package handler

import (
	"encoding/json"
	"errors"
	"flitta/internal/repository"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func CreateProfessionalTimeOffHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		w.Header().Set(
			"Allow",
			http.MethodPost,
		)

		http.Error(
			w,
			"método não permitido",
			http.StatusMethodNotAllowed,
		)

		return
	}

	clientID, ok :=
		professionalClientID(w, r)

	if !ok {
		return
	}

	var req struct {
		ProfessionalID int    `json:"professional_id"`
		StartAt        string `json:"start_at"`
		EndAt          string `json:"end_at"`
		Reason         string `json:"reason"`
	}

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {
		http.Error(
			w,
			"dados inválidos",
			http.StatusBadRequest,
		)

		return
	}

	req.StartAt = strings.TrimSpace(req.StartAt)
	req.EndAt = strings.TrimSpace(req.EndAt)
	req.Reason = strings.TrimSpace(req.Reason)

	if req.ProfessionalID <= 0 {
		http.Error(
			w,
			"id do profissional é obrigatório",
			http.StatusBadRequest,
		)

		return
	}

	if req.StartAt == "" || req.EndAt == "" {
		http.Error(
			w,
			"início e fim da indisponibilidade são obrigatórios",
			http.StatusBadRequest,
		)

		return
	}

	if len(req.Reason) > 255 {
		http.Error(
			w,
			"motivo da indisponibilidade muito longo",
			http.StatusBadRequest,
		)

		return
	}

	timeOff, err :=
		repository.CreateProfessionalTimeOff(
			clientID,
			req.ProfessionalID,
			req.StartAt,
			req.EndAt,
			req.Reason,
		)

	if errors.Is(
		err,
		repository.ErrProfessionalNotFound,
	) {
		http.Error(
			w,
			"profissional não encontrado",
			http.StatusNotFound,
		)

		return
	}

	if errors.Is(
		err,
		repository.ErrProfessionalTimeOffInvalid,
	) {
		http.Error(
			w,
			"período de indisponibilidade inválido",
			http.StatusBadRequest,
		)

		return
	}

	if errors.Is(
		err,
		repository.ErrProfessionalTimeOffOverlap,
	) {
		http.Error(
			w,
			"já existe uma indisponibilidade nesse período",
			http.StatusConflict,
		)

		return
	}

	if err != nil {
		log.Printf(
			"erro ao criar indisponibilidade do profissional %d do estabelecimento %d: %v",
			req.ProfessionalID,
			clientID,
			err,
		)

		http.Error(
			w,
			"erro ao criar indisponibilidade do profissional",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).
		Encode(timeOff)
}

func ListProfessionalTimeOffHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		w.Header().Set(
			"Allow",
			http.MethodGet,
		)

		http.Error(
			w,
			"método não permitido",
			http.StatusMethodNotAllowed,
		)

		return
	}

	clientID, ok :=
		professionalClientID(w, r)

	if !ok {
		return
	}

	rawProfessionalID := strings.TrimSpace(
		r.URL.Query().Get(
			"professional_id",
		),
	)

	professionalID, err :=
		strconv.Atoi(rawProfessionalID)

	if err != nil || professionalID <= 0 {
		http.Error(
			w,
			"id do profissional é obrigatório",
			http.StatusBadRequest,
		)

		return
	}

	timeOffs, err :=
		repository.ListProfessionalTimeOff(
			clientID,
			professionalID,
		)

	if errors.Is(
		err,
		repository.ErrProfessionalNotFound,
	) {
		http.Error(
			w,
			"profissional não encontrado",
			http.StatusNotFound,
		)

		return
	}

	if err != nil {
		log.Printf(
			"erro ao listar indisponibilidades do profissional %d do estabelecimento %d: %v",
			professionalID,
			clientID,
			err,
		)

		http.Error(
			w,
			"erro ao listar indisponibilidades do profissional",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	_ = json.NewEncoder(w).
		Encode(map[string]interface{}{
			"professional_id": professionalID,
			"time_offs":       timeOffs,
		})
}

func DeleteProfessionalTimeOffHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodDelete {
		w.Header().Set(
			"Allow",
			http.MethodDelete,
		)

		http.Error(
			w,
			"método não permitido",
			http.StatusMethodNotAllowed,
		)

		return
	}

	clientID, ok :=
		professionalClientID(w, r)

	if !ok {
		return
	}

	rawTimeOffID := strings.TrimSpace(
		r.URL.Query().Get(
			"time_off_id",
		),
	)

	timeOffID, err :=
		strconv.Atoi(rawTimeOffID)

	if err != nil || timeOffID <= 0 {
		http.Error(
			w,
			"id da indisponibilidade é obrigatório",
			http.StatusBadRequest,
		)

		return
	}

	err = repository.DeleteProfessionalTimeOff(
		clientID,
		timeOffID,
	)

	if errors.Is(
		err,
		repository.ErrProfessionalTimeOffNotFound,
	) {
		http.Error(
			w,
			"indisponibilidade do profissional não encontrada",
			http.StatusNotFound,
		)

		return
	}

	if err != nil {
		log.Printf(
			"erro ao remover indisponibilidade %d do estabelecimento %d: %v",
			timeOffID,
			clientID,
			err,
		)

		http.Error(
			w,
			"erro ao remover indisponibilidade do profissional",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	_ = json.NewEncoder(w).
		Encode(map[string]string{
			"message": "indisponibilidade removida com sucesso",
		})
}
