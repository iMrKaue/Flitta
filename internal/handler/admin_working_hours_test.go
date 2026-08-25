package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetWorkingHoursHandlerMethodNotAllowed(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/admin/hours/set",
		nil,
	)

	recorder := httptest.NewRecorder()

	SetWorkingHoursHandler(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusMethodNotAllowed,
			recorder.Code,
		)
	}

	if recorder.Header().Get("Allow") != http.MethodPost {
		t.Fatalf(
			"header Allow esperado %q, recebido %q",
			http.MethodPost,
			recorder.Header().Get("Allow"),
		)
	}
}

func TestGetWorkingHourHandlerMethodNotAllowed(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodPost,
		"/admin/hours/get",
		nil,
	)

	recorder := httptest.NewRecorder()

	GetWorkingHourHandler(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusMethodNotAllowed,
			recorder.Code,
		)
	}

	if recorder.Header().Get("Allow") != http.MethodGet {
		t.Fatalf(
			"header Allow esperado %q, recebido %q",
			http.MethodGet,
			recorder.Header().Get("Allow"),
		)
	}
}
