package event

import "encoding/json"

type JsonEncoder struct {
}

func NewJsonEncoder() Encoder {
	je := &JsonEncoder{}
	return je
}

func (j *JsonEncoder) Encode(e *Event) ([]byte, error) {
	return json.Marshal(e)
}

func NewJsonDecoder() Decoder {
	return &JsonDecoder{}
}

type JsonDecoder struct {
}

func (j *JsonDecoder) Decode(raw []byte, e *Event) error {
	return json.Unmarshal(raw, e)
}
