package filesystem

import (
	"fmt"

	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

func (g *Group) ReadFile(name string) []byte {
	s, ok := epi.GetState[DirectoryState](g.env)
	if !ok {
		g.env.FailIfError(1, fmt.Errorf("ReadFile requires a directory"))
	}
	fs, dir := s.GetDir()
	data, err := vfs.ReadFile(fs, vfs.Join(fs, dir, name))
	g.env.FailIfError(1, err)
	return data
}
