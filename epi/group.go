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

type stateFrame struct {
	DefaultFrame[None]
	state any
}

var _ StateProvider = (*stateFrame)(nil)

func (f *stateFrame) Setup(None) (Frame, error) {
	return f, nil
}

func (f *stateFrame) GetState() any {
	return f.state
}

// With adds some state to the environment processing.
func (g *Group) With(state any, body Block) {
	EvaluateWithState[None](1, g.env, "", (&stateFrame{state: state}).Setup, body)
}

// With adds some state to the environment processing.
func (g *Group) AddState(state any) {
	g.env.AddState(state)
}

// FailIfError fails if a nin-nil error is given.
func (g *Group) FailIfError(err error) {
	g.env.FailIfError(1, err)
}

// FailIfErrorWithOffset fails if a nin-nil error is given.
func (g *Group) FailIfErrorWithOffset(skip int, err error) {
	g.env.FailIfError(skip+1, err)
}

func (g *Group) Compose(block Block) (err error) {
	return g.env.Compose(block)
}
