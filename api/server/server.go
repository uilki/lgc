package server

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"git.epam.com/vadym_ulitin/lets-go-chat/pkg/hasher"
	"git.epam.com/vadym_ulitin/lets-go-chat/pkg/logger"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	ERROR = iota
	INFO
	DEBUG
)

type Server struct {
	url        string
	user       map[uuid.UUID]*userInfo
	router     Router
	history    backlog
	controller *dispatcher
	logger     logger.Logger
}

var defaultServer Server

func (s *Server) routes() {
	s.router.HandleFunc("/v1/user", http.MethodPost, []string{"Content-Type", "application/json"}, s.handleCreate())
	s.router.HandleFunc("/v1/user/login", http.MethodPost, []string{"Content-Type", "application/json"}, s.handleLogin())
	s.router.HandleFunc("/v1/ws", http.MethodGet, nil, s.handleConnect())
	s.router.HandleFunc("/v1/users", http.MethodGet, nil, s.handleActiveUsers())
}

func (s *Server) log(level int, message string, arg ...any) {
	if s.logger != nil {
		switch level {
		case ERROR:
			s.logger.LogError(message)
		case INFO:
			s.logger.LogInfo(message)
		case DEBUG:
			s.logger.LogDebug(message, arg...)
		default:
			panic("unknown log level")
		}
	}
}

func (s *Server) findUser(uname string) (uuid.UUID, error) {
	for id, info := range s.user {
		if info.name == uname {
			return id, nil
		}
	}
	return uuid.UUID{}, ErrInvalidUsernameOrPassword
}

func (s *Server) addUser(uname, passwd string) (respBody []byte, err error) {
	h, err := hasher.HashPassword(passwd)
	if err != nil {
		return nil, ErrInternalServerError
	}

	if _, err := s.findUser(uname); err == nil {
		return nil, ErrUserAlreadyRegistered
	}

	id := uuid.New()
	s.user[id] = &userInfo{name: uname, hash: h, logged: false}

	if respBody, err = marshalValue(CreateUserResponse{uname, id.String()}); err != nil {
		return nil, err
	}

	return respBody, nil
}

func (s *Server) handleCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		obj, err := validateRequest(r)
		if err != nil {
			s.log(ERROR, err.Error())
			responseFailure(w, err)
			return
		}

		resp, err := s.addUser(obj.UserName, obj.Password)

		if err != nil {
			s.log(ERROR, err.Error())
			responseFailure(w, err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(resp))
	}
}

func (s *Server) loginUser(uname, passwd string) (respBody []byte, id uuid.UUID, err error) {
	if id, err = s.findUser(uname); err != nil {
		return nil, uuid.UUID{}, err
	}
	if !hasher.CheckPasswordHash(passwd, s.user[id].hash) {
		return nil, uuid.UUID{}, ErrInvalidUsernameOrPassword
	}

	if s.user[id].logged {
		return nil, uuid.UUID{}, ErrUserAlreadyLoggedIn
	}

	s.user[id].oneTimeToken = generateSecureToken(64)
	s.user[id].expiresAfter = getXExpiresAfter()

	url := fmt.Sprintf("ws://%s/ws&token=%s", getLocalIP(), s.user[id].oneTimeToken)
	if respBody, err = marshalValue(LoginUserResponse{url}); err != nil {
		return nil, uuid.UUID{}, err
	}

	s.user[id].logged = true
	return respBody, id, nil
}

func (s *Server) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		obj, err := validateRequest(r)
		if err != nil {
			s.log(ERROR, err.Error())
			responseFailure(w, err)
			return
		}

		resp, id, err := s.loginUser(obj.UserName, obj.Password)

		if err != nil {
			s.log(ERROR, err.Error())
			responseFailure(w, err)
			return
		}

		w.Header().Add("X-Rate-Limit", getXRLimit())
		w.Header().Add("X-Expires-After", s.user[id].expiresAfter.String())
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(resp))
	}
}

func (s *Server) validateUser(token string) (uuid.UUID, error) {
	for n, v := range s.user {
		if v.oneTimeToken != "" && v.oneTimeToken == token && v.expiresAfter.After(time.Now()) {
			return n, nil
		}
	}
	return uuid.UUID{}, ErrInvalidOneTimeToken
}

func (s *Server) sendTail(c *participant) {
	for _, m := range s.history.tail() {
		message, err := marshalValue(m)

		if err != nil {
			s.log(ERROR, err.Error())
		}

		c.mes <- message
	}
}

func (s *Server) handleConnect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := validateUpgrade(r)
		if err != nil {
			s.log(ERROR, err.Error())
			responseFailure(w, err)
			return
		}

		var id uuid.UUID
		u := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				id, err = s.validateUser(token)
				if err != nil {
					s.log(ERROR, err.Error())
					return false
				}
				s.user[id].oneTimeToken = ""
				return true
			},
		}

		var conn *websocket.Conn
		conn, err = u.Upgrade(w, r, nil)

		if err != nil {
			s.log(ERROR, err.Error())
			responseFailure(w, err)
			return
		}

		newParticipant := &participant{
			name: s.user[id].name,
			srv:  s,
			conn: conn,
			mes:  make(chan []byte, 256),
		}

		s.controller.addParticipant <- newParticipant
		s.sendTail(newParticipant)

		go newParticipant.readMessages()
		go newParticipant.writeMessages()
	}
}

func (s *Server) handleActiveUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activeUsers := ActiveUsersResponce{ActiveUsers: make([]string, len(s.controller.chatroom))}
		i := 0
		for c := range s.controller.chatroom {
			activeUsers.ActiveUsers[i] = c.name
			i++
		}

		sort.Slice(activeUsers.ActiveUsers, func(i, j int) bool { return activeUsers.ActiveUsers[i] < activeUsers.ActiveUsers[j] })

		responceBody, err := marshalValue(activeUsers)
		if err != nil {
			s.log(ERROR, err.Error())
			responseFailure(w, err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, string(responceBody))
	}
}

func (s *Server) handlePanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				s.log(ERROR, fmt.Sprintf("Panic call occured: %v", r))
				responseFailure(w, ErrInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func init() {
	defaultServer.logger = logger.Logger(logger.NewLogger())
	defaultServer.url = getLocalIP() + ":8080"
	defaultServer.user = make(map[uuid.UUID]*userInfo)
	defaultServer.router.routes = make(map[routeInfo]http.HandlerFunc)
	defaultServer.routes()
	defaultServer.controller = newDispatcher()
}

func Run() error {
	go defaultServer.controller.run()
	router := mux.NewRouter()
	for route, handler := range defaultServer.router.routes {
		router.Handle(route.route, defaultServer.handlePanic(handler)).Methods(route.method).Headers(*route.headers...)
	}
	defaultServer.log(INFO, "Starting server....")
	return http.ListenAndServe(defaultServer.url, handlers.LoggingHandler(os.Stdout, router))
}
