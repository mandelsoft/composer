package filesystem

import (
	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type DirectoryState interface {
	GetDir() (vfs.FileSystem, string)
}

type _dirState struct {
	fs  vfs.FileSystem
	dir string
}

func (s *_dirState) GetDir() (vfs.FileSystem, string) {
	return s.fs, s.dir
}

type dirFrame struct {
	name string
	mode vfs.FileMode
	_dirState
}

var (
	_ epi.Frame      = (*dirFrame)(nil)
	_ DirectoryState = (*dirFrame)(nil)
)

func (f *dirFrame) Setup(s DirectoryState) (epi.Frame, error) {
	fs, dir := s.GetDir()
	f.dir = vfs.Join(fs, dir, f.name)
	f.fs = fs
	err := fs.MkdirAll(f.dir, f.mode)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (f *dirFrame) Close() error {
	return nil
}

func (b *Group) Directory(name string, mode vfs.FileMode, f ...epi.Block) {
	if mode == 0 {
		mode = 0660
	}
	epi.EvaluateWithState[DirectoryState](1, b.env, "directory required", (&dirFrame{name: name, mode: mode}).Setup, f...)
}
