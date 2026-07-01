package runner_test

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/erniebrodeur/hilighter/pkg/runner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunStdin", func() {
	It("copies the input stream to the output stream unchanged", func() {
		input := bytes.NewBufferString("hello\nworld\n")
		var output bytes.Buffer

		err := runner.RunStdin(input, &output, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(output.String()).To(Equal("hello\nworld\n"))
	})

	It("eventually applies configured highlighting while preserving the original text", func() {
		rulesPath := filepath.Join(GinkgoT().TempDir(), "rules.yaml")
		Expect(os.WriteFile(rulesPath, []byte("rules:\n  - name: error\n    pattern: 'ERROR'\n    style: error\n"), 0o644)).To(Succeed())
		highlighter, err := runner.NewHighlighter(rulesPath, "")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(highlighter.Close)

		input := bytes.NewBufferString("hello\nERROR\n")
		var output bytes.Buffer

		err = runner.RunStdin(input, &output, highlighter)

		Expect(err).NotTo(HaveOccurred())
		Expect(output.String()).To(ContainSubstring("hello\n"))
		Expect(output.String()).To(ContainSubstring("ERROR"))
		Expect(output.String()).To(ContainSubstring("\x1b["))
	})
})
