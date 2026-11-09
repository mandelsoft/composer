package filesystem

import (
	"io"

	. "github.com/mandelsoft/composer/epi/constraints"

	"github.com/mandelsoft/composer/common"
	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type fileState interface {
	GetFileName() string
}

type fileFrame struct {
	epi.DefaultFrame[DirectoryState]
	name  string
	mode  vfs.FileMode
	flags int
	dir   DirectoryState
	file  vfs.File
}

var (
	_ epi.Frame       = (*fileFrame)(nil)
	_ common.Writable = (*fileFrame)(nil)
)

func (f *fileFrame) GetFileName() string {
	return f.file.Name()
}

func (f *fileFrame) GetWriter() io.Writer {
	return f.file
}

func (f *fileFrame) Setup(elem string, s DirectoryState) (epi.Frame, error) {
	var err error
	f.dir = s
	fs, dir := s.GetDir()
	name := vfs.Join(fs, dir, f.name)
	f.file, err = fs.OpenFile(name, f.flags, f.mode)
	if err != nil {
		return nil, err
	}
	return f.DefaultFrame.Setup(elem, s)
}

func (f *fileFrame) Close() error {
	return f.file.Close()
}

// File creates a file with the given mode, or replaces the content
// of an existing one. The content is defined by the nested blocks.
func (g *Group) File(name string, mode vfs.FileMode, f ...epi.Block) {
	g.file("File", name, mode, vfs.O_RDWR|vfs.O_CREATE|vfs.O_TRUNC, f)
}

// AppendFile creates a file with the given mode, or appends to the content
// of an existing one. The appended content is defined by the nested blocks.
func (g *Group) AppendFile(name string, mode vfs.FileMode, f ...epi.Block) {
	g.file("AppendFile", name, mode, vfs.O_RDWR|vfs.O_CREATE|vfs.O_APPEND, f)
}

func (g *Group) file(elem, name string, mode vfs.FileMode, flags int, f []epi.Block) {
	if mode == 0 {
		mode = 0660
	}
	cs := Or(StateFrame, Not(FrameTypeConstraint[*fileFrame]))
	epi.EvaluateWithState[DirectoryState](2, g.env, elem, "directory required", &fileFrame{name: name, mode: mode, flags: flags}, nil, cs, f)
}
