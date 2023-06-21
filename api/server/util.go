package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Types
type userInfo struct {
	name         string
	hash         string
	logged       bool
	oneTimeToken string
	expiresAfter time.Time
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

type ActiveUsersResponce struct {
	ActiveUsers []string `json:"activeusers"`
}

// Errors
var (
	ErrInternalServerError       = errors.New("internal server error")
	ErrEmptyUsernameOrId         = errors.New("empty username or id")
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
	ErrUserAlreadyRegistered     = errors.New("user already registered")
	ErrUserAlreadyLoggedIn       = errors.New("user already logged in")
	ErrInvalidOneTimeToken       = errors.New("invalid onetime token")
)

func marshalValue[T any](v T) ([]byte, error) {
	return json.Marshal(v)
}

func unmarshalCreds(body io.ReadCloser) (UserCreds, error) {
	b, err := io.ReadAll(body)

	if err != nil {
		return UserCreds{}, err
	}

	var r UserCreds
	err = json.Unmarshal(b, &r)

	if err != nil {
		return UserCreds{}, err
	}

	return r, nil
}

func responseFailure(w http.ResponseWriter, err interface{}) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, err)
}

func validateRequest(r *http.Request) (UserCreds, error) {
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

func validateUpgrade(r *http.Request) (string, error) {
	token := r.URL.Query().Get("token")
	if token == "" {
		return "", ErrInvalidOneTimeToken
	}

	return token, nil
}

func getXRLimit() string {
	return strconv.Itoa(100)
}

func getXExpiresAfter() time.Time {
	return time.Unix(time.Now().Unix(), 0).Add(time.Minute * 3).UTC()
}

func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
