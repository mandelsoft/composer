package filesystem_test

import (
	"os"

	"github.com/mandelsoft/composer/filesystem"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filesystem Mount Test Environment", func() {
	Context("standard env", func() {
		var env *filesystem.Environment
		var fs vfs.FileSystem

		BeforeEach(func() {
			fs = memoryfs.New()
			env = Must(filesystem.New(filesystem.Filesystem(fs)))
		})

		AfterEach(func() {
			env.Cleanup()
		})

		It("reads file", func() {
			var fs vfs.FileSystem

			MustBeSuccessful(env.Compose(func() {
				fs = env.ComposableFileSystem(func() {
					env.Mount("/testdata", filesystem.Projected("testdata"))
				})
			}))
			Expect(vfs.FileExists(fs, "/testdata/dir/testfile")).To(BeTrue())
			fi := Must(fs.Stat("testdata/dir/testfile"))
			Expect(fi.Mode() & 0777).To(Equal(os.FileMode(0664)))

			Expect(vfs.ReadFile(fs, "testdata/dir/testfile")).To(Equal([]byte("this is a test file")))
		})

		It("reads readonly file", func() {
			var fs vfs.FileSystem

			MustBeSuccessful(env.Compose(func() {
				fs = env.ComposableFileSystem(func() {
					env.Mount("/testdata", filesystem.Projected("testdata"), filesystem.Readonly)
				})
			}))
			Expect(vfs.FileExists(fs, "/testdata/dir/testfile")).To(BeTrue())
			fi := Must(fs.Stat("testdata/dir/testfile"))
			Expect(fi.Mode() & 0777).To(Equal(os.FileMode(0664)))

			Expect(vfs.ReadFile(fs, "testdata/dir/testfile")).To(Equal([]byte("this is a test file")))

			err := vfs.WriteFile(fs, "/testdata/notwritable", []byte("test"), 0o600)
			ExpectError(err).To(Equal(vfs.ErrReadOnly))
		})

		It("write shadowed file", func() {
			var fs vfs.FileSystem

			MustBeSuccessful(env.Compose(func() {
				fs = env.ComposableFileSystem(func() {
					env.Mount("/testdata", filesystem.Projected("testdata"), filesystem.Readonly, filesystem.Shadowed)
				})
			}))
			Expect(vfs.FileExists(fs, "/testdata/dir/testfile")).To(BeTrue())
			fi := Must(fs.Stat("testdata/dir/testfile"))
			Expect(fi.Mode() & 0777).To(Equal(os.FileMode(0664)))

			Expect(vfs.ReadFile(fs, "testdata/dir/testfile")).To(Equal([]byte("this is a test file")))

			MustBeSuccessful(vfs.WriteFile(fs, "/testdata/writable", []byte("test"), 0o600))
			ExpectError(os.Stat("testdata/writable")).To(HaveOccurred())
			ExpectError(os.Stat("testdata/dir/testfile")).NotTo(HaveOccurred())

			Expect(vfs.ReadFile(fs, "testdata/writable")).To(Equal([]byte("test")))
		})
	})
})
