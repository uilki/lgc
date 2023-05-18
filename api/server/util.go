package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Types
type userInfo struct {
	id     string
	hash   string
	logged bool
}

type UserCreds struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}

type CreateUserRequest UserCreds

type CreateUserResponse struct {
	UserName string `json:"userName"`
	Id       string `json:"id"`
}

type LoginUserRequest UserCreds

type LoginUserResponse struct {
	Url string `json:"url"`
}

// Errors
var (
	ErrInternalServerError       = errors.New("internal server error")
	ErrEmptyUsernameOrId         = errors.New("bad request, empty username or id")
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
	ErrUnsupportedContentType    = errors.New("unsupported content type")
	ErrUnsupportedMethod         = errors.New("unsupported method")
	ErrUserAlreadyRegistered     = errors.New("user already registered")
	ErrUserAlreadyLoggedIn       = errors.New("user already logged in")
)

func marshalValue[T any](v T) ([]byte, error) {
	return json.Marshal(v)
}

func unmarshalCreds(body io.ReadCloser) (UserCreds, error) {
	b, err := io.ReadAll(body)

	if err != nil {
		return UserCreds{}, ErrInternalServerError
	}

	var r UserCreds
	err = json.Unmarshal(b, &r)

	if err != nil {
		return UserCreds{}, ErrInternalServerError
	}

	return r, nil
}

func responseFailure(w http.ResponseWriter, err error) {
	var status int
	switch err {
	case ErrUserAlreadyLoggedIn, ErrInvalidUsernameOrPassword, ErrUnsupportedMethod, ErrUnsupportedContentType, ErrEmptyUsernameOrId:
		status = http.StatusBadRequest
	case ErrUserAlreadyRegistered, ErrInternalServerError:
		status = http.StatusInternalServerError
	default:
		log.Fatalln("Unknown error")
	}
	w.WriteHeader(status)
	fmt.Fprint(w, err)
}

func validateRequest(r *http.Request) (UserCreds, error) {
	if r.Method != "POST" {
		return UserCreds{}, ErrUnsupportedMethod
	}

	if r.Header.Get("Content-Type") != "application/json" {
		return UserCreds{}, ErrUnsupportedContentType
	}

	defer r.Body.Close()

	obj, err := unmarshalCreds(r.Body)
	if err != nil {
		return UserCreds{}, err
	}

	if len(obj.UserName) < 4 || len(obj.Password) < 8 {
		return UserCreds{}, ErrEmptyUsernameOrId
	}

	return obj, nil
}

func getXRLimit() string {
	return strconv.Itoa(100)
}

func getXExpiresAfter() string {
	return time.Unix(time.Now().Unix(), 0).Add(time.Hour * 3).UTC().String()
}
