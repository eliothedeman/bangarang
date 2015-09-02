package main

import (
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/eliothedeman/bangarang/event"
)

func genEvents(hosts, services []string, next chan *event.Event) {
	gen := func() *event.Event {
		return &event.Event{
			Host:    hosts[rand.Intn(len(hosts)-1)],
			Service: services[rand.Intn(len(services)-1)],
			Metric:  rand.Float64() * 100,
		}
	}

	go func() {
		for {
			next <- gen()
		}
	}()
}

func main() {
	n := make(chan *event.Event)
	genEvents([]string{"host1", "host2", "host3"}, []string{"1", "2", "3", "4"}, n)

	for {
		c, err := net.Dial("tcp4", "localhost:5555")
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second)
		} else {
			for {
				e := <-n
				buff, _ := e.MarshalBinary()
				_, err := c.Write(buff)
				if err != nil {
					log.Println(err)
					break
				}
			}
		}
	}

}
