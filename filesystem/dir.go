package filesystem

import (
	"github.com/mandelsoft/composer/epi"
	. "github.com/mandelsoft/composer/epi/contraints"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

// --- begin state ---

type DirectoryState interface {
	GetDir() (vfs.FileSystem, string)
}

// --- end state ---

// --- begin frame ---
type _dirState struct {
	fs  vfs.FileSystem
	dir string
}

func (s *_dirState) GetDir() (vfs.FileSystem, string) {
	return s.fs, s.dir
}

type dirFrame struct {
	epi.DefaultFrame[DirectoryState]
	name string
	mode vfs.FileMode
	_dirState
}

var (
	_ epi.Frame      = (*dirFrame)(nil)
	_ DirectoryState = (*dirFrame)(nil)
)

// --- end frame ---

// --- begin setup ---

func (f *dirFrame) Setup(elem string, s DirectoryState) (epi.Frame, error) {
	fs, dir := s.GetDir()
	f.dir = vfs.Join(fs, dir, f.name)
	f.fs = fs
	err := fs.MkdirAll(f.dir, f.mode)
	if err != nil {
		return nil, err
	}
	return f.DefaultFrame.Setup(elem, s)
}

// --- end setup ---

// --- begin directory ---

func (b *Group) Directory(name string, mode vfs.FileMode, f ...epi.Block) {
	if mode == 0 {
		mode = 0660
	}
	cs := Or(TopLevel, DirectEmbedding(StateTypeConstraint[DirectoryState]))
	epi.EvaluateWithState[DirectoryState](1, b.env, "Directory", "directory required", &dirFrame{name: name, mode: mode}, nil, cs, f)
}

// --- end directory ---
