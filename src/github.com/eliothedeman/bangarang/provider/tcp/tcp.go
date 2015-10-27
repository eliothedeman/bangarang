package tcp

import (
	"encoding/binary"
	"io"
	"net"
	"runtime"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/bangarang/provider"
)

const START_HANDSHAKE = "BANGARANG: TCP_PROVIDER"

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

func readFull(conn *net.TCPConn, buff []byte) error {
	off := 0
	slp := time.Millisecond
	for off < len(buff) {
		n, err := conn.Read(buff[off:])
		if err != nil {
			return err
		}

		// exponentially back off if we don't have anything
		if n == 0 {
			slp = slp * 2
			time.Sleep(slp)
		}
		off += n
	}
	return nil
}

func (t *TCPProvider) consume(conn *net.TCPConn, p event.Passer) {
	buff := make([]byte, 1024*200)
	var size_buff = make([]byte, 8)
	var nextEventSize uint64
	var err error

	// write the start of the handshake so the client can verify this is a bangarang client
	conn.Write([]byte(START_HANDSHAKE))
	for {

		// read the size of the next event
		err = readFull(conn, size_buff)
		if err != nil {

			if err == io.EOF {
				time.Sleep(50 * time.Millisecond)
			} else {
				logrus.Error(err)
				return
			}
		} else {

			nextEventSize = binary.LittleEndian.Uint64(size_buff)

			// read the next event
			err = readFull(conn, buff[:nextEventSize])
			if err != nil {
				logrus.Error(err)
				conn.Close()
				return
			}
			e := &event.Event{}

			err = t.pool.Decode(buff[:nextEventSize], e)

			if err != nil {
				logrus.Error(err, string(buff[:nextEventSize]))
			} else {
				p.Pass(e)
			}
		}
	}
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
