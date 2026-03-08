package epi

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/exception"
	"github.com/mandelsoft/goutils/general"
)

type Block = func()

type StateExtractor[S any] func(e EnvState) (S, bool)

////////////////////////////////////////////////////////////////////////////////

type EnvState interface {
	Option
	EnvStateProvider
	With(skip int, frame Frame, body ...Block)
	AddState(state any)
	GetFrames() []Frame
	FailIfError(skip int, err error)
	FailIfErrorf(skip int, err error, msg string, args ...interface{})
	Compose(block Block) (err error)
}

func Use(env ...EnvStateProvider) EnvState {
	e := general.Optional(env...)
	if e == nil {
		e = NewEnvState()
	}
	return e.getEnvState()
}

type _envstate struct {
	frames  []Frame
	failure FailureHandler
	err     *errors.ErrorList // occurred error
}

func NewEnvState(opts ...Option) EnvState {
	var fh FailureHandler

	for _, o := range opts {
		if f, ok := o.(FailureHandler); ok {
			fh = f
		}
		if e, ok := o.(EnvState); ok {
			return e
		}
		if p, ok := o.(EnvStateProvider); ok {
			return p.getEnvState()
		}
	}
	return &_envstate{frames: []Frame{initialFrame{}}, failure: general.OptionalDefaulted(FailWithExceptionLocation, fh), err: errors.ErrList()}
}

func (e2 *_envstate) ApplyTo(e Environment) error {
	// just a marker function to be usable as Option.
	return nil
}

func (e *_envstate) getEnvState() EnvState {
	return e
}

func (e *_envstate) GetFrames() []Frame {
	return e.frames
}

func (e *_envstate) FailIfError(skip int, err error) {
	if err != nil {
		e.err.Add(err)
		e.failure(skip+1, e, err)
	}
}

func (e *_envstate) FailIfErrorf(skip int, err error, msg string, args ...interface{}) {
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf(msg, args...))
		e.err.Add(err)
		e.failure(skip+1, e, err)
	}
}

func (e *_envstate) dropUntil(f Frame) {
	for e.frames[len(e.frames)-1] != f {
		e.frames = e.frames[:len(e.frames)-1]
	}
}

func (e *_envstate) AddState(state any) {
	e.frames = append(e.frames, &stateFrame{state: state})
}

func (e *_envstate) exec(body ...Block) {
	for _, b := range body {
		if b != nil {
			b()
		}
	}
}

func (e *_envstate) cleanup(skip int, frame Frame) {
	e.dropUntil(frame)
	e.frames = e.frames[:len(e.frames)-1]
	err := frame.Close()
	if e.err.Len() > 0 {
		// regular failure
		e.FailIfError(skip+1, err)
	} else {
		// already in error handling
		e.err.Add(err)
	}
}

func (e *_envstate) With(skip int, frame Frame, body ...Block) {
	skip++
	if frame == nil {
		frame = &DefaultFrame[None]{}
	}
	e.frames = append(e.frames, frame)

	defer func() {
		e.cleanup(skip, frame)
	}()
	e.exec(body...)
}

func (e *_envstate) Compose(block Block) (err error) {
	old := e.err
	e.err = errors.ErrList()
	defer func() {
		e.err = old
	}()
	return exception.Catch(block)
}

///////////////////////////////////////////////////////////////////////////////

func GetEnvState(a any) EnvState {
	if p, ok := a.(EnvStateProvider); ok {
		return p.getEnvState()
	}
	return nil
}

func GetState[S any](p EnvStateProvider, ext ...StateExtractor[S]) (S, bool) {
	e := p.getEnvState()
	var _nil S

	f := general.Optional(ext...)
	if f != nil {
		s, ok := f(e)
		if ok {
			return s, true
		}
	}
	frames := e.GetFrames()
	for i := len(frames) - 1; i >= 0; i-- {
		var t any = frames[i]
		for t != nil {
			if s, ok := t.(S); ok {
				return s, true
			}
			if p, ok := t.(StateProvider); ok {
				t = p.GetState()
			} else {
				break
			}
		}
	}
	return _nil, false
}

type FrameProvider[S any] interface {
	Setup(S) (Frame, error)
}

func splitPath(s string) (string, string) {
	idx := strings.LastIndex(s, "/")
	if idx < 0 {
		idx = strings.LastIndex(s, "\\")
	}
	if idx < 0 {
		return "", s
	}
	return s[:idx], s[idx+1:]
}

func CallerInfo(skip int, adjust ...int) string {
	pc, file, line, ok := runtime.Caller(skip + 1)
	line += general.Optional(adjust...)
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			pkg, name := splitPath(fn.Name())
			_, file := splitPath(file)
			return fmt.Sprintf("%s %s/%s:%d", name, pkg, file, line)
		}
		return fmt.Sprintf("%s:%d", file, line)
	}
	return ""
}

func EvaluateWithState[S any](skip int, e EnvState, msg string, p FrameProvider[S], f ...Block) {
	skip++
	s, ok := GetState[S](e)
	if !ok {
		e.FailIfError(skip, fmt.Errorf(msg))
	}
	frame, err := p.Setup(s)
	e.FailIfError(skip, err)
	if frame == nil {
		// no extended self for embedded default implementations,
		// therefore we default to the final top-level object
		if fr, ok := any(p).(Frame); ok {
			frame = fr
		}
	}
	if general.Optional(f...) != nil {
		e.With(skip, frame, f...)
	}
}

func EvaluateLeafWithState[S any](skip int, e EnvState, msg string, p FrameProvider[S]) {
	EvaluateWithState[S](skip+1, e, msg, p, nil)
}
