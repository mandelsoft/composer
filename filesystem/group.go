package filesystem

import (
	"github.com/mandelsoft/composer/epi"
)

type FilesystemGroup = Group

// --- begin group ---
type Group struct {
	env epi.EnvState
}

var _ GroupMapper = (*Group)(nil)

func NewGroup(env epi.EnvState) *Group {
	return &Group{env: env}
}

func (g *Group) maptoFilesystemGroup() *Group {
	return g
}

// --- end group ---
