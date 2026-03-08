package filesystem

import (
	"io"

	"github.com/mandelsoft/composer/common"
	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type fileState interface {
	GetFileName() string
}

type fileFrame struct {
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

func (f *fileFrame) Setup(s DirectoryState) (epi.Frame, error) {
	var err error
	f.dir = s
	fs, dir := s.GetDir()
	name := vfs.Join(fs, dir, f.name)
	f.file, err = fs.OpenFile(name, vfs.O_RDWR|vfs.O_CREATE, f.mode)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (f *fileFrame) Close() error {
	return f.file.Close()
}

func (b *Group) File(name string, mode vfs.FileMode, f ...epi.Block) {
	if mode == 0 {
		mode = 0660
	}
	epi.EvaluateWithState[DirectoryState](1, b.env, "directory required", &fileFrame{name: name, mode: mode}, f...)
}
