package filesystem

import (
	"slices"

	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/vfs/pkg/composefs"
	"github.com/mandelsoft/vfs/pkg/layerfs"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/readonlyfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type fsOpt struct {
	fs vfs.FileSystem
}

func (o *fsOpt) ApplyTo(e epi.Environment) error {
	g := MapToGroup(e)
	if g == nil {
		return epi.ErrGroupNotSupported("filesystem")
	}
	e.AddState(&_fsState{o.fs})
	return nil
}

func Filesystem(fs vfs.FileSystem) epi.Option {
	return &fsOpt{fs}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type mountOpt struct {
	fs    vfs.FileSystem
	path  string
	modes []MountMode
}

func (o *mountOpt) ApplyTo(e epi.Environment) error {
	g := MapToGroup(e)
	if g == nil {
		return epi.ErrGroupNotSupported("filesystem")
	}

	if o.path == "" {
		e.AddState(&_fsState{o.fs})
		return nil
	}

	var c *composefs.ComposedFileSystem
	var t FilesystemState
	var err error

	fs := o.fs
	for _, m := range o.modes {
		fs, err = m.TransformFileSystem(fs)
		if err != nil {
			return err
		}
	}

	s, ok := epi.GetState[*_fsState](g.env)
	if !ok {
		t, ok = epi.GetState[FilesystemState](g.env)
		if ok {
			s = &_fsState{t.GetFilesystem()}
		} else {
			c = composefs.New(osfs.OsFs)
			s = &_fsState{c}
		}
		e.AddState(s)
	} else {
		if c, ok = s.fs.(*composefs.ComposedFileSystem); !ok {
			c = composefs.New(s.fs)
			s = &_fsState{c}
			e.AddState(s)
		}
	}
	err = c.MkdirAll(o.path, vfs.ModePerm)
	if err != nil {
		return err
	}
	return c.Mount(o.path, fs)
}

type MountMode interface {
	TransformFileSystem(fs vfs.FileSystem) (vfs.FileSystem, error)
}

// Mount mounts a filesystem at a given path with optional [MountMode]s.
// The order of the modes is relevant, the modes are applied
// sequentially in the order they are given.
// If no filesystem is yet defined, the osfs is used as base filesystem.
func Mount(fs vfs.FileSystem, path string, modes ...MountMode) epi.Option {
	return &mountOpt{fs: fs, path: path, modes: slices.Clone(modes)}
}

type mountmode func(fs vfs.FileSystem) (vfs.FileSystem, error)

func (m mountmode) TransformFileSystem(fs vfs.FileSystem) (vfs.FileSystem, error) {
	return m(fs)
}

// Normal uses the filesystem as given without transformation.
var Normal = mountmode(normalMode)

// Readonly maps the filesystem to readonly mode.
var Readonly = mountmode(readonlyMode)

// Shadowed uses a modifiable memory layer on-top of the filesystem
// which is therefore not modified.
var Shadowed = mountmode(shadowedMode)

// Projected creates a projection of the filesystem to a given path.
func Projected(path string) MountMode {
	return mountmode(func(fs vfs.FileSystem) (vfs.FileSystem, error) {
		return projectionfs.New(fs, path)
	})
}

func readonlyMode(fs vfs.FileSystem) (vfs.FileSystem, error) {
	return readonlyfs.New(fs), nil
}

func shadowedMode(fs vfs.FileSystem) (vfs.FileSystem, error) {
	layer := memoryfs.New()
	lfs := layerfs.New(layer, fs)

	return lfs, nil
}

func normalMode(fs vfs.FileSystem) (vfs.FileSystem, error) {
	return fs, nil
}
