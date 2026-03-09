package epi

type EnvStateProvider interface {
	// not intended to be used by an Environment user.
	getEnvState() EnvState
}

type Group struct {
	env EnvState
}

var _ EnvStateProvider = (*Group)(nil)

func NewGroup(env EnvState) *Group {
	return &Group{env: env}
}

func (g *Group) getEnvState() EnvState {
	return g.env
}

// With adds some state to the environment processing.
func (g *Group) With(state any, body ...Block) {
	EvaluateWithState[None](1, g.env, "With", "", &stateFrame{state: state}, nil, nil, body)
}

func (g *Group) Cleanup() {
	g.env.FailIfError(1, g.env.Cleanup())
}

// With adds some state to the environment processing.
func (g *Group) AddState(state any) {
	g.env.AddState(state)
}

// FailIfError fails if a non-nil error is given.
func (g *Group) FailIfError(err error) {
	g.env.FailIfError(1, err)
}

// FailIfErrorf fails if a non-nil error is given.
func (g *Group) FailIfErrorf(err error, msg string, args ...interface{}) {
	g.env.FailIfErrorf(1, err, msg, args...)
}

// FailIfErrorWithOffset fails if a nin-nil error is given.
func (g *Group) FailIfErrorWithOffset(skip int, err error) {
	g.env.FailIfError(skip+1, err)
}

// FailIfErrorWithOffsetf fails if a nin-nil error is given.
func (g *Group) FailIfErrorWithOffsetf(skip int, err error, msg string, args ...interface{}) {
	g.env.FailIfErrorf(skip+1, err, msg, args...)
}

func (g *Group) Compose(block Block) (err error) {
	return g.env.Compose(block)
}
