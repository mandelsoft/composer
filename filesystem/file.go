package filesystem

import (
	"io"

	"github.com/mandelsoft/composer/common"
	"github.com/mandelsoft/composer/epi"
	. "github.com/mandelsoft/composer/epi/contraints"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type fileState interface {
	GetFileName() string
}

type fileFrame struct {
	epi.DefaultFrame[DirectoryState]
	name string
	mode vfs.FileMode
	dir  DirectoryState
	file vfs.File
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
	f.file, err = fs.OpenFile(name, vfs.O_RDWR|vfs.O_CREATE, f.mode)
	if err != nil {
		return nil, err
	}
	return f.DefaultFrame.Setup(elem, s)
}

func (f *fileFrame) Close() error {
	return f.file.Close()
}

func (b *Group) File(name string, mode vfs.FileMode, f ...epi.Block) {
	if mode == 0 {
		mode = 0660
	}
	cs := Or(StateFrame, Not(FrameTypeConstraint[*fileFrame]))
	epi.EvaluateWithState[DirectoryState](1, b.env, "File", "directory required", &fileFrame{name: name, mode: mode}, nil, cs, f)
}
