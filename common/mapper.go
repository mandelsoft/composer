package common

import (
	"github.com/mandelsoft/composer/epi"
)

type GroupMapper interface {
	maptoCommonGroup() *Group
}

func MapToGroup(e epi.Environment) *Group {
	if m, ok := e.(GroupMapper); ok {
		return m.maptoCommonGroup()
	}
	return nil
}
