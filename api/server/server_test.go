package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"git.epam.com/vadym_ulitin/lets-go-chat/pkg/hasher"
	"git.epam.com/vadym_ulitin/lets-go-chat/pkg/logger"
	"git.epam.com/vadym_ulitin/lets-go-chat/pkg/logger/mock_logger"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func checkStatusCode(t *testing.T, rw *httptest.ResponseRecorder, expected int, testCase string) {
	if status := rw.Code; status != expected {
		t.Errorf("%s handler returned wrong status code: got %v want %v",
			testCase, status, expected)
	}
}

func checkBody(t *testing.T, rw *httptest.ResponseRecorder, expectedBody string, testCase string) {
	if actualBody := rw.Body.String(); actualBody != expectedBody {
		t.Errorf("%shandler returned unexpected body: got %s want %s",
			testCase, actualBody, expectedBody)
	}
}

func TestHandleCreate(t *testing.T) {
	var testCase string

	var s Server
	s.user = make(map[uuid.UUID]*userInfo)

	handler := http.HandlerFunc(s.handleCreate(context.Background()))

	// valid request
	testCase = "handleCreate, valid request"
	req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte(`{"userName":"john","password":"12345678"}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	checkStatusCode(t, rw, http.StatusOK, testCase)

	var id uuid.UUID
	for id = range s.user {
	}

	checkBody(t, rw, `{"userName":"john","id":"`+id.String()+`"}`, testCase)

	expectedHeader := "application/json"

	if actualHeader := rw.Header().Get("Content-Type"); actualHeader != expectedHeader {
		t.Errorf("handler returned unexpected header: got %s want %s",
			actualHeader, expectedHeader)
	}

	// user already registered
	testCase = "handleCreate, user already registered"
	req, err = http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte(`{"userName":"john","password":"12345678"}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rw = httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	checkStatusCode(t, rw, http.StatusBadRequest, testCase)
	checkBody(t, rw, ErrUserAlreadyRegistered.Error(), testCase)

	// too short name or password
	testCase = "handleCreate, too short name or password"

	req, err = http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte(`{"userName":"jim","password":"12345678"}`)))
	if err != nil {
		t.Fatal(err)
	}

	rw = httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	checkStatusCode(t, rw, http.StatusBadRequest, testCase)
	checkBody(t, rw, ErrEmptyUsernameOrId.Error(), testCase)
}

func TestHandleLogin(t *testing.T) {
	var testCase string

	var s Server
	s.user = make(map[uuid.UUID]*userInfo)
	h, _ := hasher.HashPassword("12345678")
	id := uuid.New()
	s.user[id] = &userInfo{name: "john", hash: h, logged: false}

	handler := http.HandlerFunc(s.handleLogin(context.Background()))

	// valid request
	testCase = "handleLogin, valid request"
	req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte(`{"userName":"john","password":"12345678"}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	checkStatusCode(t, rw, http.StatusOK, testCase)
	checkBody(t, rw, `{"url":"ws://`+getLocalIP()+`/ws\u0026token=`+s.user[id].oneTimeToken+`"}`, testCase)

	expectedHeader := []string{"Content-Type", "X-Rate-Limit", "X-Expires-After"}

	for _, header := range expectedHeader {
		if rw.Header().Get(header) == "" {
			t.Errorf("handler doesn't contain expected header: %s",
				header)
		}
	}

	// too short name or password
	testCase = "handleCreate, too short name or password"

	req, err = http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte(`{"userName":"jim","password":"12345678"}`)))
	if err != nil {
		t.Fatal(err)
	}

	rw = httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	checkStatusCode(t, rw, http.StatusBadRequest, testCase)
	checkBody(t, rw, ErrEmptyUsernameOrId.Error(), testCase)

	// password mismatch
	testCase = "handleLogin, password mismatch"
	req, err = http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte(`{"userName":"john","password":"12345679"}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rw = httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	checkStatusCode(t, rw, http.StatusBadRequest, testCase)
	checkBody(t, rw, ErrInvalidUsernameOrPassword.Error(), testCase)

	// unknown user
	testCase = "handleLogin, unknown user"
	req, err = http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte(`{"userName":"kate","password":"12345679"}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rw = httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	checkStatusCode(t, rw, http.StatusBadRequest, testCase)
	checkBody(t, rw, ErrInvalidUsernameOrPassword.Error(), testCase)

	// user already loggedin
	testCase = "handleLogin, user already loggedin"
	req, err = http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte(`{"userName":"john","password":"12345678"}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rw = httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	checkStatusCode(t, rw, http.StatusBadRequest, testCase)
	checkBody(t, rw, ErrUserAlreadyLoggedIn.Error(), testCase)
}

func TestHandleConnect(t *testing.T) {
	var s Server
	s.user = make(map[uuid.UUID]*userInfo)
	h, _ := hasher.HashPassword("12345678")
	id := uuid.New()
	s.user[id] = &userInfo{name: "john", hash: h, logged: false, oneTimeToken: "1", expiresAfter: getXExpiresAfter()}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = context.WithValue(context.WithValue(ctx, serverKey, &s), controllerKey, newDispatcher())

	handler := http.HandlerFunc(s.handleConnect(ctx))

	// token not provided
	req, err := http.NewRequest(http.MethodGet, "http://169.254.16.78:8080/v1/ws", nil)
	if err != nil {
		t.Fatal(err)
	}

	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)
	checkStatusCode(t, rw, http.StatusBadRequest, "handleConnect, invalid token")
	checkBody(t, rw, ErrInvalidOneTimeToken.Error(), "handleConnect, invalid token")

	// token provided
	req, err = http.NewRequest(http.MethodGet, "http://169.254.16.78:8080/v1/ws?token=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rw = httptest.NewRecorder()

	handler.ServeHTTP(rw, req)
	checkStatusCode(t, rw, http.StatusBadRequest, "handleConnect, invalid token")
}

func TestHandleActiveUsers(t *testing.T) {
	var s Server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = context.WithValue(context.WithValue(ctx, serverKey, &s), controllerKey, newDispatcher())

	ctx.Value(controllerKey).(*dispatcher).chatroom = make(map[*participant]bool)
	ctx.Value(controllerKey).(*dispatcher).chatroom[&participant{uuid: uuid.New()}] = true
	ctx.Value(controllerKey).(*dispatcher).chatroom[&participant{uuid: uuid.New()}] = true

	handler := http.HandlerFunc(s.handleActiveUsers(ctx))

	testCase := "handleActiveUsers"
	req, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatal(err)
	}

	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	checkStatusCode(t, rw, http.StatusOK, testCase)
	checkBody(t, rw, `{"activeusers":["john","kate"]}`, testCase)
}

func TestHandlePanic(t *testing.T) {
	var s Server
	handler := http.Handler(s.handlePanic(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("panic")
	})))
	rw := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(rw, req)
	checkStatusCode(t, rw, http.StatusBadRequest, "handlePanic")
	checkBody(t, rw, ErrInternalServerError.Error(), "handlePanic")
}

func TestValidateUser(t *testing.T) {
	token := generateSecureToken(64)

	var s Server
	s.user = make(map[uuid.UUID]*userInfo)
	id := uuid.New()
	s.user[id] = &userInfo{oneTimeToken: token, expiresAfter: getXExpiresAfter()}
	s.user[uuid.New()] = &userInfo{oneTimeToken: token, expiresAfter: getXExpiresAfter()}

	// valid token
	if actual, err := s.validateUser(token); err != nil {
		t.Errorf("expected user ID %s, got %s, error%s", id, actual, err)
	}

	// invalid token
	if _, err := s.validateUser(generateSecureToken(64)); err != ErrInvalidOneTimeToken {
		t.Errorf("expected error%s, got %s", ErrInvalidOneTimeToken, err)
	}

	// expired token
	s.user[id] = &userInfo{oneTimeToken: token, expiresAfter: time.Now()}
	time.Sleep(1 * time.Millisecond)
	if _, err := s.validateUser(generateSecureToken(64)); err != ErrInvalidOneTimeToken {
		t.Errorf("expected error%s, got %s", ErrInvalidOneTimeToken, err)
	}
}

func TestLog(t *testing.T) {
	var s Server
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockLogger := mock_logger.NewMockLogger(mockCtrl)
	mockLogger.EXPECT().LogError(ErrEmptyUsernameOrId.Error())
	mockLogger.EXPECT().LogInfo("Info")
	mockLogger.EXPECT().LogDebug("%s", "string")
	s.logger = mockLogger
	s.log(ERROR, ErrEmptyUsernameOrId.Error())
	s.log(INFO, "Info")
	s.log(DEBUG, "%s", "string")
}

func TestSendTail(t *testing.T) {
	var s Server
	s.history = Backlogger(&backlog{})
	s.history.Update(Message{Name: "john"})
	s.history.Update(Message{Name: "kate"})
	user := participant{mes: make(chan []byte, 256)}

	s.sendTail(&user)
	history, _ := s.history.GetHistory()
	for _, m := range history {
		message, _ := marshalValue(m)

		if actual, ok := <-user.mes; !ok || string(message) != string(actual) {
			t.Errorf("expected %s, got %s", string(message), string(actual))
		}
	}
}

func BenchmarkHandleActiveUsers(b *testing.B) {
	var s Server
	controller := newDispatcher()
	controller.chatroom = make(map[*participant]bool)
	for i := 0; i < 100; i++ {
		controller.chatroom[&participant{uuid: uuid.New()}] = true
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = context.WithValue(context.WithValue(ctx, serverKey, &s), controllerKey, controller)

	handler := http.HandlerFunc(s.handleActiveUsers(ctx))

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodGet, "", nil)
		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)
	}

}

func ExampleRun() {
	// setup server
	s := Server{
		logger:  logger.Logger(logger.NewLogger()), // 	"git.epam.com/vadym_ulitin/lets-go-chat/pkg/logger"
		url:     ":8080",
		user:    make(map[uuid.UUID]*userInfo),
		router:  Router{routes: make(map[routeInfo]http.HandlerFunc)},
		history: Backlogger(&backlog{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(context.WithValue(ctx, serverKey, &defaultServer), controllerKey, newDispatcher())
	exit := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	// define routes
	s.router.HandleFunc("/v1/user", http.MethodPost, []string{"Content-Type", "application/json"}, s.handleCreate(ctx))
	s.router.HandleFunc("/v1/user/login", http.MethodPost, []string{"Content-Type", "application/json"}, s.handleLogin(ctx))
	s.router.HandleFunc("/v1/ws", http.MethodGet, nil, s.handleConnect(ctx))
	s.router.HandleFunc("/v1/users", http.MethodGet, nil, s.handleActiveUsers(ctx))

	defer defaultServer.history.Close()

	go func() {
		<-exit
		cancel()
	}()

	go ctx.Value(controllerKey).(*dispatcher).run(ctx)

	router := mux.NewRouter() // "github.com/gorilla/mux"

	for route, handler := range defaultServer.router.routes {
		router.Handle(route.route, defaultServer.handlePanic(handler)).Methods(route.method).Headers(*route.headers...)
	}
	s.log(INFO, "Starting server....")

	// http.ListenAndServe(defaultServer.url, router)
	// Output: Starting server....
}
