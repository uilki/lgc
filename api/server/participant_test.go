package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
)

/*
	 func TestReadWriteMessages(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var upgrader = websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Fatal("Error during connection upgradation:", err)
				return
			}
			defer conn.Close()

			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					break
				}
				err = conn.WriteMessage(messageType, message)
				if err != nil {
					break
				}
			}
		}))
		defer s.Close()

		wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
		c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Error(err)
		}
		defer c.Close()

		p := participant{
			conn: c,
			mes:  make(chan []byte, 256),
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx = context.WithValue(context.WithValue(ctx, serverKey, &Server{history: Backlogger(&backlog{})}), controllerKey, newDispatcher())
		ctx.Value(serverKey).(*Server).wg.Add(1)
		go p.readMessages(ctx)
		ctx.Value(serverKey).(*Server).wg.Add(1)
		go p.writeMessages(ctx)

		// Write a message
		testMessage := "data"
		p.mes <- []byte(testMessage)

		// when readMessages receives message it puts it into broadcast channel
		_, ok := <-ctx.Value(controllerKey).(*dispatcher).broadcast
		if !ok {
			t.Fatal("Can't read broadcast chan")
		}

		// and into backlog
		messages, _ := ctx.Value(serverKey).(*Server).history.GetHistory()
		if m := messages[len(messages)-1].Message; m != testMessage {
			t.Errorf(`expected "%s" got "%s"`, testMessage, m)
		}
	}
*/
func TestReadMessages(t *testing.T) {
	controller := newDispatcher()
	MockBacklogger := NewMockBacklogger(t)
	MockBacklogger.On("Update", mock.Anything).Return(nil)

	// MockBacklogger.EXPECT().Update(&pb.Message{}).Return(nil).Once()
	mockServer := &Server{user: make(map[uuid.UUID]*userInfo), history: MockBacklogger}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal("Error during connection upgrading:", err)
			return
		}

		p := &participant{uuid: uuid.New(), conn: conn, mes: make(chan []byte, 256)}
		mockServer.user[p.uuid] = &userInfo{name: "name", logged: true}
		controller.chatroom[p] = true
		ctx := context.WithValue(context.WithValue(context.Background(), serverKey, mockServer), controllerKey, controller)
		mockServer.wg.Add(1)
		go p.readMessages(ctx)
	}))
	defer s.Close()

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	w, err := c.NextWriter(websocket.TextMessage)
	if err != nil {
		t.Fatal(err.Error())
	}
	w.Write([]byte("message"))

	if err := w.Close(); err != nil {
		t.Fatal(err.Error())
	}
	received := string(<-controller.broadcast)
	if !strings.Contains(received, "message") {
		t.Errorf(`expected "message" in  "%s"`, received)
	}
	c.WriteMessage(websocket.CloseMessage, []byte{})

	<-controller.removePaticipant

	mockServer.wg.Wait()
}
func TestWriteMessages(t *testing.T) {
	mockServer := &Server{}
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, serverKey, mockServer)

	messageText := make([]string, 10)
	for i := 0; i < 10; i++ {
		messageText[i] = "message" + strconv.Itoa(i)
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal("Error during connection upgrading:", err)
			return
		}
		p := &participant{conn: conn, mes: make(chan []byte, 256)}

		for i := 0; i < 10; i++ {
			p.mes <- []byte(messageText[i])
		}

		mockServer.wg.Add(1)
		go p.writeMessages(ctx)

	}))
	defer s.Close()

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	for i := 0; i < 10; i++ {
		fmt.Println(i)

		if _, message, err := c.ReadMessage(); err == nil {
			if !strings.Contains(string(message), messageText[i]) {
				t.Errorf(`expected "%s" in  "%s"`, messageText[i], message)
			}
		} else {
			t.Errorf(err.Error())
		}
	}
	cancel()
	mockServer.wg.Wait()
}
