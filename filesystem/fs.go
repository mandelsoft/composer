package filesystem

import (
	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type FilesystemState interface {
	GetFilesystem() vfs.FileSystem
}

type _fsState struct {
	fs vfs.FileSystem
}

func (f *_fsState) GetFilesystem() vfs.FileSystem {
	return f.fs
}

// GetDir additionally provides initial dirState.
func (f *_fsState) GetDir() (vfs.FileSystem, string) {
	return f.fs, "/"
}

type fsFrame struct {
	epi.DefaultFrame[epi.None]
	_fsState
}

var (
	_ FilesystemState = (*fsFrame)(nil)
	_ DirectoryState  = (*fsFrame)(nil)
)

func (g *Group) WithFileSystem(fs vfs.FileSystem, f epi.Block) {
	epi.EvaluateWithState[epi.None](1, g.env, "", (&fsFrame{_fsState: _fsState{fs: fs}}).Setup, f)
}
