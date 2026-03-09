package main_test

import (
	"github.com/mandelsoft/composer/filesystem"
	"github.com/mandelsoft/composer/utils"
	"github.com/mandelsoft/goutils/testutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/composer/testenv"
)

var _ = Describe("Test Environment", func() {
	var env *filesystem.Environment

	BeforeEach(func() {
		var err error

		env, err = testenv.New(filesystem.New, testenv.TestData())
		env.Directory("testdata/test", 0770, func() {
			env.File("additionalfile", 0660, func() {
				env.StringContent("some temporary data not stored in OS filesystem")
			})
		})
		env.Directory("other", 0770, func() {
			env.File("localfile", 0660, func() {
				env.StringContent("some temporary data")
			})
		})
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("some test", func() {

		tree := utils.MapNodetoASCII(utils.MapFSTree(env.GetFilesystem(), "/"))
		Expect(tree).To(testutils.StringEqualTrimmedWithContext(`
.
├── other
│   └── localfile
└── testdata
    └── test
        ├── additionalfile
        └── data
`))
	})
})
