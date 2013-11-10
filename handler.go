package gopushserver

import (
	"io/ioutil"
	"net/http"
	"time"
)

func SyncGet(w http.ResponseWriter, req *http.Request) {
	queryString := req.URL.Query()
	uid := queryString.Get("uid")

	conn := AddConnection(uid)

	timer := time.NewTimer(time.Minute * 6)
	select {
	case messageData := <-conn.Receiver:
		w.WriteHeader(http.StatusOK)
		w.Write(messageData)
	case <-timer.C:
		// timeout
		DelConnection(conn)
		w.WriteHeader(http.StatusOK)
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
