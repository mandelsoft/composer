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

var (
	_ FilesystemState = (*_fsState)(nil)
	_ DirectoryState  = (*_fsState)(nil)
)

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

func (f *fsFrame) xSetup(epi.None) (epi.Frame, error) {
	return f, nil
}

var (
	_ FilesystemState = (*fsFrame)(nil)
	_ DirectoryState  = (*fsFrame)(nil)
)

func (g *Group) FileSystem(fs vfs.FileSystem, f ...epi.Block) {
	if len(f) == 0 {
		g.env.AddState(&_fsState{fs: fs})
	} else {
		epi.EvaluateWithState[epi.None](1, g.env, "", &fsFrame{_fsState: _fsState{fs: fs}}, f...)
	}
}
