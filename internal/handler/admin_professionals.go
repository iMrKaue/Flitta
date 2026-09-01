package handler

import (
	"encoding/json"
	"errors"
	"flitta/internal/middleware"
	"flitta/internal/repository"
	"log"
	"net/http"
	"strings"

	"github.com/lib/pq"
)

func professionalClientID(
	w http.ResponseWriter,
	r *http.Request,
) (int, bool) {
	clientID, ok := r.Context().
		Value(middleware.ClientIDKey).(int)

	if !ok {
		http.Error(
			w,
			"cliente não identificado",
			http.StatusUnauthorized,
		)

		return 0, false
	}

	return clientID, true
}

func CreateProfessionalHandler(
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
		Name string `json:"name"`
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

	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" {
		http.Error(
			w,
			"nome do profissional é obrigatório",
			http.StatusBadRequest,
		)

		return
	}

	if len(req.Name) > 150 {
		http.Error(
			w,
			"nome do profissional muito longo",
			http.StatusBadRequest,
		)

		return
	}

	professional, err :=
		repository.CreateProfessional(
			clientID,
			req.Name,
		)

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) &&
			pqErr.Code == "23505" &&
			pqErr.Constraint ==
				"professionals_client_name_normalized_unique" {

			http.Error(
				w,
				"já existe um profissional com esse nome",
				http.StatusConflict,
			)

			return
		}

		if errors.Is(
			err,
			repository.ErrProfessionalNameRequired,
		) {
			http.Error(
				w,
				err.Error(),
				http.StatusBadRequest,
			)

			return
		}

		log.Printf(
			"erro ao criar profissional do estabelecimento %d: %v",
			clientID,
			err,
		)

		http.Error(
			w,
			"erro ao criar profissional",
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
		Encode(professional)
}

func ListProfessionalsHandler(
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

	professionals, err :=
		repository.ListProfessionals(clientID)

	if err != nil {
		log.Printf(
			"erro ao listar profissionais do estabelecimento %d: %v",
			clientID,
			err,
		)

		http.Error(
			w,
			"erro ao listar profissionais",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	_ = json.NewEncoder(w).
		Encode(professionals)
}

func SetProfessionalActiveHandler(
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
		ProfessionalID int  `json:"professional_id"`
		Active         bool `json:"active"`
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

	err := repository.SetProfessionalActive(
		clientID,
		req.ProfessionalID,
		req.Active,
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
			"erro ao alterar profissional %d do estabelecimento %d: %v",
			req.ProfessionalID,
			clientID,
			err,
		)

		http.Error(
			w,
			"erro ao atualizar profissional",
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
			"message": "profissional atualizado com sucesso",
		})
}

func ReplaceProfessionalServicesHandler(
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
		ProfessionalID int   `json:"professional_id"`
		ServiceIDs     []int `json:"service_ids"`
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

	err := repository.ReplaceProfessionalServices(
		clientID,
		req.ProfessionalID,
		req.ServiceIDs,
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
		repository.ErrProfessionalServiceNotFound,
	) {
		http.Error(
			w,
			"serviço não encontrado no estabelecimento",
			http.StatusNotFound,
		)

		return
	}

	if err != nil {
		log.Printf(
			"erro ao atualizar serviços do profissional %d do estabelecimento %d: %v",
			req.ProfessionalID,
			clientID,
			err,
		)

		http.Error(
			w,
			"erro ao atualizar serviços do profissional",
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
			"message": "serviços do profissional atualizados com sucesso",
		})
}
