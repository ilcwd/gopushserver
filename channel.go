package gopushserver

import (
	"container/list"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	DEBUG = false
)

type connection struct {
	ID       int64
	Receiver chan []byte
	Channel  string
}

type channel struct {
	Connections *list.List
	Queue       *Queue
	sync.Mutex
}

var (
	// TODO: all channel shares the same register map,
	// lock is too expensive.
	channelRegister map[string]channel
	registerMutex   sync.Mutex
)

func init() {
	channelRegister = make(map[string]channel)
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

func deleteChannel(channelName string) *channel {
	registerMutex.Lock()
	defer registerMutex.Unlock()
	channel, ok := channelRegister[channelName]
	if ok {
		return &channel
	}
	return nil
}

func PushMessage(channelName string, data []byte) {
	c := getOrCreateChannel(channelName)

	c.Lock()
	defer c.Unlock()

	if c.Connections.Len() == 0 {
		// if no connection is ready, enqueue message.
		c.Queue.Enqueue(data)
		if DEBUG {
			log.Printf("enqueue `%s`", string(data))
		}
	} else {
		// otherwise, send it to all connections.
		timer := time.NewTimer(time.Second)

		for e := c.Connections.Front(); e != nil; e = e.Next() {
			if connE, ok := e.Value.(*connection); ok {
				// try to send,
				select {
				case connE.Receiver <- data:
					if DEBUG {
						log.Printf("send `%s` to %s", string(data), connE.Receiver)
					}
				case <-timer.C:
					if DEBUG {
						log.Printf("send `%s` to %s timeout", string(data), connE.Receiver)
					}
					return
				}
			} else {
				log.Println("[channel][ERROR]Never reach here.")
			}
		}
	}
}

func AddConnection(channelName string) *connection {
	c := getOrCreateChannel(channelName)

	recvChan := make(chan []byte, 1)
	idRander := rand.New(rand.NewSource(time.Now().UnixNano()))
	conn := connection{idRander.Int63(), recvChan, channelName}

	c.Lock()
	defer c.Unlock()
	c.Connections.PushBack(&conn)

	// Try to find buffered messages.
	// Q: Why dequeue only one message?
	// A: Once connection receive a message,
	//    it disconnect immediately,
	//    some we just wait for next connection.
	message := c.Queue.Dequeue()

	if DEBUG {
		log.Printf("dequeue `%s` %s", string(message), conn.Receiver)
	}

	if message != nil {
		// the connection is not ready to
		// receive the message now,
		// so that we use a goroutine.
		go func(_conn *connection) {
			_conn.Receiver <- message
		}(&conn)
	}

	return &conn
}

func DelConnection(conn *connection) bool {
	channel := deleteChannel(conn.Channel)
	if channel != nil {
		return false
	}

	channel.Lock()
	defer channel.Unlock()
	for e := channel.Connections.Front(); e != nil; e = e.Next() {
		if connE, ok := e.Value.(*connection); ok {
			close(connE.Receiver)
			if connE.ID == conn.ID {
				channel.Connections.Remove(e)
				break
			}
		}
	}
	return true
}
