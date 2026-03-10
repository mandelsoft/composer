package filesystem

import (
	"slices"

	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/composefs"
	"github.com/mandelsoft/vfs/pkg/layerfs"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/readonlyfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

// --- begin filesystem option ---

type fsOpt struct {
	_fsState
}

func (o *fsOpt) ApplyTo(e epi.Environment) error {
	g := MapToGroup(e)
	if g == nil {
		return epi.ErrGroupNotSupported("filesystem")
	}
	e.AddState(&o._fsState)
	return nil
}

// --- end filesystem option ---

func Filesystem(fs vfs.FileSystem, cleanup ...bool) epi.Option {
	return &fsOpt{_fsState{saveFS(fs, general.Optional(cleanup...))}}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type mountOpt struct {
	fs    vfs.FileSystem
	path  string
	modes []MountMode
}

func (o *mountOpt) ApplyTo(e epi.Environment) error {
	var err error

	g := MapToGroup(e)
	if g == nil {
		return epi.ErrGroupNotSupported("filesystem")
	}

	mi := mountInfo{fs: o.fs}
	for _, m := range o.modes {
		mi, err = m.TransformFilesystem(mi)
		if err != nil {
			return err
		}
	}
	fs := saveFS(mi.fs, mi.cleanup)

	if o.path == "" {
		e.AddState(&_fsState{fs: fs})
		return nil
	}

	var c *composefs.ComposedFileSystem
	var t FilesystemState

	s, ok := epi.GetState[*_fsState](g.env)
	if !ok {
		t, ok = epi.GetState[FilesystemState](g.env)
		if ok {
			s = &_fsState{saveFS(t.GetFilesystem(), false)}
		} else {
			s = &_fsState{fs: osfs.OsFs}
		}
		e.AddState(s)
	}
	// Hmm, is it ok to modify an outer FS?
	c, ok = effFS(s.fs).(*composefs.ComposedFileSystem)
	if !ok {
		c = composefs.New(saveFS(s.fs, false))
		s = &_fsState{fs: c}
		e.AddState(s)
	}
	err = c.MkdirAll(o.path, vfs.ModePerm)
	if err != nil {
		return err
	}
	return c.Mount(o.path, fs)
}

type mountInfo struct {
	fs      vfs.FileSystem
	cleanup bool
}

func (i mountInfo) TransformedFilesystem(fs vfs.FileSystem, err error) (mountInfo, error) {
	i.fs = fs
	return i, err
}

type MountMode interface {
	TransformFilesystem(mountInfo) (mountInfo, error)
}

// Mount mounts a filesystem at a given path with optional [MountMode]s.
// The order of the modes is relevant, the modes are applied
// sequentially in the order they are given.
// If no filesystem is yet defined, the osfs is used as base filesystem.
func Mount(fs vfs.FileSystem, path string, modes ...MountMode) epi.Option {
	return &mountOpt{fs: fs, path: path, modes: slices.Clone(modes)}
}

type mountmode func(mountInfo) (mountInfo, error)

func (m mountmode) TransformFilesystem(i mountInfo) (mountInfo, error) {
	return m(i)
}

// Normal uses the filesystem as given without transformation.
var Normal = mountmode(normalMode)

// Readonly maps the filesystem to readonly mode.
var Readonly = mountmode(readonlyMode)

// Shadowed uses a modifiable memory layer on-top of the filesystem
// which is therefore not modified.
var Shadowed = mountmode(shadowedMode)

// Cleanup calls Cleanup of filesystem after use.
var Cleanup = mountmode(cleanupMode)

// Projected creates a projection of the filesystem to a given path.
func Projected(path string) MountMode {
	return mountmode(func(i mountInfo) (mountInfo, error) {
		return i.TransformedFilesystem(projectionfs.New(i.fs, path))
	})
}

func readonlyMode(i mountInfo) (mountInfo, error) {
	return i.TransformedFilesystem(readonlyfs.New(i.fs), nil)
}

func shadowedMode(i mountInfo) (mountInfo, error) {
	layer := memoryfs.New()
	lfs := layerfs.New(layer, i.fs)
	return i.TransformedFilesystem(lfs, nil)
}

func normalMode(i mountInfo) (mountInfo, error) {
	return i, nil
}

func cleanupMode(i mountInfo) (mountInfo, error) {
	i.cleanup = true
	return i, nil
}
