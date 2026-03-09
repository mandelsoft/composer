package testenv

import (
	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/composer/filesystem"
	"github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/osfs"
)

// New creates a new test environment based on some group aggregation and additional
// options, the aggregation factory must provide an aggregation supporting
// all the functional areas required by the options, but at least the filesystem
// area.
func New[E epi.Environment](f epi.EnvironmentFactory[E], opts ...epi.Option) (E, error) {
	var _nil E
	tempfs, err := osfs.NewTempFileSystem()
	if err != nil {
		return _nil, err
	}

	return f(append([]epi.Option{filesystem.Filesystem(tempfs, true), epi.FailureHandler(ExpectFailureHandler)}, opts...)...)
}

func ExpectFailureHandler(skip int, env epi.EnvState, err error) {
	testutils.MustBeSuccessfulWithOffset(skip+1, err)
}
