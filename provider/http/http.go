package http

import (
	"io/ioutil"
	"log"
	"net"
	std_http "net/http"

	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/bangarang/provider"
)

func init() {
	provider.LoadEventProviderFactory("http", NewHTTPProvider)
}

// provides events from HTTP connections
type HTTPProvider struct {
	pool   *event.EncodingPool
	listen string
}

func NewHTTPProvider() provider.EventProvider {
	return &HTTPProvider{}
}

// the config struct for the HTTP provider
type HTTPConfig struct {
	Encoding    string `json:"encoding"`
	Listen      string `json:"listen"`
	MaxEncoders int    `json:"max_encoders"`
}

func (t *HTTPProvider) Init(i interface{}) error {
	conf := i.(*HTTPConfig)

	// make sure the port is open
	l, err := net.Listen("tcp", conf.Listen)
	if err != nil {
		return err
	}

	// stop the test
	err = l.Close()
	if err != nil {
		return err
	}

	t.pool = event.NewEncodingPool(event.EncoderFactories[conf.Encoding], event.DecoderFactories[conf.Encoding], conf.MaxEncoders)

	return nil
}

func (t *HTTPProvider) ConfigStruct() interface{} {
	return &HTTPConfig{}
}

// start accepting connections and consume each of them as they come in
func (h *HTTPProvider) Start(dst chan *event.Event) {
	std_http.HandleFunc("/ingest", func(w std_http.ResponseWriter, r *std_http.Request) {
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			std_http.Error(w, err.Error(), std_http.StatusInternalServerError)
			return
		}
		var e *event.Event
		h.pool.Decode(func(d event.Decoder) {
			e, err = d.Decode(buff)
		})
		if err != nil {
			log.Println(err)
			std_http.Error(w, err.Error(), std_http.StatusInternalServerError)
			return
		}
		dst <- e
	})

	log.Fatal(std_http.ListenAndServe(h.listen, nil))
}
