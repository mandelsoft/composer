package filesystem

import (
	"github.com/mandelsoft/composer/common"
	"github.com/mandelsoft/composer/epi"
)

type FilesystemEnvironment = Environment

type Environment struct {
	common.CommonEnvironment
	FilesystemGroup
}

func New(opts ...epi.Option) (*Environment, error) {
	e := epi.NewEnvState(opts...)
	c, err := common.New(e)
	if err != nil {
		return nil, err
	}
	env := &Environment{CommonEnvironment: *c, FilesystemGroup: *NewGroup(e)}
	return epi.WithOptionsApplied(env, opts...)
}
