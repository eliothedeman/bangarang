package event

type BinEncoder struct {
}

func NewBinEncoder() Encoder {
	return &BinEncoder{}
}

func (m *BinEncoder) Encode(e *Event) (buff []byte, err error) {
	buff, err = e.MarshalBinary()
	return
}

func NewBinDecoder() Decoder {
	return &BinDecoder{}
}

type BinDecoder struct {
}

func (m *BinDecoder) Decode(raw []byte, e *Event) error {
	return e.UnmarshalBinary(raw)
}
