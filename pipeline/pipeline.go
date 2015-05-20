package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/event"
	"github.com/pquerna/ffjson/ffjson"
)

type Pipeline struct {
	tcpPort, httpPort *int
	escalations       []*alarm.Escalation
	keepAliveAge      time.Duration
	index             *event.Index
}

func NewPipeline(conf *config.AppConfig) *Pipeline {
	return &Pipeline{
		tcpPort:      conf.TcpPort,
		httpPort:     conf.HttpPort,
		keepAliveAge: conf.KeepAliveAge,
		escalations:  conf.Escalations,
		index:        event.NewIndex(conf.DbPath),
	}
}

func (p *Pipeline) checkExpired() {
	for {
		time.Sleep(30 * time.Second)

		hosts := p.index.GetExpired(p.keepAliveAge)
		for _, host := range hosts {
			e := &event.Event{
				Host:    host,
				Service: "Keepalive",
				Metric:  float64(p.keepAliveAge),
			}
			p.Process(e)
		}
	}
}

func (p *Pipeline) Start() {
	go p.IngestHttp()
	go p.IngestTcp()
	go p.checkExpired()
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
	term := []byte{byte(0)}
	for {
		n, err := c.Read(buff)
		if err != nil {
			log.Println(err)
			c.Close()
			return
		}

		buffs := bytes.Split(buff[:n], term)
		for _, b := range buffs {
			p.consume(b)
		}
	}
}

func (p *Pipeline) consume(buff []byte) {
	if len(buff) < 2 {
		return
	}
	e := &event.Event{}
	err := ffjson.UnmarshalFast(buff, e)
	if err != nil {
		log.Println(err)
	} else {
		p.Process(e)
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
		p.index = event.NewIndex("test")
	}

	p.index.Put(e)
	for _, v := range p.escalations {
		if v.Match(e) {
			v.StatusOf(e)
			if e.StatusChanged() {
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
