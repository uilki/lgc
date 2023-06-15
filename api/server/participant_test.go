package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

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
		srv:  &Server{controller: newDispatcher()},
		conn: c,
		mes:  make(chan []byte, 256),
	}
	go p.readMessages()
	go p.writeMessages()

	// Write a message
	testMessage := "data"
	p.mes <- []byte(testMessage)

	// when readMessages receives message it puts it into broadcast channel
	_, ok := <-p.srv.controller.broadcast
	if !ok {
		t.Fatal("Can't read broadcast chan")
	}

	// and into backlog
	if m := p.srv.history.data[len(p.srv.history.data)-1].Message; m != testMessage {
		t.Errorf(`expected "%s" got "%s"`, testMessage, m)
	}
}
