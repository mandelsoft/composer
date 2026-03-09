package testenv_test

import (
	"github.com/mandelsoft/composer/filesystem"
	"github.com/mandelsoft/composer/testenv"
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
})
