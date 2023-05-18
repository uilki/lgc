package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"git.epam.com/vadym_ulitin/lets-go-chat/pkg/hasher"
)

type server struct {
	url    string
	user   map[string]*userInfo
	uuid   int
	router Router
	logger log.Logger
}

var defaultServer server

func (s *server) routes() {
	s.router.HandleFunc("/v1/user", s.handleCreate())
	s.router.HandleFunc("/v1/user/login", s.handleLogin())
}

func (s *server) handleCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Create user request")
		obj, err := validateRequest(r)
		if err != nil {
			s.logger.Printf("%s\n", err)
			responseFailure(w, err)
			return
		}

		resp, err := s.addUser(obj.UserName, obj.Password)

		if err != nil {
			responseFailure(w, err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(resp))
	}
}

func (s *server) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Login user request")
		obj, err := validateRequest(r)
		if err != nil {
			s.logger.Printf("%s\n", err)
			responseFailure(w, err)
			return
		}

		resp, err := s.loginUser(obj.UserName, obj.Password)

		if err != nil {
			responseFailure(w, err)
			return
		}

		w.Header().Add("X-Rate-Limit", getXRLimit())
		w.Header().Add("X-Expires-After", getXExpiresAfter())
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(resp))
	}
}

func (s *server) loginUser(uname, passwd string) (respBody []byte, err error) {
	if _, ok := s.user[uname]; !ok || !hasher.CheckPasswordHash(passwd, s.user[uname].hash) {
		s.logger.Printf("User %s dosen't exist or password mismatch\n", uname)
		return nil, ErrInvalidUsernameOrPassword
	}

	if s.user[uname].logged {
		s.logger.Printf("User %s, err: %s\n", uname, ErrUserAlreadyLoggedIn)
		return nil, ErrUserAlreadyLoggedIn
	}

	if respBody, err = marshalValue(LoginUserResponse{"ws://fancy-chat.io/ws&token=one-time-token"}); err != nil {
		s.logger.Printf("Faled to marshal LoginUserResponse, err: %s\n", err)
		return nil, ErrInternalServerError
	}

	s.user[uname].logged = true
	return respBody, nil
}

func (s *server) addUser(uname, passwd string) (respBody []byte, err error) {
	h, err := hasher.HashPassword(passwd)
	if err != nil {
		s.logger.Printf("Failed to encrypt password, err: %s\n", err)
		return nil, ErrInternalServerError
	}

	if _, ok := s.user[uname]; ok {
		s.logger.Printf("User with name: %s, already exists\n", uname)
		return nil, ErrUserAlreadyRegistered
	}

	s.logger.Printf("Add user: %s, id: %d\n", uname, s.uuid)

	s.user[uname] = &userInfo{strconv.Itoa(s.uuid), h, false}
	s.uuid++

	if respBody, err = marshalValue(CreateUserResponse{uname, s.user[uname].id}); err != nil {
		s.logger.Printf("Faled to marshal CreateUserResponse, err: %s\n", err)
		return nil, ErrInternalServerError
	}

	return respBody, nil
}

func init() {
	f, err := os.OpenFile("chat-server-go.log", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal(err)
	}

	defaultServer.logger = *log.New(f, "server", log.Lshortfile)
	defaultServer.url = ":8080"
	defaultServer.user = make(map[string]*userInfo)
	defaultServer.router.routes = make(map[string]http.HandlerFunc)
	defaultServer.routes()
	defaultServer.logger.Println("Initialized")
}

func Run() error {
	for r, h := range defaultServer.router.routes {
		http.HandleFunc(r, h)
	}
	return http.ListenAndServe(defaultServer.url, nil)
}
