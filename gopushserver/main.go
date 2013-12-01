package main

import (
	"flag"
	"gopushserver"
	"log"
	"net/http"

	"encoding/json"
	"os"
	"runtime"
	"time"
)

type GoPushServerConfig struct {
	ADDR string
}

var DefaultConfig GoPushServerConfig

func main() {
	go func() {
		f, _ := os.Create("mem.prof")
		var m runtime.MemStats

		for {
			runtime.ReadMemStats(&m)
			data, _ := json.Marshal(m)
			f.Write(data)
			f.WriteString("\n")
			time.Sleep(time.Second)
		}

	}()

	log.SetFlags(log.Llongfile | log.Ltime | log.Ldate)
	flag.StringVar(&DefaultConfig.ADDR, "host", "0.0.0.0:8002", "http listening host.")
	flag.Parse()

	gopushserver.RegisterAPI()
	log.Printf("Listening to %s ...", DefaultConfig.ADDR)
	err := http.ListenAndServe(DefaultConfig.ADDR, nil)
	if err != nil {
		panic(err)
	}
}
