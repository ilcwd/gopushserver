package gopushserver

import (
	"io/ioutil"
	"net/http"
	"time"
)

// block to get messages.
// if the client does not close normally,
// we don't know if the client is alive,
// so the message will be dequeue and send to the dead client.
func SyncGet(w http.ResponseWriter, req *http.Request) {
	queryString := req.URL.Query()
	uid := queryString.Get("uid")
	conn := AddConnection(uid)
	defer DelConnection(conn)

	timer := time.NewTimer(time.Minute * 6)
	select {
	case messageData := <-conn.Receiver:
		w.WriteHeader(http.StatusOK)
		w.Write(messageData)
	case <-timer.C:
		// timeout
		w.WriteHeader(http.StatusNoContent)
		break
	}

}

func SyncPush(w http.ResponseWriter, req *http.Request) {
	queryString := req.URL.Query()
	uid := queryString.Get("uid")

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	go PushMessage(uid, data)
}
