package tcp

import (
	"bytes"
	"log"
	"net"
	"runtime"

	"github.com/eliothedeman/bangarang/client"
	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/bangarang/provider"
)

func init() {
	provider.LoadEventProviderFactory("tcp", NewTCPProvider)
}

// provides events from tcp connections
type TCPProvider struct {
	encoding string
	pool     *event.EncodingPool
	laddr    *net.TCPAddr
	listener *net.TCPListener
}

func NewTCPProvider() provider.EventProvider {
	return &TCPProvider{}
}

// the config struct for the tcp provider
type TCPConfig struct {
	Encoding    string `json:"encoding"`
	Listen      string `json:"listen"`
	MaxDecoders int    `json:"max_decoders"`
}

func (t *TCPProvider) Init(i interface{}) error {
	c := i.(*TCPConfig)

	// make sure we have a valid address
	addr, err := net.ResolveTCPAddr("tcp4", c.Listen)
	if err != nil {
		return err
	}

	t.laddr = addr

	// start listening on that addr
	err = t.listen()
	if err != nil {
		return err
	}

	// build an encoding pool
	t.pool = event.NewEncodingPool(event.EncoderFactories[c.Encoding], event.DecoderFactories[c.Encoding], c.MaxDecoders)
	return nil
}

func (t *TCPProvider) ConfigStruct() interface{} {
	return &TCPConfig{
		Encoding:    event.ENCODING_TYPE_JSON,
		MaxDecoders: runtime.NumCPU(),
	}
}

// start accepting connections and consume each of them as they come in
func (t *TCPProvider) Start(dst chan *event.Event) {

	// listen for ever
	for {
		c, err := t.listener.AcceptTCP()
		if err != nil {
			log.Println(err)
		} else {
			// consume the connection
			t.consume(c, dst)
		}
	}
}

func (t *TCPProvider) consume(conn *net.TCPConn, dst chan *event.Event) {
	buff := make([]byte, 1024*200)
	var e *event.Event
	for {
		n, err := conn.Read(buff)
		if err != nil {
			log.Println(err)
			conn.Close()
			return
		}

		buffs := bytes.Split(buff[:n], client.DELIMITER)
		for _, b := range buffs {
			if len(b) > 2 {
				t.pool.Decode(func(d event.Decoder) {
					e, err = d.Decode(b)
				})

				if err != nil {
					log.Println(err)
				} else {
					dst <- e
				}
			}
		}
	}
}

func (t *TCPProvider) listen() error {
	l, err := net.ListenTCP("tcp", t.laddr)
	if err != nil {
		return err
	}

	t.listener = l
	return nil
}
