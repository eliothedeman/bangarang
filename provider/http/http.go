package http

import (
	"fmt"
	"io/ioutil"
	"net"
	std_http "net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/bangarang/provider"
)

const (
	START_HANDSHAKE = "BANGARANG: HTTP_PROVIDER"
	ENDPOINT        = "/ingest"
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
		logrus.Error(err)

		// check to see if the busy port is due to another provider
		resp, err := std_http.Get("http://" + conf.Listen + ENDPOINT + "?init_check=true")
		if err != nil {
			return err
		}

		buff, _ := ioutil.ReadAll(resp.Body)
		if string(buff) != START_HANDSHAKE {
			return err

		}
	}

	// stop the test
	err = l.Close()
	if err != nil {
		logrus.Error(err)
		return err
	}

	// update the providers litening address
	t.listen = conf.Listen
	t.pool = event.NewEncodingPool(event.EncoderFactories[conf.Encoding], event.DecoderFactories[conf.Encoding], conf.MaxEncoders)

	return nil
}

func (t *HTTPProvider) ConfigStruct() interface{} {
	return &HTTPConfig{
		MaxEncoders: 4,
		Encoding:    event.ENCODING_TYPE_JSON,
	}
}

// start accepting connections and consume each of them as they come in
func (h *HTTPProvider) Start(p event.EventPasser) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	std_http.HandleFunc(ENDPOINT, func(w std_http.ResponseWriter, r *std_http.Request) {

		// handle the case where a provider is restarting and needs to check if a listener is a bangarang provider or not
		if r.URL.Query().Get("init_check") == "true" {
			w.Write([]byte(START_HANDSHAKE))
		}
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logrus.Error(err)
			std_http.Error(w, err.Error(), std_http.StatusInternalServerError)
			return
		}

		e := &event.Event{}
		err = h.pool.Decode(buff, e)

		if err != nil {
			logrus.Error(err)
			std_http.Error(w, err.Error(), std_http.StatusInternalServerError)
			return
		}
		p.PassEvent(e)
		logrus.Debug("Done processing http event")
	})

	logrus.Infof("Serving http listener on %s", h.listen)
	logrus.Fatal(std_http.ListenAndServe(h.listen, nil))
}
