package epi

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/mandelsoft/composer/epi/contraints"
	"github.com/mandelsoft/composer/epi/internal"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/exception"
	"github.com/mandelsoft/goutils/general"
)

type Block = func()

type StateExtractor[S any] func(e EnvState) (state S, inner, outer []Frame, ok bool)

////////////////////////////////////////////////////////////////////////////////

type EnvState interface {
	Option
	EnvStateProvider
	Cleanup() error
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
	return &_envstate{frames: []Frame{&initialFrame{}}, failure: general.OptionalDefaulted(FailWithExceptionLocation, fh), err: errors.ErrList()}
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

func (e *_envstate) Cleanup() error {
	return e.dropUntil(nil) // drop state frames o top of element frames
}

func (e *_envstate) dropUntil(f Frame) error {
	list := errors.ErrorList{}
	for {
		if f != nil {
			if e.frames[len(e.frames)-1] == f {
				break
			}
		} else {
			// if no frame ids givven cleanup pure state frames
			// until the next outer element frame is reached
			if !IsStateFrame(e.frames[len(e.frames)-1]) {
				break
			}
		}
		list.Add(e.frames[len(e.frames)-1].Close())
		e.frames = e.frames[:len(e.frames)-1]
	}
	return list.Result()
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
	list := errors.ErrorList{}
	list.Add(e.dropUntil(frame))
	e.frames = e.frames[:len(e.frames)-1]
	list.Add(frame.Close())
	if e.err.Len() > 0 {
		// regular failure
		e.FailIfError(skip+1, list.Result())
	} else {
		// already in error handling
		e.err.Add(list.Entries()...)
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
	return exception.Catch(func() {
		EvaluateWithState[None](1, e, "Compose", "", &dummyFrame{}, nil, nil, []Block{block})
	})
}

///////////////////////////////////////////////////////////////////////////////

func GetEnvState(a any) EnvState {
	if p, ok := a.(EnvStateProvider); ok {
		return p.getEnvState()
	}
	return nil
}

func GetFrameState[S any](frame Frame) (S, bool) {
	return internal.GetFrameState[S](frame)
}

func GetState[S any](p EnvStateProvider, ext ...StateExtractor[S]) (S, bool) {
	s, _, _, ok := GetStateAndEmbedding(p, ext...)
	return s, ok
}

func GetStateAndEmbedding[S any](p EnvStateProvider, ext ...StateExtractor[S]) (S, []Frame, []Frame, bool) {
	e := p.getEnvState()
	var state S
	var inner []Frame
	var outer []Frame

	f := general.Optional(ext...)
	if f != nil {
		s, inner, outer, ok := f(e)
		if ok {
			return s, inner, outer, true
		}
	}
	frames := e.GetFrames()
	gather := &inner
	for i := len(frames) - 1; i >= 0; i-- {
		if IsElementFrame(frames[i]) {
			*gather = append(*gather, frames[i])
		}
		s, ok := GetFrameState[S](frames[i])
		if ok && gather == &inner {
			state = s
			gather = &outer
		}
	}
	return state, inner, outer, true
}

type FrameProvider[S any] interface {
	Setup(name string, state S) (Frame, error)
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

func EvaluateWithState[S any](skip int, e EnvState, name, msg string, p FrameProvider[S], ext StateExtractor[S], cs contraints.Constraint, f []Block) {
	skip++
	s, inner, outer, ok := GetStateAndEmbedding[S](e, ext)
	if !ok {
		e.FailIfError(skip, fmt.Errorf("%s: %s", name, msg))
	}
	if cs != nil {
		e.FailIfError(skip, errors.Wrap(cs(inner, outer), name))
	}
	frame, err := p.Setup(name, s)
	e.FailIfError(skip, errors.Wrap(err, name))
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

func EvaluateLeafWithState[S any](skip int, e EnvState, name, msg string, p FrameProvider[S], ext StateExtractor[S], cs contraints.Constraint) {
	EvaluateWithState[S](skip+1, e, name, msg, p, ext, cs, nil)
}
