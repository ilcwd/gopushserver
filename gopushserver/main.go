package main

import (
	"flag"
	"gopushserver"
	"log"
	"net/http"
)

type GoPushServerConfig struct {
	ADDR string
}

var DefaultConfig GoPushServerConfig

func main() {
	flag.StringVar(&DefaultConfig.ADDR, "host", "0.0.0.0:8002", "http listening host.")
	log.SetFlags(log.Llongfile | log.Ltime | log.Ldate)

	gopushserver.RegisterAPI()
	log.Printf("Listening to %s ...", DefaultConfig.ADDR)
	http.ListenAndServe(DefaultConfig.ADDR, nil)
}
