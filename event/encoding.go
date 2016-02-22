package event

const (
	ENCODING_TYPE_JSON = "json"
	ENCODING_TYPE_BIN  = "bin"
)

var (
	EncoderFactories = map[string]EncoderFactory{
		ENCODING_TYPE_JSON: NewJsonEncoder,
		ENCODING_TYPE_BIN:  NewBinEncoder,
	}
	DecoderFactories = map[string]DecoderFactory{
		ENCODING_TYPE_JSON: NewJsonDecoder,
		ENCODING_TYPE_BIN:  NewBinDecoder,
	}
)

// encodes the given event and sends the encoded value on the dst channel
type Encoder interface {
	Encode(e *Event) ([]byte, error)
}

// decodes the raw event and sends the decoded event to it's destination
type Decoder interface {
	Decode(raw []byte, ev *Event) error
}

type EncoderDecoder interface {
	Encoder
	Decoder
}

type EncodingPool struct {
	encoders chan Encoder
	decoders chan Decoder
}

type EncoderFactory func() Encoder
type DecoderFactory func() Decoder

func NewEncodingPool(enc EncoderFactory, dec DecoderFactory, maxSize int) *EncodingPool {
	p := &EncodingPool{
		encoders: make(chan Encoder, maxSize),
		decoders: make(chan Decoder, maxSize),
	}

	// fill the encoders/decoders
	for i := 0; i < maxSize; i++ {
		p.encoders <- enc()
		p.decoders <- dec()
	}

	return p
}

// use one of the pooled encoders to encode the given event
func (t *EncodingPool) Encode(ev *Event) ([]byte, error) {
	// grab a decoder from the pool
	en := <-t.encoders

	// run the encoding
	buff, err := en.Encode(ev)

	// re-add the encoder to the pool
	t.encoders <- en

	return buff, err
}

// use one of the pooled decoders to decode the given event
func (t *EncodingPool) Decode(buff []byte, ev *Event) error {

	// grab a decoder from the pool
	d := <-t.decoders

	// run the decodeing
	err := d.Decode(buff, ev)

	// re-add the decoder to the pool
	t.decoders <- d

	return err
}
