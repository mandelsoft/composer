package epi

import (
	"github.com/mandelsoft/goutils/errors"
)

const KIND_GROUP = "environment group"

func ErrGroupNotSupported(name string) error {
	return errors.ErrNotSupported(KIND_GROUP, name)
}
