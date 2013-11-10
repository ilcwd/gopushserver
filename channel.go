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
	sync.Mutex
}

var (
	channelRegister map[string]channel
	registerMutex   sync.Mutex
	idRander        *rand.Rand
)

func init() {
	channelRegister = make(map[string]channel)
	idRander = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func getOrCreateChannel(channelName string) *channel {
	registerMutex.Lock()
	defer registerMutex.Unlock()
	c, ok := channelRegister[channelName]
	if !ok {
		connList := list.New()
		queue := NewQueue(64, time.Minute, time.Minute)
		c = channel{connList, queue, sync.Mutex{}}
		channelRegister[channelName] = c
	}
	return &c
}

func PushMessage(channelName string, data []byte) {
	c := getOrCreateChannel(channelName)

	c.Queue.Enqueue(data)

	go notifyClient(c)
}

func notifyClient(c *channel) {
	c.Lock()
	defer c.Unlock()
	if c.Connections.Len() == 0 {
		return
	}
	message := c.Queue.Dequeue()
	if message != nil {
		for e := c.Connections.Front(); e != nil; e = e.Next() {
			if connE, ok := e.Value.(connection); ok {
				// non blocking
				select {
				case connE.Receiver <- message:
				default:
				}

			}
		}
	}
}

func AddConnection(channelName string) *connection {
	c := getOrCreateChannel(channelName)

	recvChan := make(chan []byte)
	conn := connection{idRander.Int(), recvChan, channelName}

	c.Lock()
	defer c.Unlock()
	c.Connections.PushBack(conn)
	// new here
	go notifyClient(c)
	return &conn
}

func DelConnection(conn *connection) bool {
	registerMutex.Lock()
	channel, ok := channelRegister[conn.Channel]
	registerMutex.Unlock()
	if !ok {
		return false
	}

	channel.Lock()
	defer channel.Unlock()
	for e := channel.Connections.Front(); e != nil; e = e.Next() {
		if connE, ok := e.Value.(connection); ok {
			close(connE.Receiver)
			if connE.ID == conn.ID {
				channel.Connections.Remove(e)
				break
			}
		}
	}
	return true
}
