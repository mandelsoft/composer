package common

import (
	"github.com/mandelsoft/composer/epi"
)

type CommonGroup = Group

type Group struct {
	env epi.EnvState
}

var _ GroupMapper = (*Group)(nil)

func NewGroup(env epi.EnvState) *Group {
	return &Group{env: env}
}

func (g *Group) maptoCommonGroup() *Group {
	return g
}
