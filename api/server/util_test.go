package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"testing"
)

var jsonData = []byte(`{"userName":"john","password":"12345678"}`)
var jsonDataShort = []byte(`{"userName":"joe","password":"12345678"}`)

func TestValidateRequest(t *testing.T) {
	// Valid request
	request, err := http.NewRequest("POST", "", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	request.Header.Set("Content-Type", "application/json")
	obj, err := validateRequest(request)

	if err != nil {
		t.Errorf("validateRequest() doesn't expect %v error\n", err)
	}

	b, _ := json.Marshal(obj)

	if string(b) != string(jsonData) {
		t.Errorf("validateRequest() expect %s got %s\n", jsonData, b)
	}

	// Unsupported method
	request, err = http.NewRequest("GET", "", bytes.NewBuffer(jsonData))

	if err != nil {
		log.Fatal(err)
	}

	_, err = validateRequest(request)

	if err != ErrUnsupportedMethod {
		t.Errorf("validateRequest() expects %v error, got %v\n", ErrUnsupportedMethod, err)
	}

	// Unsupported content type
	request, err = http.NewRequest("POST", "", bytes.NewBuffer(jsonData))

	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "text/html")

	_, err = validateRequest(request)

	if err != ErrUnsupportedContentType {
		t.Errorf("validateRequest() expects %v error, got %v\n", ErrUnsupportedContentType, err)
	}

	// Empty user name or password
	request, err = http.NewRequest("POST", "", bytes.NewBuffer(jsonDataShort))

	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")

	_, err = validateRequest(request)

	if err != ErrEmptyUsernameOrId {
		t.Errorf("validateRequest() expects %v error, got %v\n", ErrEmptyUsernameOrId, err)
	}
}
