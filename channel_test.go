package gopushserver

import (
	"bytes"
	"fmt"
	"log"
	"testing"
	"time"
)

func testCanReceive(conn *connection, expectData []byte) chan bool {
	ok := make(chan bool, 1)

	go func() {
		var returnData []byte
		select {
		case returnData = <-conn.Receiver:
			DelConnection(conn)
			if bytes.Compare(expectData, returnData) != 0 {
				log.Printf("Data not matched, expect `%s`, but`%s`.", string(expectData), string(returnData))
				ok <- false
				break
			}
			ok <- true
		case <-time.After(time.Second):
			log.Printf("timeout while try to receive `%s`, %s.", string(expectData))
			ok <- false
			break
		}
	}()
	return ok
}

func TestChannelBasic(t *testing.T) {
	channelname := "chan1Basic"
	data := []byte("TestChannelBasic")

	conn := AddConnection(channelname)
	ok := testCanReceive(conn, data)

	PushMessage(channelname, data)

	if !<-ok {
		t.Fatalf("timeout while try to receive `%s`.", string(data))
	}

}

func TestChannelBufferMessages(t *testing.T) {
	channelname := "chanBufferMessages"
	testcases := [][]byte{
		[]byte("TestChannelBufferMessages__1"),
		[]byte("TestChannelBufferMessages__2"),
		[]byte("TestChannelBufferMessages__3"),
	}

	for _, data := range testcases {
		PushMessage(channelname, data)
	}

	for _, data := range testcases {
		conn := AddConnection(channelname)
		ok1 := testCanReceive(conn, data)
		if !<-ok1 {
			t.Fatalf("Failed")
		}
	}
}

func TestChannelMultiConnections(t *testing.T) {
	channelname := "chanMultiConnections"
	data1 := []byte("TestChannelMultiConnections___1")
	data2 := []byte("TestChannelMultiConnections___2")
	connNum := 3

	var connections []*connection
	for i := 0; i < connNum; i++ {
		_conn := AddConnection(channelname)
		connections = append(connections, _conn)
	}

	PushMessage(channelname, data1)

	for _, conn := range connections {
		ok := testCanReceive(conn, data1)
		if !<-ok {
			t.Fatalf("Expect receive `%s` errors.", string(data1))
		}
	}

	// wait till all goroutines are ready.
	time.Sleep(time.Millisecond * 50)

	// more connections
	connections = nil
	for i := 0; i < connNum+connNum/2; i++ {
		_conn := AddConnection(channelname)
		connections = append(connections, _conn)
	}

	PushMessage(channelname, data2)

	// receives the 2nd message.
	for _, conn := range connections {
		ok := testCanReceive(conn, data2)
		if !<-ok {
			t.Fatalf("Expect receive `%s` errors.", string(data2))
		}
	}

}

func BenchmarkChannel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		channelName := fmt.Sprintf("%d", i)
		AddConnection(channelName)
	}
}
