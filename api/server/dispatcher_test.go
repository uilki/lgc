package server

import (
	"context"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	d := dispatcher{
		chatroom:         make(map[*participant]bool),
		broadcast:        make(chan []byte),
		addParticipant:   make(chan *participant),
		removePaticipant: make(chan *participant),
	}

	ctx, _ := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, serverKey, &Server{})

	go d.run(ctx)
	firstParticipant := participant{mes: make(chan []byte, 256), name: "john"}
	secondParticipant := participant{mes: make(chan []byte, 256), name: "kate"}
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
