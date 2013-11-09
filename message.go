package gopushserver

import (
	"log"
	"sync"
	"time"
)

type Message struct {
	Body    []byte
	Expires time.Time
}

type Queue struct {
	Messages []Message
	Current  int
	Length   int
	Expires  time.Duration
	sync.Mutex
}

func NewQueue(size int, expires time.Duration, recyleTime time.Duration) *Queue {
	q := Queue{make([]Message, size), 0, 0, expires, sync.Mutex{}}

	go func(q *Queue) {
		for {
			time.Sleep(recyleTime)
			q.recyleMessages()
		}
	}(&q)

	return &q
}

func (q *Queue) Enqueue(data []byte) bool {
	q.Lock()
	defer q.Unlock()

	maxSize := len(q.Messages)

	if q.Length >= maxSize {
		return false
	}
	newMessage := &(q.Messages[((q.Current + q.Length) % maxSize)])
	newMessage.Body = data
	newMessage.Expires = time.Now().Add(q.Expires)
	q.Length++
	return true
}

func (q *Queue) Dequeue() []byte {
	q.Lock()
	defer q.Unlock()
	var message *Message
	maxSize := len(q.Messages)
	now := time.Now()
	for {
		if q.Length == 0 {
			break
		}
		message = &(q.Messages[q.Current])
		q.Current = (q.Current + 1) % maxSize
		q.Length--
		body := message.Body
		message.Body = nil
		if message.Expires.After(now) {
			return body
		}
	}
	return nil

}

func (q *Queue) recyleMessages() {
	q.Lock()
	defer q.Unlock()
	if q.Length == 0 {
		return
	}
	MaxSize := len(q.Messages)
	lastMessage := q.Messages[(q.Current+q.Length-1)%MaxSize]
	now := time.Now()
	if lastMessage.Expires.Before(now) {
		for _, m := range q.Messages {
			m.Body = nil
		}
		log.Printf("[Queue]recycled %d messages", q.Length)
	}
	q.Current = 0
	q.Length = 0
}
