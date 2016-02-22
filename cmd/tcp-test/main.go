package main

import (
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/newman"
)

func genEvents(hosts, services []string, next chan *event.Event) {
	gen := func() *event.Event {
		e := event.NewEvent()
		e.Metric = rand.Float64() * 100
		e.Time = time.Now()
		e.Tags.Set("host", hosts[rand.Intn(len(hosts)-1)])
		e.Tags.Set("service", services[rand.Intn(len(services)-1)])
		return e
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
	slp := time.Second

	for {
		c, err := net.Dial("tcp4", "localhost:5555")
		if err != nil {
			log.Println(err)
			time.Sleep(slp)
			slp *= 2
		} else {
			slp = time.Second

			conn := newman.NewConn(c)
			for {
				e := <-n
				err = conn.Write(e)
				if err != nil {
					log.Println(err)
					break
				}
			}
		}
	}
}
