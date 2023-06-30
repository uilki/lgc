package server

import (
	"context"
	"os"
)

type dispatcher struct {
	chatroom         map[*participant]bool
	broadcast        chan []byte
	addParticipant   chan *participant
	removePaticipant chan *participant
}

func newDispatcher() *dispatcher {
	return &dispatcher{
		chatroom:         make(map[*participant]bool),
		broadcast:        make(chan []byte),
		addParticipant:   make(chan *participant),
		removePaticipant: make(chan *participant),
	}
}

func (d *dispatcher) run(ctx context.Context) {
	for {
		select {
		case participant := <-d.addParticipant:
			d.chatroom[participant] = true
		case participant := <-d.removePaticipant:
			delete(d.chatroom, participant)
		case message := <-d.broadcast:
			for participant := range d.chatroom {
				select {
				case participant.mes <- message:
				default:
					close(participant.mes)
					delete(d.chatroom, participant)
				}
			}
		case <-ctx.Done():
			ctx.Value(serverKey).(*Server).wg.Wait()
			os.Exit(0)
		}
	}
}
