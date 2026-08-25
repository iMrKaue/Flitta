package handler

import (
	"context"
	"errors"
	"flitta/internal/database"
	"flitta/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDeleteServiceHandlerMethodNotAllowed(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/admin/service/delete?name=Corte",
		nil,
	)

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			middleware.ClientIDKey,
			1,
		),
	)

	recorder := httptest.NewRecorder()

	DeleteServiceHandler(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusMethodNotAllowed,
			recorder.Code,
		)
	}

	if recorder.Header().Get("Allow") != http.MethodDelete {
		t.Fatalf(
			"header Allow esperado %q, recebido %q",
			http.MethodDelete,
			recorder.Header().Get("Allow"),
		)
	}
}

func TestDeleteServiceHandlerRequiresName(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodDelete,
		"/admin/service/delete",
		nil,
	)

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			middleware.ClientIDKey,
			1,
		),
	)

	recorder := httptest.NewRecorder()

	DeleteServiceHandler(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestDeleteServiceHandlerReturnsNotFoundWhenNothingWasDeleted(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro ao criar sqlmock: %v", err)
	}
	defer db.Close()

	previousDB := database.DB
	database.DB = db
	defer func() {
		database.DB = previousDB
	}()

	mock.ExpectExec(`DELETE FROM services`).
		WithArgs(2, "Alpha Exclusive").
		WillReturnResult(sqlmock.NewResult(0, 0))

	request := httptest.NewRequest(
		http.MethodDelete,
		"/admin/service/delete?name=Alpha%20Exclusive",
		nil,
	)

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			middleware.ClientIDKey,
			2,
		),
	)

	recorder := httptest.NewRecorder()

	DeleteServiceHandler(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas SQL não atendidas: %v", err)
	}
}

func TestDeleteServiceHandlerDeletesOwnService(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro ao criar sqlmock: %v", err)
	}
	defer db.Close()

	previousDB := database.DB
	database.DB = db
	defer func() {
		database.DB = previousDB
	}()

	mock.ExpectExec(`DELETE FROM services`).
		WithArgs(1, "Alpha Exclusive").
		WillReturnResult(sqlmock.NewResult(0, 1))

	request := httptest.NewRequest(
		http.MethodDelete,
		"/admin/service/delete?name=Alpha%20Exclusive",
		nil,
	)

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			middleware.ClientIDKey,
			1,
		),
	)

	recorder := httptest.NewRecorder()

	DeleteServiceHandler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas SQL não atendidas: %v", err)
	}
}

func TestDeleteServiceHandlerHandlesDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro ao criar sqlmock: %v", err)
	}
	defer db.Close()

	previousDB := database.DB
	database.DB = db
	defer func() {
		database.DB = previousDB
	}()

	mock.ExpectExec(`DELETE FROM services`).
		WithArgs(1, "Corte").
		WillReturnError(errors.New("database unavailable"))

	request := httptest.NewRequest(
		http.MethodDelete,
		"/admin/service/delete?name=Corte",
		nil,
	)

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			middleware.ClientIDKey,
			1,
		),
	)

	recorder := httptest.NewRecorder()

	DeleteServiceHandler(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas SQL não atendidas: %v", err)
	}
}
