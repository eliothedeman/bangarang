package newman

import "encoding"

type Message interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}
