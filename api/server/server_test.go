package server

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.epam.com/vadym_ulitin/lets-go-chat/pkg/hasher"
)

func TestHandleCreate(t *testing.T) {
	req, err := http.NewRequest("POST", "", bytes.NewBuffer([]byte(`{"userName":"john","password":"12345678"}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	var s server
	s.user = make(map[string]*userInfo)
	s.logger = *log.Default()

	rw := httptest.NewRecorder()
	handler := http.HandlerFunc(s.handleCreate())

	handler.ServeHTTP(rw, req)

	if status := rw.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expectedBody := `{"userName":"john","id":"0"}`

	if actualBody := rw.Body.String(); actualBody != expectedBody {
		t.Errorf("handler returned unexpected body: got %s want %s",
			actualBody, expectedBody)
	}

	expectedHeader := "application/json"

	if actualHeader := rw.Header().Get("Content-Type"); actualHeader != expectedHeader {
		t.Errorf("handler returned unexpected header: got %s want %s",
			actualHeader, expectedHeader)
	}
}

func TestHandleLogin(t *testing.T) {
	req, err := http.NewRequest("POST", "", bytes.NewBuffer([]byte(`{"userName":"john","password":"12345678"}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	var s server
	s.user = make(map[string]*userInfo)
	h, _ := hasher.HashPassword("12345678")
	s.user["john"] = &userInfo{"0", h, false}
	s.logger = *log.Default()

	rw := httptest.NewRecorder()
	handler := http.HandlerFunc(s.handleLogin())

	handler.ServeHTTP(rw, req)

	if status := rw.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expectedBody := []byte(`{"url":"ws://fancy-chat.io/ws\u0026token=one-time-token"}`)
	var actualBody []byte = make([]byte, rw.Body.Len())
	if rw.Body.Read(actualBody); string(actualBody) != string(expectedBody) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			string(actualBody), string(expectedBody))
	}

	expectedHeader := []string{"Content-Type", "X-Rate-Limit", "X-Expires-After"}

	for _, header := range expectedHeader {
		if rw.Header().Get(header) == "" {
			t.Errorf("handler doesn't contain expected header: %s",
				header)
		}
	}
}
