package server

import (
	"math"
	"sync"
	"time"

	pb "github.com/uilki/lgc/api/server/generated"
)

const (
	backLogMaxLen = 1000000
	backLogTail   = 10
)

type Message struct {
	TimeStamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
	Message   string    `json:"message"`
}

type backlog struct {
	mu   sync.Mutex
	data []pb.Message
}

func (b *backlog) GetHistory() ([]pb.Message, error) {
	return b.tail(), nil
}

func (b *backlog) Update(m *pb.Message) error {
	b.pushBack(m)
	return nil
}

func (b *backlog) Close() {
}

func (b *backlog) pushBack(m *pb.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.data) >= backLogMaxLen {
		b.data = b.data[1:]
	}

	b.data = append(b.data, pb.Message{Name: m.Name, TimeStamp: m.TimeStamp, Message: m.Message})
}

func (b *backlog) tail() []pb.Message {
	b.mu.Lock()
	defer b.mu.Unlock()

	tailBegin := int(math.Max(0, float64(len(b.data)-backLogTail)))

	return b.data[tailBegin:]
}
