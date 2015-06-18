package event

import (
	"bytes"

	"github.com/pquerna/ffjson/ffjson"
)

type JsonEncoder struct {
	enc  *ffjson.Encoder
	buff *bytes.Buffer
}

func NewJsonEncoder() Encoder {
	je := &JsonEncoder{
		buff: bytes.NewBuffer(make([]byte, 1024*200)),
	}
	je.enc = ffjson.NewEncoder(je.buff)
	return je
}

func (j *JsonEncoder) Encode(e *Event) ([]byte, error) {
	j.buff.Reset()
	err := j.enc.EncodeFast(e)
	return j.buff.Bytes(), err
}

func NewJsonDecoder() Decoder {
	return &JsonDecoder{}
}

type JsonDecoder struct {
}

func (j *JsonDecoder) Decode(raw []byte) (*Event, error) {
	e := &Event{}
	err := ffjson.UnmarshalFast(raw, e)
	return e, err
}
