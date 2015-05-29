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
	tcpPort, httpPort  *int
	keepAliveAge       time.Duration
	keepAliveCheckTime time.Duration
	globalPolicy       *alarm.Policy
	escalations        alarm.AlarmCollection
	policies           []*alarm.Policy
	index              *event.Index
	encodingPool       *event.EncodingPool
}

func NewPipeline(conf *config.AppConfig) *Pipeline {
	p := &Pipeline{
		encodingPool:       event.NewEncodingPool(event.EncoderFactories[*conf.Encoding], event.DecoderFactories[*conf.Encoding], runtime.NumCPU()),
		tcpPort:            conf.TcpPort,
		httpPort:           conf.HttpPort,
		keepAliveAge:       conf.KeepAliveAge,
		keepAliveCheckTime: 30 * time.Second,
		escalations:        *conf.Escalations,
		index:              event.NewIndex(conf.DbPath),
		policies:           conf.Policies,
		globalPolicy:       conf.GlobalPolicy,
	}
	return p
}

func (p *Pipeline) checkExpired() {
	for {
		time.Sleep(p.keepAliveCheckTime)

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

// Listen for, and serve new incomming tcp connections
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

	if p.globalPolicy != nil {
		if !p.globalPolicy.CheckMatch(e) || !p.globalPolicy.CheckNotMatch(e) {
			return event.OK
		}
	}

	p.index.PutEvent(e)
	for _, pol := range p.policies {
		if pol.Matches(e) {
			act := pol.Action(e)

			// if there is an action to be taken
			if act != "" {

				// create a new incident for this event
				in := p.NewIncident(pol.Name, e)

				// dedup the incident
				if p.Dedupe(in) {

					// update the incident in the index
					if in.Status != event.OK {
						p.index.PutIncident(in)
					} else {
						p.index.DeleteIncidentById(in.IndexName())
					}

					// fetch the escalation to take
					esc, ok := p.escalations[act]
					if ok {

						// send to every alarm in the escalation
						for _, a := range esc {
							a.Send(in)
						}
					} else {
						log.Println("unknown escalation", act)
					}
				}
			}
		}
	}

	return e.Status
}

// returns true if this is a new incident, false if it is a duplicate
func (p *Pipeline) Dedupe(i *event.Incident) bool {
	old := p.index.GetIncident(i.IndexName())

	if old == nil {
		return i.Status != event.OK
	}

	return old.Status != i.Status
}

func (p *Pipeline) ListIncidents() []*event.Incident {
	return p.index.ListIncidents()
}

func (p *Pipeline) PutIncident(in *event.Incident) {
	if in.Id == 0 {
		in.Id = p.index.GetIncidentCounter()
		p.index.UpdateIncidentCounter(in.Id + 1)
	}
	p.index.PutIncident(in)
}

func (p *Pipeline) NewIncident(policy string, e *event.Event) *event.Incident {
	return event.NewIncident(policy, e)
}
