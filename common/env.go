package common

import (
	"github.com/mandelsoft/composer/epi"
)

type Environment struct {
	env epi.EnvState
	epi.Group
	CommonGroup
}

func New(opts ...epi.Option) (*Environment, error) {
	e := epi.NewEnvState(opts...)
	env := &Environment{env: e, CommonGroup: *NewGroup(e), Group: *epi.NewGroup(e)}
	err := epi.ApplyOptionsTo(env, opts...)
	if err != nil {
		return nil, err
	}
	return env, nil
}
