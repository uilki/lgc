package server

import (
	"fmt"
	"net/http"

	"git.epam.com/vadym_ulitin/lets-go-chat/pkg/hasher"
	"git.epam.com/vadym_ulitin/lets-go-chat/pkg/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type server struct {
	url        string
	user       map[uuid.UUID]*userInfo
	router     Router
	history    backlog
	controller *dispatcher
	logger     *logger.ServerLogger
}

var defaultServer server

func (s *server) routes() {
	s.router.HandleFunc("/v1/user", s.handleCreate())
	s.router.HandleFunc("/v1/user/login", s.handleLogin())
	s.router.HandleFunc("/v1/ws", s.handleConnect())
	s.router.HandleFunc("/v1/users", s.handleActiveUsers())
}

func (s *server) findUser(uname string) (uuid.UUID, error) {
	for id, info := range s.user {
		if info.name == uname {
			return id, nil
		}
	}
	return uuid.UUID{}, ErrInvalidUsernameOrPassword
}

func (s *server) addUser(uname, passwd string) (respBody []byte, err error) {
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

func (s *server) handleCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.LogInfo(r.Method + " " + r.Host + r.RequestURI)
		obj, err := validateRequest(r)
		if err != nil {
			s.logger.LogError(err)
			responseFailure(w, err)
			return
		}

		resp, err := s.addUser(obj.UserName, obj.Password)

		if err != nil {
			s.logger.LogError(err)
			responseFailure(w, err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(resp))
	}
}

func (s *server) loginUser(uname, passwd string) (respBody []byte, err error) {
	var id uuid.UUID
	if id, err = s.findUser(uname); err != nil {
		return nil, err
	}
	if !hasher.CheckPasswordHash(passwd, s.user[id].hash) {
		return nil, ErrInvalidUsernameOrPassword
	}

	if s.user[id].logged {
		return nil, ErrUserAlreadyLoggedIn
	}

	s.user[id].oneTimeToken = generateSecureToken(64)

	url := fmt.Sprintf("ws://fancy-chat.io/ws&token=%s", s.user[id].oneTimeToken)
	if respBody, err = marshalValue(LoginUserResponse{url}); err != nil {
		return nil, err
	}

	s.user[id].logged = true
	return respBody, nil
}

func (s *server) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.LogInfo(r.Method + " " + r.Host + r.RequestURI)
		obj, err := validateRequest(r)
		if err != nil {
			s.logger.LogError(err)
			responseFailure(w, err)
			return
		}

		resp, err := s.loginUser(obj.UserName, obj.Password)

		if err != nil {
			s.logger.LogError(err)
			responseFailure(w, err)
			return
		}

		w.Header().Add("X-Rate-Limit", getXRLimit())
		w.Header().Add("X-Expires-After", getXExpiresAfter().String())
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(resp))
	}
}

func (s *server) validateUser(token string) (uuid.UUID, error) {
	for n, v := range s.user {
		if v.oneTimeToken != "" && v.oneTimeToken == token {
			return n, nil
		}
	}
	return uuid.UUID{}, ErrInvalidOneTimeToken
}

func (s *server) sendTail(c *participant) {
	for _, m := range s.history.tail() {
		message, err := marshalValue(m)

		if err != nil {
			c.logger.LogError(err)
		}

		c.mes <- message
	}
}

func (s *server) handleConnect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.LogInfo(r.Method + " " + r.Host + r.RequestURI)

		token, err := validateUpgrade(r)
		if err != nil {
			s.logger.LogError(err)
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
					s.logger.LogError(err)
					return false
				}
				s.user[id].oneTimeToken = ""
				return true
			},
		}

		var conn *websocket.Conn
		conn, err = u.Upgrade(w, r, nil)

		if err != nil {
			s.logger.LogError(err)
			return
		}

		newParticipant := &participant{
			name:       s.user[id].name,
			controller: s.controller,
			history:    &s.history,
			conn:       conn,
			mes:        make(chan []byte, 256),
			logger:     s.logger,
		}

		s.controller.addParticipant <- newParticipant
		s.sendTail(newParticipant)

		go newParticipant.readMessges()
		go newParticipant.writeMessages()
	}
}

func (s *server) handleActiveUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.LogInfo(r.Method + " " + r.Host + r.RequestURI)
		if r.Method != "GET" {
			s.logger.LogError(ErrUnsupportedMethod)
			responseFailure(w, ErrUnsupportedMethod)
			return
		}

		activeUsers := ActiveUsersResponce{ActiveUsers: make([]string, len(s.controller.chatroom))}
		i := 0
		for c := range s.controller.chatroom {
			activeUsers.ActiveUsers[i] = c.name
			i++
		}

		responceBody, err := marshalValue(activeUsers)
		if err != nil {
			s.logger.LogError(err)
			responseFailure(w, err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, string(responceBody))
	}
}

func init() {
	defaultServer.logger = logger.NewLogger()
	defaultServer.url = ":8080"
	defaultServer.user = make(map[uuid.UUID]*userInfo)
	defaultServer.router.routes = make(map[string]http.HandlerFunc)
	defaultServer.routes()
	defaultServer.controller = newDispatcher()
	go defaultServer.controller.run()
	defaultServer.logger.LogInfo("Initialized")
}

func Run() error {
	for r, h := range defaultServer.router.routes {
		http.HandleFunc(r, h)
	}
	return http.ListenAndServe(defaultServer.url, nil)
}
