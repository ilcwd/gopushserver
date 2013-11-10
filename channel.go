package gopushserver

import (
	"container/list"
	"math/rand"
	"sync"
	"time"
)

const (
	OP_NEW_CONNECTION = iota
)

type connection struct {
	ID       int
	Receiver chan []byte
	Channel  string
}

type channel struct {
	Connections *list.List
	Queue       *Queue
	stop        chan int
}

var (
	defaultChannels map[string]channel
	mu              sync.Mutex
	r               *rand.Rand
)

func init() {
	defaultChannels = make(map[string]channel)
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func getOrCreateChannel(channelName string) *channel {
	mu.Lock()
	defer mu.Unlock()
	c, ok := defaultChannels[channelName]
	if !ok {
		connList := list.New()
		queue := NewQueue(64, time.Minute, time.Minute)
		c = channel{connList, queue, make(chan int)}
		defaultChannels[channelName] = c
	}
	return &c
}

func PushMessage(channelName string, data []byte) {
	c := getOrCreateChannel(channelName)

	c.Queue.Enqueue(data)

	go notifyClient(c)
}

func notifyClient(c *channel) {
	message := c.Queue.Dequeue()
	if message != nil {
		for e := c.Connections.Front(); e != nil; e = e.Next() {
			if connE, ok := e.Value.(connection); ok {
				connE.Receiver <- message
			}
		}
	}
}

func AddConnection(channelName string) *connection {
	c := getOrCreateChannel(channelName)

	recvChan := make(chan []byte)
	conn := connection{r.Int(), recvChan, channelName}
	c.Connections.PushBack(conn)
	// new here
	go notifyClient(c)
	return &conn
}

func DelConnection(conn *connection) bool {
	channel, ok := defaultChannels[conn.Channel]
	if !ok {
		return false
	}
	for e := channel.Connections.Front(); e != nil; e = e.Next() {
		if connE, ok := e.Value.(connection); ok {
			if connE.ID == conn.ID {
				channel.Connections.Remove(e)
				break
			}
		}
	}
	return true
}

func GetMessage(channel string) chan *Message {
	c := make(chan *Message)

	go func() {
		c <- &Message{Body: []byte("hello")}
	}()
	return c
}
