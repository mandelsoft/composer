package filesystem

import (
	"fmt"

	. "github.com/mandelsoft/composer/epi/constraints"
	"github.com/mandelsoft/vfs/pkg/osfs"

	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/vfs/pkg/composefs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type mountPoint struct {
	epi.DefaultFrame[FilesystemState]
	fs   vfs.FileSystem
	path string
	mode []FileSystemModifier
}

func (e *mountPoint) Setup(elem string, s FilesystemState) (epi.Frame, error) {
	fs := effFS(s.GetFilesystem())
	if cfs, ok := fs.(*composefs.ComposedFileSystem); !ok {
		return nil, fmt.Errorf("expected composable filesystem, got %T", fs)

	} else {
		mfs, err := applyMountModes(e.fs, e.mode...)
		if err != nil {
			return nil, err
		}
		err = cfs.MkdirAll(e.path, 0755)
		if err != nil {
			return nil, err
		}
		err = cfs.Mount(e.path, mfs)
		if err != nil {
			return nil, err
		}
		return e.DefaultFrame.Setup(elem, s)
	}
}

func (g *Group) Mount(path string, args ...any) {
	fs, mode, f := g.evalElementArgs(1, args...)
	if len(f) > 0 {
		g.env.FailIfError(1, fmt.Errorf("no nested block supported"))
	}
	if len(fs) == 0 {
		fs = []vfs.FileSystem{osfs.OsFs}
	}
	if len(fs) > 1 {
		g.env.FailIfError(1, fmt.Errorf("only one mount filesystem possible"))
	}
	cs := Or(StateFrame, DirectEmbedding(StateTypeConstraint[FilesystemState]))
	epi.EvaluateLeafWithState[FilesystemState](1, g.env, "Mount", "filesystem required", &mountPoint{mode: mode, path: path, fs: fs[0]}, nil, cs)
}
