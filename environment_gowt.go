//go:build ignore

package composer

import (
	"github.com/mandelsoft/composer/epi"
)

// Environment[Base] wraps any base environment type and adds ConfigSetGroup.
// Base must satisfy epi.Environment.
//
// Instantiate with:
//
//	// gowt: type MyEnv wrapping github.com/mandelsoft/composer.Environment[*mypkg.MyBase via mypkg.New]
type Environment[Base epi.Environment] struct {
	Base
	Group
}

// New creates a new Environment, applying any supplied options.
// NEW is replaced by the "via" constructor from the instantiation directive.
func New[Base epi.Environment](opts ...epi.Option) (*Environment[Base], error) {
	e := epi.NewEnvState(opts...)
	base, err := NEW(e)
	if err != nil {
		return nil, err
	}
	env := &Environment[Base]{
		Base:  base,
		Group: *PACKAGE.NewGroup(e),
	}
	return epi.WithOptionsApplied(env, opts...)
}
