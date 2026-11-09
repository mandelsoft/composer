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
	modes []FileSystemModifier
}

func (o *mountOpt) ApplyTo(e epi.Environment) error {
	var err error

	g := MapToGroup(e)
	if g == nil {
		return epi.ErrGroupNotSupported("filesystem")
	}

	fs, err := applyMountModes(o.fs, o.modes...)
	if err != nil {
		return err
	}

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

func applyMountModes(fs vfs.FileSystem, modes ...FileSystemModifier) (vfs.FileSystem, error) {
	var err error

	for _, m := range modes {
		fs, err = m.TransformFilesystem(fs)
		if err != nil {
			return nil, err
		}
	}
	return fs, nil
}

type FileSystemModifier interface {
	TransformFilesystem(system vfs.FileSystem) (vfs.FileSystem, error)
}

// Mount mounts a filesystem at a given path with optional [MountMode]s.
// The order of the modes is relevant, the modes are applied
// sequentially in the order they are given.
// If no filesystem is yet defined, the osfs is used as base filesystem.
func Mount(fs vfs.FileSystem, path string, modes ...FileSystemModifier) epi.Option {
	return &mountOpt{fs: fs, path: path, modes: slices.Clone(modes)}
}

type mountmode func(system vfs.FileSystem) (vfs.FileSystem, error)

func (m mountmode) TransformFilesystem(fs vfs.FileSystem) (vfs.FileSystem, error) {
	return m(fs)
}

// Normal uses the filesystem as given without transformation.
var Normal = mountmode(normalMode)

// Readonly maps the filesystem to readonly mode.
var Readonly = mountmode(readonlyMode)

// Shadowed uses a modifiable memory layer on-top of the filesystem
// which is therefore not modified.
var Shadowed = mountmode(shadowedMode)

// Cleanup calls Cleanup of filesystem after use.
var Cleanup cleanupMode

// NoCleanup disables Cleanup of filesystem after use.
var NoCleanup = mountmode(nocleanupMode)

// Projected creates a projection of the filesystem to a given path.
func Projected(path string) FileSystemModifier {
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

func nocleanupMode(fs vfs.FileSystem) (vfs.FileSystem, error) {
	return saveFS(fs, false), nil
}

type cleanupMode struct {
}

var _ FileSystemModifier = cleanupMode{}

func (_ cleanupMode) TransformFilesystem(fs vfs.FileSystem) (vfs.FileSystem, error) {
	return fs, nil
}

func requireCleanup(modes ...FileSystemModifier) bool {
	for _, m := range modes {
		if m == Cleanup {
			return true
		}
	}
	return false
}
