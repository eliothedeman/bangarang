package event

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Event) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "host":
			z.Host, err = dc.ReadString()
			if err != nil {
				return
			}
		case "service":
			z.Service, err = dc.ReadString()
			if err != nil {
				return
			}
		case "sub_service":
			z.SubService, err = dc.ReadString()
			if err != nil {
				return
			}
		case "metric":
			z.Metric, err = dc.ReadFloat64()
			if err != nil {
				return
			}
		case "occurences":
			z.Occurences, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "tags":
			var msz uint32
			msz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Tags == nil && msz > 0 {
				z.Tags = make(map[string]string, msz)
			} else if len(z.Tags) > 0 {
				for key, _ := range z.Tags {
					delete(z.Tags, key)
				}
			}
			for msz > 0 {
				msz--
				var xvk string
				var bzg string
				xvk, err = dc.ReadString()
				if err != nil {
					return
				}
				bzg, err = dc.ReadString()
				if err != nil {
					return
				}
				z.Tags[xvk] = bzg
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
func (z *Event) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(6)
	if err != nil {
		return
	}
	err = en.WriteString("host")
	if err != nil {
		return
	}
	err = en.WriteString(z.Host)
	if err != nil {
		return
	}
	err = en.WriteString("service")
	if err != nil {
		return
	}
	err = en.WriteString(z.Service)
	if err != nil {
		return
	}
	err = en.WriteString("sub_service")
	if err != nil {
		return
	}
	err = en.WriteString(z.SubService)
	if err != nil {
		return
	}
	err = en.WriteString("metric")
	if err != nil {
		return
	}
	err = en.WriteFloat64(z.Metric)
	if err != nil {
		return
	}
	err = en.WriteString("occurences")
	if err != nil {
		return
	}
	err = en.WriteInt(z.Occurences)
	if err != nil {
		return
	}
	err = en.WriteString("tags")
	if err != nil {
		return
	}
	err = en.WriteMapHeader(uint32(len(z.Tags)))
	if err != nil {
		return
	}
	for xvk, bzg := range z.Tags {
		err = en.WriteString(xvk)
		if err != nil {
			return
		}
		err = en.WriteString(bzg)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Event) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendMapHeader(o, 6)
	o = msgp.AppendString(o, "host")
	o = msgp.AppendString(o, z.Host)
	o = msgp.AppendString(o, "service")
	o = msgp.AppendString(o, z.Service)
	o = msgp.AppendString(o, "sub_service")
	o = msgp.AppendString(o, z.SubService)
	o = msgp.AppendString(o, "metric")
	o = msgp.AppendFloat64(o, z.Metric)
	o = msgp.AppendString(o, "occurences")
	o = msgp.AppendInt(o, z.Occurences)
	o = msgp.AppendString(o, "tags")
	o = msgp.AppendMapHeader(o, uint32(len(z.Tags)))
	for xvk, bzg := range z.Tags {
		o = msgp.AppendString(o, xvk)
		o = msgp.AppendString(o, bzg)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Event) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "host":
			z.Host, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "service":
			z.Service, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "sub_service":
			z.SubService, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "metric":
			z.Metric, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				return
			}
		case "occurences":
			z.Occurences, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "tags":
			var msz uint32
			msz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Tags == nil && msz > 0 {
				z.Tags = make(map[string]string, msz)
			} else if len(z.Tags) > 0 {
				for key, _ := range z.Tags {
					delete(z.Tags, key)
				}
			}
			for msz > 0 {
				var xvk string
				var bzg string
				msz--
				xvk, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				bzg, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				z.Tags[xvk] = bzg
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

func (z *Event) Msgsize() (s int) {
	s = msgp.MapHeaderSize + msgp.StringPrefixSize + 4 + msgp.StringPrefixSize + len(z.Host) + msgp.StringPrefixSize + 7 + msgp.StringPrefixSize + len(z.Service) + msgp.StringPrefixSize + 11 + msgp.StringPrefixSize + len(z.SubService) + msgp.StringPrefixSize + 6 + msgp.Float64Size + msgp.StringPrefixSize + 10 + msgp.IntSize + msgp.StringPrefixSize + 4 + msgp.MapHeaderSize
	if z.Tags != nil {
		for xvk, bzg := range z.Tags {
			_ = bzg
			s += msgp.StringPrefixSize + len(xvk) + msgp.StringPrefixSize + len(bzg)
		}
	}
	return
}
