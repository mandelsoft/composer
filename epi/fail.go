package epi

import (
	"github.com/mandelsoft/goutils/exception"
)

type FailureHandler func(skip int, env EnvState, err error)

func (f FailureHandler) ApplyTo(e Environment) error {
	return nil
}

func FailWithException(skip int, env EnvState, err error) {
	if err == nil {
		return
	}
	exception.Throw(err)
}

func FailWithExceptionLocation(skip int, env EnvState, err error) {
	if err == nil {
		return
	}
	info := CallerInfo(skip + 1)
	if info != "" {
		exception.Throwf(err, "%s", info)
	} else {
		exception.Throw(err)
	}
}
