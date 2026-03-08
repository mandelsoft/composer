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

func (e *writer) Setup(s Writable) (epi.Frame, error) {
	w := s.GetWriter()
	data := e.data
	for len(data) > 0 {
		written, err := w.Write(e.data)
		if err != nil {
			return nil, err
		}
		data = data[written:]
	}
	return nil, nil
}

func (b *Group) StringContent(s string) {
	epi.EvaluateLeafWithState[Writable](1, b.env, "", &writer{data: []byte(s)})
}

func (b *Group) ByteContent(s []byte) {
	epi.EvaluateLeafWithState[Writable](1, b.env, "", &writer{data: s})

}
