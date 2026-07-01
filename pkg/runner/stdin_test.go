package runner_test

import (
	"bytes"

	"github.com/erniebrodeur/hilighter/pkg/runner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunStdin", func() {
	It("copies the input stream to the output stream unchanged", func() {
		input := bytes.NewBufferString("hello\nworld\n")
		var output bytes.Buffer

		err := runner.RunStdin(input, &output)

		Expect(err).NotTo(HaveOccurred())
		Expect(output.String()).To(Equal("hello\nworld\n"))
	})
})
