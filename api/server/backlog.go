package server

import (
	"math"
	"sync"
	"time"
)

const (
	backLogMaxLen = 1000000
	backLogTail   = 10
)

type Backlogger interface {
	GetHistory() ([]Message, error)
	Update(Message) error
	Close()
}
type Message struct {
	TimeStamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
	Message   string    `json:"message"`
}

type backlog struct {
	mu   sync.Mutex
	data []Message
}

func (b *backlog) GetHistory() ([]Message, error) {
	return b.tail(), nil
}

func (b *backlog) Update(m Message) error {
	b.pushBack(m)
	return nil
}
func (b *backlog) Close() {
}

func (b *backlog) pushBack(m Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.data) >= backLogMaxLen {
		b.data = b.data[1:]
	}

	b.data = append(b.data, m)
}

func (b *backlog) tail() []Message {
	b.mu.Lock()
	defer b.mu.Unlock()

	tailBegin := int(math.Max(0, float64(len(b.data)-backLogTail)))

	return b.data[tailBegin:]
}
