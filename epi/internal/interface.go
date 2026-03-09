// common stuff which is shared by multiple epi packages is put
// into a provate package to avoid package cycles.

package internal

type Frame interface {
	Element() string
	Close() error
}

type StateProvider interface {
	GetState() any
}

func GetFrameState[S any](frame Frame) (S, bool) {
	var _nil S
	t := any(frame)
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
	return _nil, false
}
