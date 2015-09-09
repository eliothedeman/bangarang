package event

type MsgPackEncoder struct {
}

func NewMsgPackEncoder() Encoder {
	return &MsgPackEncoder{}
}

func (m *MsgPackEncoder) Encode(e *Event) (buff []byte, err error) {
	buff, err = e.MarshalMsg(nil)
	return
}

func NewMsgPackDecoder() Decoder {
	return &MsgPackDecoder{}
}

type MsgPackDecoder struct {
}

func (m *MsgPackDecoder) Decode(raw []byte, e *Event) error {
	_, err := e.UnmarshalMsg(raw)
	return err
}
