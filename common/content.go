package common

import (
	"io"

	"github.com/mandelsoft/composer/epi"
)

type Writable interface {
	GetWriter() io.Writer
}

type writer struct {
	epi.DefaultFrame[Writable]
	data []byte
}

func (e *writer) Setup(elem string, s Writable) (epi.Frame, error) {
	w := s.GetWriter()
	data := e.data
	for len(data) > 0 {
		written, err := w.Write(e.data)
		if err != nil {
			return nil, err
		}
		data = data[written:]
	}
	return e.DefaultFrame.Setup(elem, s)
}

func (b *Group) StringContent(s string) {
	epi.EvaluateLeafWithState[Writable](1, b.env, "StringContent", "writable element required in outer scope", &writer{data: []byte(s)}, nil, nil)
}

func (b *Group) ByteContent(s []byte) {
	epi.EvaluateLeafWithState[Writable](1, b.env, "ByteContent", "writable element required in outer scope", &writer{data: s}, nil, nil)

}
