package main

import (
	"gopushserver"
	"log"
	"net/http"
)

func main() {

	log.SetFlags(log.Llongfile)

	http.HandleFunc("/sync/get", gopushserver.SyncGet)
	http.HandleFunc("/sync/push", gopushserver.SyncPush)
	http.ListenAndServe("0.0.0.0:8002", nil)
}
