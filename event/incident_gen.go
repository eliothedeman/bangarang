package event

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Incident) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "event_name":
			z.EventName, err = dc.ReadBytes(z.EventName)
			if err != nil {
				return
			}
		case "time":
			z.Time, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "id":
			z.Id, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "active":
			z.Active, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "status":
			z.Status, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "escalation":
			z.Escalation, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Incident) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(6)
	if err != nil {
		return
	}
	err = en.WriteString("event_name")
	if err != nil {
		return
	}
	err = en.WriteBytes(z.EventName)
	if err != nil {
		return
	}
	err = en.WriteString("time")
	if err != nil {
		return
	}
	err = en.WriteInt64(z.Time)
	if err != nil {
		return
	}
	err = en.WriteString("id")
	if err != nil {
		return
	}
	err = en.WriteInt64(z.Id)
	if err != nil {
		return
	}
	err = en.WriteString("active")
	if err != nil {
		return
	}
	err = en.WriteBool(z.Active)
	if err != nil {
		return
	}
	err = en.WriteString("status")
	if err != nil {
		return
	}
	err = en.WriteInt(z.Status)
	if err != nil {
		return
	}
	err = en.WriteString("escalation")
	if err != nil {
		return
	}
	err = en.WriteString(z.Escalation)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Incident) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendMapHeader(o, 6)
	o = msgp.AppendString(o, "event_name")
	o = msgp.AppendBytes(o, z.EventName)
	o = msgp.AppendString(o, "time")
	o = msgp.AppendInt64(o, z.Time)
	o = msgp.AppendString(o, "id")
	o = msgp.AppendInt64(o, z.Id)
	o = msgp.AppendString(o, "active")
	o = msgp.AppendBool(o, z.Active)
	o = msgp.AppendString(o, "status")
	o = msgp.AppendInt(o, z.Status)
	o = msgp.AppendString(o, "escalation")
	o = msgp.AppendString(o, z.Escalation)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Incident) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "event_name":
			z.EventName, bts, err = msgp.ReadBytesBytes(bts, z.EventName)
			if err != nil {
				return
			}
		case "time":
			z.Time, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "id":
			z.Id, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "active":
			z.Active, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "status":
			z.Status, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "escalation":
			z.Escalation, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *Incident) Msgsize() (s int) {
	s = msgp.MapHeaderSize + msgp.StringPrefixSize + 10 + msgp.BytesPrefixSize + len(z.EventName) + msgp.StringPrefixSize + 4 + msgp.Int64Size + msgp.StringPrefixSize + 2 + msgp.Int64Size + msgp.StringPrefixSize + 6 + msgp.BoolSize + msgp.StringPrefixSize + 6 + msgp.IntSize + msgp.StringPrefixSize + 10 + msgp.StringPrefixSize + len(z.Escalation)
	return
}
