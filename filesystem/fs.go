package filesystem

import (
	"fmt"

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
	// return vfs.Cleanup(f.fs)
	return nil
}

// GetDir additionally provides initial dirState.
func (f *_fsState) GetDir() (vfs.FileSystem, string) {
	return f.fs, "/"
}

type fsFrame struct {
	epi.DefaultFrame[FilesystemState]
	_fsState
	fsin vfs.FileSystem
	mode []FileSystemModifier
}

var (
	_ epi.StateProvider = (*fsFrame)(nil)
	_ FilesystemState   = (*fsFrame)(nil)
	_ DirectoryState    = (*fsFrame)(nil)
)

func (f *fsFrame) GetState() any {
	return &f._fsState
}

func (f *fsFrame) Close() error {
	return f._fsState.Close()
}

func (e *fsFrame) Setup(elem string, s FilesystemState) (epi.Frame, error) {
	var err error
	var fsout vfs.FileSystem

	if s != nil {
		fsout = effFS(s.GetFilesystem())
	}
	if fsout == nil && e.fsin == nil {
		return nil, fmt.Errorf("filesystem required in outer frame or as argument")
	}
	if e.fsin == nil {
		e.fsin = fsout
	}

	e.fs, err = applyMountModes(e.fsin, e.mode...)
	if err != nil {
		return nil, err
	}

	return e.DefaultFrame.Setup(elem, s)
}

////////////////////////////////////////////////////////////////////////////////

// FileSystem accepts a filesystem to handle (defaulted by an optional outer state),
// FileSystemModifier(s) and nested blocks. Because it returns a vfs.FileSystem,
// it can also be uses as composition arg for itself.
// It provides an env state for the configured filesystem.
// If no nested blocks are given, the state is extended for subsequent method calls,
// otherwise it is only valid for the nested blocks.
func (g *Group) FileSystem(args ...any) vfs.FileSystem {
	fs, mode, f := g.evalElementArgs(1, args...)

	if len(fs) > 1 {
		g.env.FailIfError(1, fmt.Errorf("only one filesystem argument possible"))
	}

	var eff vfs.FileSystem = nil
	if len(fs) > 0 {
		eff = fs[0]
	}
	frame := &fsFrame{fsin: eff, mode: mode}
	// creates state frame for epi.StateProvider.GetState() without nested blocks or adds the nested frame.
	epi.EvaluateWithState[FilesystemState](1, g.env, "FileSystem", "", frame, epi.OptionalStateExtractor, nil, f)
	return frame.fs
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
	if fs != nil {
		if s, ok := fs.(*_saveFS); ok {
			return s.FileSystem
		}
	}
	return fs
}

////////////////////////////////////////////////////////////////////////////////

func (g *Group) evalElementArgs(skip int, args ...any) ([]vfs.FileSystem, []FileSystemModifier, []epi.Block) {
	var f []epi.Block
	var mode []FileSystemModifier
	var fs []vfs.FileSystem

	for _, arg := range args {
		switch arg := arg.(type) {
		case epi.Block:
			f = append(f, arg)
		case FileSystemModifier:
			mode = append(mode, arg)
		case vfs.FileSystem:
			fs = append(fs, arg)
		default:
			g.env.FailIfError(skip+1, fmt.Errorf("invalid filesystem modifier"))
		}
	}
	if mode == nil {
		mode = []FileSystemModifier{Normal}
	}
	return fs, mode, f
}
