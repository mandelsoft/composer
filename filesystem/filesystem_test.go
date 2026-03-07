package filesystem_test

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/composer/filesystem"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filesystem Test Environment", func() {
	Context("standard env", func() {
		var env *filesystem.Environment
		var fs vfs.FileSystem

		BeforeEach(func() {
			fs = memoryfs.New()
			env = Must(filesystem.New(filesystem.Filesystem(fs)))
		})

		It("catch error", func() {
			info := epi.CallerInfo(0, 2)
			fmt.Printf("%s\n", info)
			Expect(env.Compose(func() {
				env.FailIfError(fmt.Errorf("this is an error"))
			})).To(MatchError(ContainSubstring(info[strings.LastIndex(info, ""):] + ": this is an error")))
		})

		It("creates directory", func() {
			MustBeSuccessful(env.Compose(func() {
				env.Directory("test", 0770)
			}))
			Expect(vfs.DirExists(fs, "test")).To(BeTrue())
			fi := Must(fs.Stat("/test"))
			Expect(fi.Mode() & 0777).To(Equal(os.FileMode(0770)))
		})

		It("creates file", func() {
			MustBeSuccessful(env.Compose(func() {
				env.Directory("test", 0770, func() {
					env.File("file", 0600, func() {
						env.StringContent("this is a test file")
					})
				})
			}))
			Expect(vfs.FileExists(fs, "test/file")).To(BeTrue())
			fi := Must(fs.Stat("/test/file"))
			Expect(fi.Mode() & 0777).To(Equal(os.FileMode(0600)))

			Expect(vfs.ReadFile(fs, "test/file")).To(Equal([]byte("this is a test file")))
		})
	})

	It("tests recover", func() {
		defer func() {
			r := recover()
			fmt.Println("Recovered panic:", r)
			fmt.Println(string(debug.Stack()))
		}()
		p(4)
	})
})

func cp(r any) {
	panic(r)
}

func p(i int) {
	defer func() {
		r := recover()
		if r != nil {
			cp(r)
		}
	}()
	i--
	if i <= 0 {
		panic("test")
	}
	p(i)
}
