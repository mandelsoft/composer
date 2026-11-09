package testenv

import (
	"strings"

	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/composer/filesystem"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/pkgutils"
	"github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/osfs"
)

type tdOpt struct {
	path       string
	source     string
	modifiable bool
}

func testData(modifiable bool, paths ...string) epi.Option {
	path := "/testdata"
	source := "testdata"

	switch len(paths) {
	case 0:
	case 1:
		source = paths[0]
	case 2:
		source = paths[0]
		path = paths[1]
	default:
		panic("invalid number of arguments")
	}

	var modes []filesystem.FileSystemModifier
	if source != "" {
		modes = append(modes, filesystem.Projected(source))
	}
	// test data from the OS filesystem is never modified:
	// it is either readonly or shadowed by a memory layer.
	if modifiable {
		modes = append(modes, filesystem.Shadowed)
	} else {
		modes = append(modes, filesystem.Readonly)
	}
	return filesystem.Mount(osfs.OsFs, path, modes...)
}

// TestData mounts the package-local folder testdata (or the given source)
// readonly at /testdata (or the given path). Write attempts fail with
// vfs.ErrReadOnly.
func TestData(paths ...string) epi.Option {
	return testData(false, paths...)
}

// ModifiableTestData mounts the package-local folder testdata (or the given
// source) at /testdata (or the given path) with a modifiable memory layer
// on-top. Changes are never written through to the source folder.
func ModifiableTestData(paths ...string) epi.Option {
	return testData(true, paths...)
}

func projectTestData(modifiable bool, source string, dest ...string) epi.Option {
	pathToRoot, err := testutils.GetRelativePathToProjectRoot()
	if err != nil {
		panic(err)
	}
	pathToTestdata := filepath.Join(pathToRoot, source)

	return testData(modifiable, pathToTestdata, general.OptionalNonZeroDefaulted("/testdata", dest...))
}

// ProjectTestData mounts a folder of the project readonly, see [TestData].
func ProjectTestData(source string, dest ...string) epi.Option {
	return projectTestData(false, source, dest...)
}

// ModifiableProjectTestData mounts a folder of the project with a modifiable
// memory layer on-top, see [ModifiableTestData].
func ModifiableProjectTestData(source string, dest ...string) epi.Option {
	return projectTestData(true, source, dest...)
}

func projectTestDataForCaller(modifiable bool, source string, dest ...string) epi.Option {
	packagePath, err := pkgutils.GetPackageName(2)
	if err != nil {
		panic(err)
	}

	moduleName, err := testutils.GetModuleName()
	if err != nil {
		panic(err)
	}
	path, ok := strings.CutPrefix(packagePath, moduleName+"/")
	if !ok {
		panic("unable to find package name")
	}

	return projectTestData(modifiable, filepath.Join(path, source), dest...)
}

func ProjectTestDataForCaller(source string, dest ...string) epi.Option {
	return projectTestDataForCaller(false, source, dest...)
}

func ModifiableProjectTestDataForCaller(source string, dest ...string) epi.Option {
	return projectTestDataForCaller(true, source, dest...)
}
