package runner_test

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/erniebrodeur/hilighter/pkg/runner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunCommand", func() {
	It("forwards stdout and stderr from the executed command", func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		err := runner.RunCommand("printf 'out'; printf 'err' >&2", &stdout, &stderr, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(stdout.String()).To(Equal("out"))
		Expect(stderr.String()).To(Equal("err"))
	})

	It("returns an error when the command exits non-zero", func() {
		err := runner.RunCommand("exit 7", &bytes.Buffer{}, &bytes.Buffer{}, nil)

		Expect(err).To(HaveOccurred())
	})

	It("eventually streams command output through the highlighting engine", func() {
		rulesPath := filepath.Join(GinkgoT().TempDir(), "rules.yaml")
		Expect(os.WriteFile(rulesPath, []byte("rules:\n  - name: error\n    pattern: 'ERROR'\n    style: error\n"), 0o644)).To(Succeed())
		highlighter, err := runner.NewHighlighter(rulesPath, "", "")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(highlighter.Close)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		err = runner.RunCommand("printf 'ERROR\\n'; printf 'plain\\n' >&2", &stdout, &stderr, highlighter)

		Expect(err).NotTo(HaveOccurred())
		Expect(stdout.String()).To(ContainSubstring("ERROR"))
		Expect(stdout.String()).To(ContainSubstring("\x1b["))
		Expect(stderr.String()).To(Equal("plain\n"))
	})
})
