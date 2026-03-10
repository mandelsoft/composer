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

func (f *_fsState) Close() error {
	return vfs.Cleanup(f.fs)
}

// GetDir additionally provides initial dirState.
func (f *_fsState) GetDir() (vfs.FileSystem, string) {
	return f.fs, "/"
}

type fsFrame struct {
	epi.DefaultFrame[epi.None]
	_fsState
}

func (f *fsFrame) Close() error {
	return f._fsState.Close()
}

var (
	_ FilesystemState = (*fsFrame)(nil)
	_ DirectoryState  = (*fsFrame)(nil)
)

func (g *Group) FileSystem(fs vfs.FileSystem, f ...epi.Block) {
	if len(f) == 0 {
		g.env.AddState(&_fsState{fs: saveFS(fs, true)})
	} else {
		epi.EvaluateWithState[epi.None](1, g.env, "FileSystem", "", &fsFrame{_fsState: _fsState{fs: fs}}, nil, nil, f)
	}
}

func (g *Group) GetFilesystem() vfs.FileSystem {
	fs, ok := epi.GetState[FilesystemState](g.env)
	if !ok {
		return nil
	}
	return fs.GetFilesystem()
}

////////////////////////////////////////////////////////////////////////////////

type _saveFS struct {
	vfs.FileSystem
}

func (s *_saveFS) Cleanup() error {
	return nil
}

func saveFS(fs vfs.FileSystem, cleanup bool) vfs.FileSystem {
	if _, ok := fs.(*_saveFS); ok {
		return fs
	}
	if _, ok := fs.(vfs.FileSystemCleanup); !cleanup && ok {
		return &_saveFS{fs}
	}
	return fs
}

func effFS(fs vfs.FileSystem) vfs.FileSystem {
	if s, ok := fs.(*_saveFS); ok {
		return s.FileSystem
	}
	return fs
}
