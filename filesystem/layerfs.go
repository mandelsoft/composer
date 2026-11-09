package filesystem

import (
	"fmt"

	"github.com/mandelsoft/composer/epi"
	. "github.com/mandelsoft/composer/epi/constraints"
	"github.com/mandelsoft/vfs/pkg/layerfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type layerfsFrame struct {
	epi.DefaultFrame[FilesystemState]
	_fsState
	layers []vfs.FileSystem
	mode   []FileSystemModifier
}

var (
	_ epi.Frame         = (*layerfsFrame)(nil)
	_ epi.StateProvider = (*layerfsFrame)(nil)
	_ FilesystemState   = (*layerfsFrame)(nil)
	_ DirectoryState    = (*layerfsFrame)(nil)
)

func (e *layerfsFrame) GetState() any {
	return &e._fsState
}

func (e *layerfsFrame) Close() error {
	return nil
}

func (e *layerfsFrame) Setup(elem string, s FilesystemState) (epi.Frame, error) {
	var err error

	e.fs = s.GetFilesystem()
	for _, layer := range e.layers {
		e.fs = layerfs.New(layer, e.fs)
	}
	e.fs, err = applyMountModes(e.fs, e.mode...)
	if err != nil {
		return nil, err
	}
	return e.DefaultFrame.Setup(elem, s)
}

// FileSystemLayer accepts a filesystem to extend,
// FileSystemModifier(s) and nested blocks. Because it returns a vfs.FileSystem,
// it can also be uses as composition arg for itself.
// It provides an env state for a layered filesystem using the given filesystem on top
// of the next outer filesystem.
// If no nested blocks are given, the state is extended for subsequent method calls,
// otherwise it is only valid for the nested blocks.
func (g *Group) FileSystemLayer(args ...any) vfs.FileSystem {
	fs, mode, f := g.evalElementArgs(1, args...)
	if len(fs) == 0 {
		g.env.FailIfError(1, fmt.Errorf("filesystem for new layer required as argument"))
	}

	cs := Or(StateFrame, DirectEmbedding(StateTypeConstraint[FilesystemState]))
	frame := epi.EvaluateWithState[FilesystemState](1, g.env, "FilesystemLayer", "filesystem required", &layerfsFrame{layers: fs, mode: mode}, nil, cs, f)
	return frame.(*layerfsFrame).fs
}
