package tcp

import (
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/bangarang/provider"
	"github.com/eliothedeman/newman"
)

const START_HANDSHAKE = "BANGARANG: TCP_PROVIDER"

func init() {
	provider.LoadEventProviderFactory("tcp", NewTCPProvider)
}

// provides events from tcp connections
type TCPProvider struct {
	encoding string
	laddr    *net.TCPAddr
	listener *net.TCPListener
}

func NewTCPProvider() provider.EventProvider {
	return &TCPProvider{}
}

// the config struct for the tcp provider
type TCPConfig struct {
	Listen string `json:"listen"`
}

func (t *TCPProvider) Init(i interface{}) error {
	c := i.(*TCPConfig)

	// make sure we have a valid address
	addr, err := net.ResolveTCPAddr("tcp4", c.Listen)
	if err != nil {
		return err
	}

	t.laddr = addr

	return nil
}

func (t *TCPProvider) ConfigStruct() interface{} {
	return &TCPConfig{}
}

// start accepting connections and consume each of them as they come in
func (t *TCPProvider) Start(p event.Passer) {

	logrus.Infof("TCP Provider listening on %s", t.laddr.String())
	// start listening on that addr
	err := t.listen()
	if err != nil {
		logrus.Error(err)
		return
	}

	go func() {
		// listen for ever
		for {
			c, err := t.listener.AcceptTCP()
			if err != nil {
				logrus.Errorf("Cannot accept new tcp connection %s", err.Error())
			} else {
				// consume the connection
				logrus.Infof("Accpeted new tcp connection from %s", c.RemoteAddr().String())
				go t.consume(c, p)
			}
		}
	}()
}

func (t *TCPProvider) consume(c *net.TCPConn, p event.Passer) {
	// create a newman connection
	conn := newman.NewConn(c)

	// drain the connection for ever
	in, _ := conn.Generate(func() newman.Message {
		return event.NewEvent()
	})
	for raw := range in {
		// convert it to an event because we know that is what we are getting
		e := raw.(*event.Event)

		// pass it on to the next step
		p.Pass(e)
	}

	// when it is done, close the connection
	c.Close()

}

func (t *TCPProvider) listen() error {
	l, err := net.ListenTCP("tcp", t.laddr)
	if err != nil {
		logrus.Error(err)
		return err
	}

	t.listener = l
	return nil
}
