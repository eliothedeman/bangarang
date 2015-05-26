package pipeline

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/event"
)

type Pipeline struct {
	tcpPort, httpPort *int
	escalations       []*alarm.Escalation
	keepAliveAge      time.Duration
	globalPolicy      *alarm.Policy
	index             *event.Index
	encodingPool      *event.EncodingPool
}

func NewPipeline(conf *config.AppConfig) *Pipeline {
	p := &Pipeline{
		encodingPool: event.NewEncodingPool(event.EncoderFactories[*conf.Encoding], event.DecoderFactories[*conf.Encoding], runtime.NumCPU()),
		tcpPort:      conf.TcpPort,
		httpPort:     conf.HttpPort,
		keepAliveAge: conf.KeepAliveAge,
		escalations:  conf.Escalations,
		index:        event.NewIndex(conf.DbPath),
		globalPolicy: conf.GlobalPolicy,
	}
	return p
}

func (p *Pipeline) checkExpired() {
	for {
		time.Sleep(30 * time.Second)

		hosts := p.index.GetExpired(p.keepAliveAge)
		for _, host := range hosts {
			e := &event.Event{
				Host:    host,
				Service: "KeepAlive",
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
			return
		}
		var e *event.Event
		p.encodingPool.Decode(func(d event.Decoder) {
			e, err = d.Decode(buff)
		})
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
	var e *event.Event
	var err error
	p.encodingPool.Decode(func(d event.Decoder) {
		e, err = d.Decode(buff)
	})
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

// Run the given event though the pipeline
func (p *Pipeline) Process(e *event.Event) int {
	if p.index == nil {
		p.index = event.NewIndex("event.db")
	}

	p.index.PutEvent(e)
	if p.globalPolicy != nil {
		if !p.globalPolicy.CheckMatch(e) || !p.globalPolicy.CheckNotMatch(e) {
			return event.OK
		}
	}

	for _, esc := range p.escalations {
		if esc.Match(e) {
			esc.StatusOf(e)
			if e.StatusChanged() {
				for _, a := range esc.Alarms {
					err := a.Send(e)
					if err != nil {
						log.Println(err)
					}
				}

				if e.Status != event.OK {
					p.index.PutIncident(p.NewIncident(esc.EscalationPolicy, e))
				} else {
					p.index.DeleteIncidentByEvent(e)
				}
				p.index.UpdateEvent(e)
				return e.Status
			}
		}
	}
	p.index.UpdateEvent(e)
	return e.Status
}

func (p *Pipeline) ListIncidents() []*event.Incident {
	return p.index.ListIncidents()
}

func (p *Pipeline) GetIncident(id int64) *event.Incident {
	return p.index.GetIncident(id)
}

func (p *Pipeline) PutIncident(in *event.Incident) {
	if in.Id == 0 {
		in.Id = p.index.GetIncidentCounter()
		p.index.UpdateIncidentCounter(in.Id + 1)
	}
	p.index.PutIncident(in)
}

func (p *Pipeline) NewIncident(escalation string, e *event.Event) *event.Incident {
	return event.NewIncident(escalation, p.index, e)
}
