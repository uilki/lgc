package server

import (
	"fmt"
	"strconv"
	"testing"
)

func TestPushBack(t *testing.T) {
	bl := backlog{}
	for i := 0; i < backLogMaxLen+1; i++ {
		bl.pushBack(Message{Name: fmt.Sprintf("%d", i)})
	}

	if len(bl.data) > backLogMaxLen {
		t.Errorf("Backlog len should be less than %d, actual %d", backLogMaxLen, len(bl.data))
	}
}

func generate[T any](s *[]T, count int, g func(n int) T) {
	for i := 0; i < count; i++ {
		*s = append(*s, g(i))
	}
}

func TestTail(t *testing.T) {
	bl := backlog{}
	for i := 0; i < backLogMaxLen; i++ {
		bl.pushBack(Message{Name: fmt.Sprintf("%d", i)})
	}
	var expected []string
	generate(&expected, backLogTail, func(n int) string { return strconv.Itoa(backLogMaxLen - backLogTail + n) })

	tail := bl.tail()
	if len(tail) != backLogTail {
		t.Errorf("tail length expected %d, actual %d", backLogTail, len(tail))
	}
	for i, n := range expected {
		if tail[i].Name != n {
			t.Errorf("expected %s equals %s", n, tail[i].Name)
		}
	}
}

func benchmarkPushBack(bl *backlog, b *testing.B) {
	for i := 0; i < b.N; i++ {
		bl.pushBack(Message{Name: fmt.Sprintf("%d", i)})
	}
}

func BenchmarkPushBack(b *testing.B) {
	bl := backlog{}
	benchmarkPushBack(&bl, b)
}

func BenchmarkPushBackFull(b *testing.B) {
	bl := backlog{data: make([]Message, backLogMaxLen)}
	benchmarkPushBack(&bl, b)
}
