package main

import (
	"github.com/nats-io/stan.go"
	"time"
)

// Публикация данных в nats-streaming
func main() {
	sc, _ := stan.Connect("prod", "testPub")
	defer sc.Close()
	// Simple Synchronous Publisher
	time.Sleep(3 * time.Second)
	for _, val := range Datas {
		sc.Publish("example", []byte(val))
		time.Sleep(3 * time.Second)
	}

}
