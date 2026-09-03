package handler

import (
	"encoding/json"
	"errors"
	"flitta/internal/model"
	"flitta/internal/repository"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func ReplaceProfessionalWorkingHoursHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPut {
		w.Header().Set(
			"Allow",
			http.MethodPut,
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
		ProfessionalID int `json:"professional_id"`

		Hours *[]model.ProfessionalWorkingHour `json:"hours"`
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

	if req.ProfessionalID <= 0 {
		http.Error(
			w,
			"id do profissional é obrigatório",
			http.StatusBadRequest,
		)

		return
	}

	if req.Hours == nil {
		http.Error(
			w,
			"hours é obrigatório",
			http.StatusBadRequest,
		)

		return
	}

	err := repository.ReplaceProfessionalWorkingHours(
		clientID,
		req.ProfessionalID,
		*req.Hours,
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
		repository.ErrProfessionalWorkingHoursInvalid,
	) {
		http.Error(
			w,
			"horário do profissional inválido",
			http.StatusBadRequest,
		)

		return
	}

	if errors.Is(
		err,
		repository.ErrProfessionalWorkingHoursOverlap,
	) {
		http.Error(
			w,
			"existem horários sobrepostos para o profissional",
			http.StatusBadRequest,
		)

		return
	}

	if err != nil {
		log.Printf(
			"erro ao atualizar horários do profissional %d do estabelecimento %d: %v",
			req.ProfessionalID,
			clientID,
			err,
		)

		http.Error(
			w,
			"erro ao atualizar horários do profissional",
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
			"message": "horários do profissional atualizados com sucesso",
		})
}

func GetProfessionalWorkingHoursHandler(
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

	hours, err :=
		repository.ListProfessionalWorkingHours(
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
			"erro ao listar horários do profissional %d do estabelecimento %d: %v",
			professionalID,
			clientID,
			err,
		)

		http.Error(
			w,
			"erro ao listar horários do profissional",
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
			"hours":           hours,
		})
}
