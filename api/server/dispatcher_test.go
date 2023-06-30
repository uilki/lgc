package server

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRun(t *testing.T) {
	d := dispatcher{
		chatroom:         make(map[*participant]bool),
		broadcast:        make(chan []byte),
		addParticipant:   make(chan *participant),
		removePaticipant: make(chan *participant),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = context.WithValue(ctx, serverKey, &Server{})

	go d.run(ctx)
	firstParticipant := participant{mes: make(chan []byte, 256), uuid: uuid.New()}
	secondParticipant := participant{mes: make(chan []byte, 256), uuid: uuid.New()}
	d.addParticipant <- &firstParticipant
	d.addParticipant <- &secondParticipant

	testMes := []byte("Hello")
	d.broadcast <- testMes
	time.Sleep(time.Millisecond)

	if chatRoomLen := len(d.chatroom); chatRoomLen != 2 {
		t.Errorf("chatroom expected len %d, actual %d", 2, chatRoomLen)
	}

	if actual, ok := <-firstParticipant.mes; !ok || string(actual) != string(testMes) {
		t.Errorf("expected mesage %s, actual %s", string(testMes), string(actual))
	}

	if actual, ok := <-secondParticipant.mes; !ok || string(actual) != string(testMes) {
		t.Errorf("expected mesage %s, actual %s", string(testMes), string(actual))
	}

	for i := 0; i < 256; i++ {
		firstParticipant.mes <- testMes
	}
	d.broadcast <- testMes

	time.Sleep(time.Millisecond)

	if chatRoomLen := len(d.chatroom); chatRoomLen != 1 {
		t.Errorf("chatroom expected len %d, actual %d", 1, chatRoomLen)
	}

	d.removePaticipant <- &secondParticipant
	time.Sleep(time.Millisecond)

	if chatRoomLen := len(d.chatroom); chatRoomLen != 0 {
		t.Errorf("chatroom expected len %d, actual %d", 0, chatRoomLen)
	}

}
