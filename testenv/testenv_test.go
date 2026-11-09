package testenv_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/composer/filesystem"
	"github.com/mandelsoft/composer/testenv"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

var _ = Describe("Testenv Test Environment", func() {
	Context("env creation", func() {
		It("creates env with a temp fs with mounted test data", func() {
			// create a filesystem environment for a temporary folder
			// with a gomega failure handling and a mounted test data from testdata folder.
			// --- begin creation ---
			env := Must(testenv.New(filesystem.New, testenv.TestData()))
			// --- end creation ---

			defer env.Cleanup()

			env.Directory("/testdata/test", 0770, func() {
				Expect(env.ReadFile("data")).To(Equal([]byte("this is a test file")))
			})
			Expect("").To(Equal(""))
		})
	})

	Context("test data", func() {
		It("mounts test data readonly", func() {
			// use the exception based failure handling to get the error from Compose.
			env := Must(testenv.New(filesystem.New, epi.FailureHandler(epi.FailWithExceptionLocation), testenv.TestData()))
			defer env.Cleanup()

			err := env.Compose(func() {
				env.File("/testdata/test/newfile", 0o600, func() {
					env.StringContent("not possible")
				})
			})
			Expect(err).To(MatchError(vfs.ErrReadOnly))
			Expect(vfs.FileExists(osfs.OsFs, "testdata/test/newfile")).To(BeFalse())
		})

		It("mounts modifiable test data without write-through", func() {
			env := Must(testenv.New(filesystem.New, testenv.ModifiableTestData()))
			defer env.Cleanup()

			env.File("/testdata/test/newfile", 0o600, func() {
				env.StringContent("only in memory")
			})
			env.File("/testdata/test/data", 0o600, func() {
				env.StringContent("modified")
			})
			Expect(env.ReadFile("/testdata/test/newfile")).To(Equal([]byte("only in memory")))
			Expect(env.ReadFile("/testdata/test/data")).To(Equal([]byte("modified")))

			Expect(vfs.FileExists(osfs.OsFs, "testdata/test/newfile")).To(BeFalse())
			Expect(vfs.ReadFile(osfs.OsFs, "testdata/test/data")).To(Equal([]byte("this is a test file")))
		})

		It("appends to modifiable test data without write-through", func() {
			env := Must(testenv.New(filesystem.New, testenv.ModifiableTestData()))
			defer env.Cleanup()

			env.AppendFile("/testdata/test/data", 0o600, func() {
				env.StringContent(", appended")
			})
			Expect(env.ReadFile("/testdata/test/data")).To(Equal([]byte("this is a test file, appended")))
			Expect(vfs.ReadFile(osfs.OsFs, "testdata/test/data")).To(Equal([]byte("this is a test file")))
		})
	})
})
