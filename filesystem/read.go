package filesystem

import (
	"fmt"

	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

func getPath(s DirectoryState, path string) (vfs.FileSystem, string) {
	fs, dir := s.GetDir()
	if vfs.IsAbs(fs, path) {
		return fs, path
	}
	return fs, vfs.Join(fs, dir, path)
}

func (g *Group) ReadFile(name string) []byte {
	s, ok := epi.GetState[DirectoryState](g.env)
	if !ok {
		g.env.FailIfError(1, fmt.Errorf("ReadFile requires a directory"))
	}
	fs, name := getPath(s, name)
	data, err := vfs.ReadFile(fs, name)
	g.env.FailIfError(1, err)
	return data
}

func (g *Group) FileExists(name string) bool {
	s, ok := epi.GetState[DirectoryState](g.env)
	if !ok {
		g.env.FailIfError(1, fmt.Errorf("ExistsFile requires a directory"))
	}
	fs, name := getPath(s, name)
	data, err := vfs.FileExists(fs, name)
	g.env.FailIfError(1, err)
	return data
}

func (g *Group) DirExists(name string) bool {
	s, ok := epi.GetState[DirectoryState](g.env)
	if !ok {
		g.env.FailIfError(1, fmt.Errorf("DirExists requires a directory"))
	}
	fs, name := getPath(s, name)
	data, err := vfs.DirExists(fs, name)
	g.env.FailIfError(1, err)
	return data
}
