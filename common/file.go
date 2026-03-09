package common

import (
	"io"
	"os"

	"github.com/mandelsoft/composer/epi"
)

type fileContent struct {
	epi.DefaultFrame[Writable]
	path string
}

func (e *fileContent) Setup(elem string, s Writable) (epi.Frame, error) {
	w := s.GetWriter()
	file, err := os.Open(e.path)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(w, file)
	if err != nil {
		return nil, err
	}
	return e.DefaultFrame.Setup(elem, s)
}

// OSFileContent copies the content of an operating system file
// into a writable context.
func (b *Group) OSFileContent(s string) {
	epi.EvaluateLeafWithState[Writable](1, b.env, "OSFileContent", "writable element required in outer scope", &fileContent{path: s}, nil, nil)
}
