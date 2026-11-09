package filesystem

import (
	"github.com/mandelsoft/composer/epi"
	. "github.com/mandelsoft/composer/epi/constraints"
	"github.com/mandelsoft/vfs/pkg/composefs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type compfsFrame struct {
	epi.DefaultFrame[FilesystemState]
	_fsState
}

var (
	_ epi.Frame       = (*compfsFrame)(nil)
	_ FilesystemState = (*compfsFrame)(nil)
	_ DirectoryState  = (*compfsFrame)(nil)
)

func (f *compfsFrame) Close() error {
	return nil
}

func (f *compfsFrame) Setup(elem string, s FilesystemState) (epi.Frame, error) {
	fs := s.GetFilesystem()
	if _, ok := effFS(fs).(*composefs.ComposedFileSystem); !ok {
		fs = composefs.New(fs, "/tmp")
	}
	f.fs = fs
	return f.DefaultFrame.Setup(elem, s)
}

func (g *Group) ComposableFileSystem(f ...epi.Block) vfs.FileSystem {
	cs := Or(StateFrame, DirectEmbedding(StateTypeConstraint[FilesystemState]))
	frame := epi.EvaluateWithState[FilesystemState](1, g.env, "ComposableFileSystem", "filesystem required", &compfsFrame{}, nil, cs, f)
	return frame.(*compfsFrame).fs
}
