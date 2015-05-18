package pipeline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/event"
)

type Pipeline struct {
	tcpPort, httpPort *int
	escalations       []*alarm.Escalation
	index             *event.Index
}

func NewPipeline(tcpPort, httpPort *int) *Pipeline {
	return &Pipeline{
		tcpPort:  tcpPort,
		httpPort: httpPort,
		index:    event.NewIndex(),
	}
}

func (p *Pipeline) Start() {
	go p.IngestHttp()
	go p.IngestTcp()
}

func (p *Pipeline) IngestHttp() {
	if p.httpPort == nil {
		return
	}
	http.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		e := &event.Event{}

		err = json.Unmarshal(buff, e)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		p.Process(e)
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *p.httpPort), nil))
}

func (p *Pipeline) consumeTcp(c *net.TCPConn) {
	buff := make([]byte, 1024*200)
	for {
		n, err := c.Read(buff)
		log.Println(string(buff))
		if err != nil {
			log.Println(err)
			c.Close()
			return
		}
		e := &event.Event{}
		err = json.Unmarshal(buff[:n], e)
		if err != nil {
			log.Println(err)
		} else {
			p.Process(e)
		}

	}
}

func (p *Pipeline) IngestTcp() {
	if p.tcpPort == nil {
		return
	}
	addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", *p.tcpPort))
	if err != nil {
		log.Fatal(err)
	}

	c, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := c.AcceptTCP()
		if err != nil {
			log.Println(err)
		} else {
			go p.consumeTcp(conn)
		}
	}
}

func (p *Pipeline) Process(e *event.Event) int {
	if p.index == nil {
		p.index = event.NewIndex()
	}

	p.index.Put(e)
	for _, v := range p.escalations {
		if v.Match(e) {
			if v.StatusOf(e) != event.OK && e.StatusChanged() {
				for _, a := range v.Alarms {
					err := a.Send(e)
					if err != nil {
						log.Println(err)
					}
				}
				return e.Status
			}
		}
	}
	return e.Status
}
