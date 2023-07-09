package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/uilki/lgc/pkg/hasher"
	"github.com/uilki/lgc/pkg/logger"
	"google.golang.org/protobuf/proto"
)

const (
	ERROR = iota
	INFO
	DEBUG
)

type Server struct {
	url     string
	user    map[uuid.UUID]*userInfo
	router  Router
	history Backlogger
	logger  logger.Logger
	wg      sync.WaitGroup
}

type key int

const (
	serverKey     key = 0
	controllerKey key = 1
)

var defaultServer Server

func (s *Server) routes(ctx context.Context) {
	s.router.HandleFunc("/v1/user", http.MethodPost, []string{"Content-Type", "application/json"}, s.handleCreate(ctx))
	s.router.HandleFunc("/v1/user/login", http.MethodPost, []string{"Content-Type", "application/json"}, s.handleLogin(ctx))
	s.router.HandleFunc("/v1/ws", http.MethodGet, nil, s.handleConnect(ctx))
	s.router.HandleFunc("/v1/users", http.MethodGet, nil, s.handleActiveUsers(ctx))
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

func (s *Server) handleCreate(_ context.Context) http.HandlerFunc {
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

func (s *Server) handleLogin(_ context.Context) http.HandlerFunc {
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
	messages, err := s.history.GetHistory()
	if err != nil {
		s.log(ERROR, err.Error())
		return
	}

	for i := 0; i < len(messages); i++ {
		message, err := proto.Marshal(&messages[i])

		if err != nil {
			s.log(ERROR, err.Error())
		}

		c.mes <- message
	}
}

func (s *Server) handleConnect(ctx context.Context) http.HandlerFunc {
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
			uuid: id,
			conn: conn,
			mes:  make(chan []byte, 256),
		}

		ctx.Value(controllerKey).(*dispatcher).addParticipant <- newParticipant
		s.sendTail(newParticipant)

		s.wg.Add(1)
		go newParticipant.readMessages(ctx)
		s.wg.Add(1)
		go newParticipant.writeMessages(ctx)
	}
}

func (s *Server) handleActiveUsers(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		controller := ctx.Value(controllerKey).(*dispatcher)
		activeUsers := ActiveUsersResponce{ActiveUsers: make([]string, len(controller.chatroom))}
		i := 0
		for c := range controller.chatroom {
			activeUsers.ActiveUsers[i] = s.user[c.uuid].name
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
}

func NewServer(backlog Backlogger) *Server {
	return &Server{
		url:     getLocalIP() + ":8080",
		user:    make(map[uuid.UUID]*userInfo),
		router:  Router{routes: make(map[routeInfo]http.HandlerFunc)},
		history: backlog,
		logger:  logger.Logger(logger.NewLogger()),
		wg:      sync.WaitGroup{},
	}
}

func Run(s *Server) error {
	// s, err := wired.InitializeServer(pass)
	// if err != nil {
	// 	panic(err)
	// }

	// setup history
	/* 	if pass != "" {
	   		sqlBacklog, err := NewSqlBacklog(pass)
	   		if err != nil {
	   			panic(err)
	   		}
	   		defaultServer.history = Backlogger(sqlBacklog)
	   	} else {
	   		defaultServer.history = Backlogger(&backlog{})
	   	}
	*/defer s.history.Close()

	// setup shutdown
	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	// run controller
	controllerCtx, controllerCancel := context.WithCancel(context.Background())
	controllerCtx = context.WithValue(controllerCtx, serverKey, &s)
	controller := newDispatcher()
	go controller.run(controllerCtx)
	go func() {
		sig := <-exit
		ctx.Value(serverKey).(*Server).log(INFO, fmt.Sprintf("Signal %v caught", sig))
		cancel()
		s.wg.Wait() // wait for connections shutdowm
		controllerCancel()
		os.Exit(0)
	}()

	// setup routes
	ctx = context.WithValue(
		context.WithValue(ctx,
			controllerKey,
			controller,
		),
		serverKey,
		&s,
	)
	s.routes(ctx)
	router := mux.NewRouter()
	for route, handler := range s.router.routes {
		router.Handle(route.route, s.handlePanic(handler)).Methods(route.method).Headers(*route.headers...)
	}

	// run sserver
	s.log(INFO, fmt.Sprintf("Starting server on %s", s.url))
	return http.ListenAndServe(s.url, handlers.LoggingHandler(os.Stdout, router))
}
