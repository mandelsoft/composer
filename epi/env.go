package epi

type Environment interface {
	EnvStateProvider
	Compose(block Block) error

	With(state any, body Block)
	AddState(state any)

	FailIfError(error)
	FailIfErrorWithOffset(skip int, err error)
}

// EnvironmentFactory is the signature of a typical New
// function used to create a composed Environment.
// For example the New methods provided by functional
// areas to provide a default environment.
type EnvironmentFactory[E Environment] = func(...Option) (E, error)
