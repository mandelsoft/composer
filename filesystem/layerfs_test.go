package filesystem_test

import (
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/composer/filesystem"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

var _ = Describe("Layering Test Environment", func() {
	Context("standard env", func() {
		var env *filesystem.Environment

		BeforeEach(func() {
			env = Must(filesystem.New())
		})

		AfterEach(func() {
			env.Cleanup()
		})

		It("reads file from composed base", func() {
			var fs vfs.FileSystem

			MustBeSuccessful(env.Compose(func() {
				fs = env.FileSystem(osfs.OsFs, filesystem.Projected("testdata"), filesystem.Readonly, func() {
					Expect(env.FileExists("/dir/testfile")).To(BeTrue())
				}, func() { // working on the readonly base filesystem

				})
			}))
			Expect(vfs.FileExists(fs, "/dir/testfile")).To(BeTrue())
			fi := Must(fs.Stat("/dir/testfile"))
			Expect(fi.Mode() & 0777).To(Equal(os.FileMode(0664)))

			Expect(vfs.ReadFile(fs, "/dir/testfile")).To(Equal([]byte("this is a test file")))
		})

		It("maintains a layer", func() {
			var fs vfs.FileSystem

			MustBeSuccessful(env.Compose(func() {
				env.FileSystem(osfs.OsFs, filesystem.Projected("testdata"), filesystem.Readonly, func() {
					Expect(env.FileExists("dir/testfile")).To(BeTrue())
					fs = env.FileSystemLayer(memoryfs.New(), func() { // working on the layer directly
						env.File("test", 0o600, func() {
							env.StringContent("test data")
						})
					}, func() { // working on the layered filesystem (assigned to fs)
						env.Directory("dir", 0o700, func() {
							env.File("test", 0o600, func() {
								env.StringContent("other data")
							})
						})
					})
				}, func() { // working on the readonly base filesystem
					err := env.Compose(func() {
						env.Directory("doesnotwork", 0o700)
					})
					Expect(err).To(MatchError(vfs.ErrReadOnly))
				})
			}))
			Expect(vfs.FileExists(fs, "/dir/testfile")).To(BeTrue())
			fi := Must(fs.Stat("/dir/testfile"))
			Expect(fi.Mode() & 0777).To(Equal(os.FileMode(0664)))
			Expect(vfs.ReadFile(fs, "/dir/testfile")).To(Equal([]byte("this is a test file")))

			Expect(vfs.FileExists(fs, "/dir/test")).To(BeTrue())
			fi = Must(fs.Stat("/dir/test"))
			Expect(fi.Mode() & 0777).To(Equal(os.FileMode(0600)))
			Expect(vfs.ReadFile(fs, "/dir/test")).To(Equal([]byte("other data")))
		})

		// two additional layers are added sequenially in one block
		It("maintains layers as sequence", func() {
			var outer, inner vfs.FileSystem

			MustBeSuccessful(env.Compose(func() {
				env.FileSystem(osfs.OsFs, filesystem.Projected("testdata"), filesystem.Readonly, func() {
					Expect(env.FileExists("dir/testfile")).To(BeTrue())

					// switch context to new layer on top.
					inner = env.FileSystemLayer(memoryfs.New())
					env.Directory("dir", 0o700, func() {
						env.File("test", 0o600, func() {
							env.StringContent("test data")
						})
					})

					// switch context to new layer on top.
					outer = env.FileSystemLayer(memoryfs.New())
					env.File("dir/test", 0o600, func() {
						env.StringContent("other data")
					})
				})
			}))
			Expect(vfs.FileExists(outer, "/dir/testfile")).To(BeTrue())
			fi := Must(outer.Stat("/dir/testfile"))
			Expect(fi.Mode() & 0777).To(Equal(os.FileMode(0664)))
			Expect(vfs.ReadFile(outer, "/dir/testfile")).To(Equal([]byte("this is a test file")))

			Expect(vfs.FileExists(inner, "/dir/test")).To(BeTrue())
			fi = Must(inner.Stat("/dir/test"))
			Expect(fi.Mode() & 0777).To(Equal(os.FileMode(0600)))
			Expect(vfs.ReadFile(inner, "/dir/test")).To(Equal([]byte("test data")))

			Expect(vfs.FileExists(outer, "/dir/test")).To(BeTrue())
			fi = Must(outer.Stat("/dir/test"))
			Expect(fi.Mode() & 0777).To(Equal(os.FileMode(0600)))
			Expect(vfs.ReadFile(outer, "/dir/test")).To(Equal([]byte("other data")))
		})

	})
})
