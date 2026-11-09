package common_test

import (
	"bytes"
	"fmt"
	"io"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/composer/common"
)

// chunkWriter accepts at most size bytes per Write call.
type chunkWriter struct {
	size  int
	calls int
	buf   bytes.Buffer
	err   error
}

func (w *chunkWriter) Write(data []byte) (int, error) {
	w.calls++
	if w.err != nil {
		return 0, w.err
	}
	n := min(len(data), w.size)
	return w.buf.Write(data[:n])
}

// writable offers the chunkWriter as state for content elements.
type writable struct {
	w *chunkWriter
}

var _ common.Writable = (*writable)(nil)

func (s *writable) GetWriter() io.Writer {
	return s.w
}

var _ = Describe("Content", func() {
	var env *common.Environment
	var w *chunkWriter

	BeforeEach(func() {
		env = Must(common.New())
		w = &chunkWriter{size: 2}
	})

	It("requires a writable element", func() {
		ExpectError(env.Compose(func() {
			env.StringContent("test")
		})).To(MatchError(ContainSubstring("StringContent: writable element required in outer scope")))
	})

	It("writes string content completely on partial writes", func() {
		MustBeSuccessful(env.Compose(func() {
			env.With(&writable{w}, func() {
				env.StringContent("12345")
			})
		}))
		Expect(w.buf.String()).To(Equal("12345"))
		Expect(w.calls).To(Equal(3))
	})

	It("writes byte content completely on partial writes", func() {
		MustBeSuccessful(env.Compose(func() {
			env.With(&writable{w}, func() {
				env.ByteContent([]byte("12345"))
			})
		}))
		Expect(w.buf.String()).To(Equal("12345"))
		Expect(w.calls).To(Equal(3))
	})

	It("reports a write error", func() {
		w.err = fmt.Errorf("write failed")
		ExpectError(env.Compose(func() {
			env.With(&writable{w}, func() {
				env.StringContent("12345")
			})
		})).To(MatchError(ContainSubstring("StringContent: write failed")))
		Expect(w.calls).To(Equal(1))
	})
})
