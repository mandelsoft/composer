package epi

// None represents no (required) state.
// It is only implemented by the initial empty frame
// of an environment.
type None interface {
	isNone() bool
}

type Frame interface {
	Close() error
}

// StateProvider is an optional interface for a Frame
// to provide an explicit state representation.
type StateProvider interface {
	GetState() any
}

type initialFrame struct{}

func (e initialFrame) Close() error {
	return nil
}

func (e initialFrame) isNone() bool {
	return true
}

////////////////////////////////////////////////////////////////////////////////

type DefaultFrame[S any] struct {
}

func (e *DefaultFrame[S]) Close() error {
	return nil
}

func (e *DefaultFrame[S]) Setup(S) (Frame, error) {
	// no frame provided, we don't have an extended self in Go!!!!
	// therefore returning e would be contra-productive.
	return nil, nil
}

func EmptyFrameProvider(None) (Frame, error) {
	return &DefaultFrame[None]{}, nil
}

////////////////////////////////////////////////////////////////////////////////

type FrameSetup[S any] interface {
	Frame
	Setup(S) (Frame, error)
}
