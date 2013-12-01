package gopushserver

import (
	"net/http"
)

func RegisterAPI() {
	http.HandleFunc("/sync/get", SyncGet)
	http.HandleFunc("/sync/push", SyncPush)
}
