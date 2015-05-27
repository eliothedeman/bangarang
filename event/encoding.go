package event

const (
	ENCODING_TYPE_JSON    = "json"
	ENCODING_TYPE_MSGPACK = "msgp"
)

var (
	EncoderFactories = map[string]EncoderFactory{
		"json": NewJsonEncoder,
		"msgp": NewMsgPackEncoder,
	}
	DecoderFactories = map[string]DecoderFactory{
		"json": NewJsonDecoder,
		"msgp": NewMsgPackDecoder,
	}
)

// encodes the given event and sends the encoded value on the dst channel
type Encoder interface {
	Encode(e *Event) ([]byte, error)
}

// decodes the raw event and sends the decoded event to it's destination
type Decoder interface {
	Decode(raw []byte) (*Event, error)
}

type EncoderDecoder interface {
	Encoder
	Decoder
}

type EncodingPool struct {
	sigBuff     chan chan struct{}
	encoders    chan Encoder
	decoders    chan Decoder
	encWorkChan chan EncFunc
	decWorkChan chan DecFunc
}

type EncoderFactory func() Encoder
type DecoderFactory func() Decoder

func NewEncodingPool(enc EncoderFactory, dec DecoderFactory, maxSize int) *EncodingPool {
	p := &EncodingPool{
		encoders:    make(chan Encoder, maxSize),
		decoders:    make(chan Decoder, maxSize),
		encWorkChan: make(chan EncFunc, maxSize),
		decWorkChan: make(chan DecFunc, maxSize),
		sigBuff:     make(chan chan struct{}, maxSize*2),
	}

	// fill the encoders/decoders
	for i := 0; i < maxSize; i++ {
		p.encoders <- enc()
		p.decoders <- dec()
		p.sigBuff <- make(chan struct{})
		p.sigBuff <- make(chan struct{})
	}

	// start your engines!
	p.manage()

	return p
}

type EncFunc func(e Encoder)
type DecFunc func(d Decoder)

// starts up a goroutine for each encoder/decoder and managages incoming work on them
func (t *EncodingPool) manage() {
	// manage encoders
	for i := 0; i < cap(t.encoders); i++ {
		go func() {
			var work EncFunc
			var enc Encoder
			for {
				work = <-t.encWorkChan
				enc = <-t.encoders
				work(enc)
				t.encoders <- enc
			}
		}()
	}

	// manage decoders
	for i := 0; i < cap(t.decoders); i++ {
		go func() {
			var work DecFunc
			var dec Decoder
			for {
				work = <-t.decWorkChan
				dec = <-t.decoders
				work(dec)
				t.decoders <- dec
			}
		}()
	}
}

// use one of the pooled encoders to encode the given event
func (t *EncodingPool) Encode(e EncFunc) {
	sig := <-t.sigBuff
	t.encWorkChan <- func(enc Encoder) {
		e(enc)
		sig <- struct{}{}
	}
	<-sig
	t.sigBuff <- sig
}

// use one of the pooled decoders to decode the given event
func (t *EncodingPool) Decode(d DecFunc) {
	sig := <-t.sigBuff
	t.decWorkChan <- func(dec Decoder) {
		d(dec)
		sig <- struct{}{}
	}

	<-sig
	t.sigBuff <- sig
}
