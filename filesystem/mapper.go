package filesystem

import (
	"github.com/mandelsoft/composer/epi"
)

// --- begin mapper ---

type GroupMapper interface {
	maptoFilesystemGroup() *Group
}

func MapToGroup(e epi.Environment) *Group {
	if m, ok := e.(GroupMapper); ok {
		return m.maptoFilesystemGroup()
	}
	return nil
}

// --- end mapper ---
