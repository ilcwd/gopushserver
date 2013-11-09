package main

import (
	"gopushserver"
	"time"
)

func main() {
	q := gopushserver.NewQueue(10, time.Second)

	q.Enqueue([]byte("data1"))
	q.Enqueue([]byte("data2"))
	q.Enqueue([]byte("data3"))
	q.Enqueue([]byte("data4"))
	q.Enqueue([]byte("data5"))
	q.Enqueue([]byte("data6"))
}
