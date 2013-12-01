package gopushserver

import (
	"io/ioutil"
	"net/http"
	"time"
)

const (
	KEEPALIVE_TIMEOUT = time.Minute * 6
)

// Block to get messages.
//
// If the client does not close normally,
// we don't know if the client is alive,
// so message will be dequeue and send to the dead client.
// Client SHOULD wait until it receives a `204 No Content`,
// normally it is 6 minutes' timeout.
func SyncGet(w http.ResponseWriter, req *http.Request) {
	queryString := req.URL.Query()
	uid := queryString.Get("uid")
	conn := AddConnection(uid)
	defer DelConnection(conn)

	timer := time.NewTimer(KEEPALIVE_TIMEOUT)
	select {
	case messageData := <-conn.Receiver:
		w.WriteHeader(http.StatusOK)
		w.Write(messageData)
	case <-timer.C:
		// timeout to release old / dead connections
		w.WriteHeader(http.StatusNoContent)
		break
	}

}

// Send a message to the specified channel.
func SyncPush(w http.ResponseWriter, req *http.Request) {
	queryString := req.URL.Query()
	uid := queryString.Get("uid")

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	go PushMessage(uid, data)

	if DEBUG {
		// echo
		w.Write(data)
	}
}
