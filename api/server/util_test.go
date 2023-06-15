package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var jsonData = []byte(`{"userName":"john","password":"12345678"}`)
var jsonDataShort = []byte(`{"userName":"joe","password":"12345678"}`)

func TestResponceFailure(t *testing.T) {
	w := httptest.NewRecorder()

	responseFailure(w, ErrInvalidOneTimeToken)

	body, _ := io.ReadAll(w.Result().Body)
	if w.Code != http.StatusBadRequest || string(body) != ErrInvalidOneTimeToken.Error() {
		t.Errorf(
			"responseFailure()\nexpect\n%d status code, with \"%s\" body\ngot\n%d status code with \"%s\" message\n",
			http.StatusBadRequest,
			ErrInvalidOneTimeToken.Error(),
			w.Code,
			string(body),
		)
	}
}
func TestValidateRequest(t *testing.T) {
	request, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	// Valid request
	request.Header.Set("Content-Type", "application/json")
	obj, err := validateRequest(request)

	if err != nil {
		t.Errorf("validateRequest() doesn't expect %v error\n", err)
	}

	b, _ := json.Marshal(obj)

	if string(b) != string(jsonData) {
		t.Errorf("validateRequest() expect %s got %s\n", jsonData, b)
	}

	// Empty user name or password
	request, err = http.NewRequest(http.MethodPost, "", bytes.NewBuffer(jsonDataShort))

	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")

	_, err = validateRequest(request)

	if err != ErrEmptyUsernameOrId {
		t.Errorf("validateRequest() expects %v error, got %v\n", ErrEmptyUsernameOrId, err)
	}

	// Empty body
	request, err = http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte{}))

	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")

	_, err = validateRequest(request)

	if err == nil {
		t.Errorf("validateRequest() expects error, got %v\n", nil)
	}
}

func TestValidateUpgrade(t *testing.T) {
	request, err := http.NewRequest(http.MethodGet, "http://host.net/v1/ws?token=ea98117f1358381915f6daf23225c1fd07192c73961f3c27eecc5690525adcc9ecc1330c5e2828c5076f3a28ffd8231918d966df43405855a98b24e3b84e9997", nil)
	if err != nil {
		log.Fatal(err)
	}

	// Valid request
	_, err = validateUpgrade(request)

	if err != nil {
		t.Errorf("validateUpgrade() doesn't expect %v error\n", err)
	}

	// Empty token
	request, err = http.NewRequest(http.MethodGet, "http://host.net/v1/ws", nil)
	if err != nil {
		log.Fatal(err)
	}

	_, err = validateUpgrade(request)

	if err != ErrInvalidOneTimeToken {
		t.Errorf("validateUpgrade() expect %v error, got %v\n", ErrInvalidOneTimeToken, err)
	}
}
