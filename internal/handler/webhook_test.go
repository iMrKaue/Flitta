package handler

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteTwiMLResponseEscapesXMLCharacters(t *testing.T) {
	message := `Olá João & Maria <3 > teste "aspas" 'apóstrofo'`

	recorder := httptest.NewRecorder()

	if err := writeTwiMLResponse(recorder, message); err != nil {
		t.Fatalf("writeTwiMLResponse retornou erro: %v", err)
	}

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"status esperado %d, recebido %d",
			http.StatusOK,
			response.StatusCode,
		)
	}

	contentType := response.Header.Get("Content-Type")
	if contentType != "application/xml; charset=utf-8" {
		t.Fatalf(
			"Content-Type inesperado: %q",
			contentType,
		)
	}

	body := recorder.Body.String()

	if !strings.Contains(body, "&amp;") {
		t.Fatalf(
			"caractere & não foi escapado: %s",
			body,
		)
	}

	if !strings.Contains(body, "&lt;") {
		t.Fatalf(
			"caractere < não foi escapado: %s",
			body,
		)
	}

	var decoded twimlResponse

	if err := xml.Unmarshal(
		recorder.Body.Bytes(),
		&decoded,
	); err != nil {
		t.Fatalf(
			"TwiML gerado não é XML válido: %v",
			err,
		)
	}

	if decoded.XMLName.Local != "Response" {
		t.Fatalf(
			"elemento raiz inesperado: %q",
			decoded.XMLName.Local,
		)
	}

	if decoded.Message != message {
		t.Fatalf(
			"mensagem mudou após serialização: esperado %q, recebido %q",
			message,
			decoded.Message,
		)
	}
}
