package epi

import (
	"io"

	"github.com/mandelsoft/composer/epi/internal"
)

// None represents no (required) state.
// It is only implemented by the initial empty frame
// of an environment.
type None interface {
	isNone() bool
}

type Frame = internal.Frame

type DefaultFrame[S any] struct {
	elem string
}

func (f *DefaultFrame[S]) SetElem(n string) {
	f.elem = n
}

func (f *DefaultFrame[S]) Element() string {
	return f.elem
}

func (f *DefaultFrame[S]) Setup(elem string, state S) (Frame, error) {
	f.elem = elem
	// no frame provided, we don't have an extended self in Go!!!!
	// therefore, returning f would be contra-productive.
	// We use nil as fallback for self handled by the caller below.
	return nil, nil
}

func (f *DefaultFrame[S]) Close() error {
	return nil
}

// StateProvider is an optional interface for a Frame
// to provide an explicit state representation.
type StateProvider = internal.StateProvider

type initialFrame struct {
	DefaultFrame[None]
}

func (e initialFrame) isNone() bool {
	return true
}

type dummyFrame struct {
	DefaultFrame[None]
}

////////////////////////////////////////////////////////////////////////////////

func IsStateFrame(f Frame) bool {
	_, ok := f.(*stateFrame)
	return ok
}

func IsDummyFrame(f Frame) bool {
	_, ok := f.(*dummyFrame)
	return ok
}

func IsInitialFrame(f Frame) bool {
	i, ok := f.(None)
	return ok && i.isNone()
}

func IsElementFrame(f Frame) bool {
	return !IsDummyFrame(f) && !IsStateFrame(f) && !IsInitialFrame(f)
}

////////////////////////////////////////////////////////////////////////////////

func EmptyFrameProvider(None) (Frame, error) {
	return &DefaultFrame[None]{}, nil
}

////////////////////////////////////////////////////////////////////////////////

type stateFrame struct {
	DefaultFrame[None]
	state any
}

var _ StateProvider = (*stateFrame)(nil)

func (f *stateFrame) Close() error {
	if f.state == nil {
		return nil
	}
	if c, ok := f.state.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (f *stateFrame) GetState() any {
	return f.state
}
