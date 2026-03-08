package filesystem

import (
	"github.com/mandelsoft/composer/common"
	"github.com/mandelsoft/composer/epi"
)

// --- begin environment ---

type Environment struct {
	common.Environment
	FilesystemGroup
}

// --- end environment ---

// --- begin constructor ---

func New(opts ...epi.Option) (*Environment, error) {
	e := epi.NewEnvState(opts...)
	c, err := common.New(e)
	if err != nil {
		return nil, err
	}
	env := &Environment{Environment: *c, FilesystemGroup: *NewGroup(e)}
	return epi.WithOptionsApplied(env, opts...)
}

// --- end constructor ---
