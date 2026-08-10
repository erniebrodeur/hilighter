package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("resolveOptions", func() {
	It("uses shipped rules and Monokai for an empty customization layout", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "rules.yaml"), nil, 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("theme: monokai\n"), 0o644)).To(Succeed())

		rulesPath, themePath, err := resolveOptions(Options{}, dir)

		Expect(err).NotTo(HaveOccurred())
		Expect(rulesPath).To(BeEmpty())
		Expect(themePath).To(Equal("monokai"))
	})

	It("uses non-empty user rules and the configured relative theme", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "rules.yaml"), []byte("rules: []\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("theme: themes/custom.yaml\n"), 0o644)).To(Succeed())

		rulesPath, themePath, err := resolveOptions(Options{}, dir)

		Expect(err).NotTo(HaveOccurred())
		Expect(rulesPath).To(Equal(filepath.Join(dir, "rules.yaml")))
		Expect(themePath).To(Equal(filepath.Join(dir, "themes", "custom.yaml")))
	})

	It("lets expressions replace user rules and -t replace config", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "rules.yaml"), []byte("rules: []\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("theme: missing.yaml\n"), 0o644)).To(Succeed())

		rulesPath, themePath, err := resolveOptions(Options{Expressions: []string{"ERROR"}, ThemePath: "monokai"}, dir)

		Expect(err).NotTo(HaveOccurred())
		Expect(rulesPath).To(BeEmpty())
		Expect(themePath).To(Equal("monokai"))
	})

	It("resolves every shipped theme without treating its slug as a file", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "rules.yaml"), nil, 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("theme: dark-plus\n"), 0o644)).To(Succeed())

		_, themePath, err := resolveOptions(Options{}, dir)

		Expect(err).NotTo(HaveOccurred())
		Expect(themePath).To(Equal("dark-plus"))
	})
})

var _ = Describe("run", func() {
	It("reads stdin when no operands are provided", func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		err := run(nil, strings.NewReader("plain text\n"), &stdout, &stderr, GinkgoT().TempDir())

		Expect(err).NotTo(HaveOccurred())
		Expect(stdout.String()).To(Equal("plain text\n"))
		Expect(stderr.String()).To(BeEmpty())
	})

	It("validates expressions before reading input", func() {
		input := &countingReader{}

		err := run([]string{"-e", "("}, input, io.Discard, io.Discard, GinkgoT().TempDir())

		Expect(err).To(HaveOccurred())
		Expect(input.reads).To(BeZero())
	})

	It("prints help without reading input", func() {
		input := &countingReader{}
		var stdout bytes.Buffer

		err := run([]string{"--help"}, input, &stdout, io.Discard, GinkgoT().TempDir())

		Expect(err).NotTo(HaveOccurred())
		Expect(stdout.String()).To(Equal("Usage:\n  hilighter [-e <pattern> ...] [-t <theme>] [--] [file ...]\n  hilighter --help\n  hilighter --version\n"))
		Expect(input.reads).To(BeZero())
	})

	It("prints the version without reading input", func() {
		input := &countingReader{}
		var stdout bytes.Buffer

		err := run([]string{"--version"}, input, &stdout, io.Discard, GinkgoT().TempDir())

		Expect(err).NotTo(HaveOccurred())
		Expect(stdout.String()).To(Equal("hilighter-1.0.0\n"))
		Expect(input.reads).To(BeZero())
	})

	It("returns a reported error after an input failure", func() {
		var stderr bytes.Buffer
		missing := filepath.Join(GinkgoT().TempDir(), "missing")

		err := run([]string{missing}, strings.NewReader("ignored"), io.Discard, &stderr, GinkgoT().TempDir())

		Expect(err).To(HaveOccurred())
		Expect(SuppressError(err)).To(BeTrue())
		Expect(stderr.String()).To(ContainSubstring(missing))
	})

	It("silences only broken-pipe output errors", func() {
		err := run(nil, strings.NewReader("plain text\n"), errorWriter{err: syscall.EPIPE}, io.Discard, GinkgoT().TempDir())

		Expect(errors.Is(err, syscall.EPIPE)).To(BeTrue())
		Expect(SuppressError(err)).To(BeTrue())
	})

	It("does not silence ordinary output errors", func() {
		writeErr := errors.New("write failed")

		err := run(nil, strings.NewReader("plain text\n"), errorWriter{err: writeErr}, io.Discard, GinkgoT().TempDir())

		Expect(errors.Is(err, writeErr)).To(BeTrue())
		Expect(SuppressError(err)).To(BeFalse())
	})
})

type countingReader struct {
	reads int
}

func (reader *countingReader) Read([]byte) (int, error) {
	reader.reads++
	return 0, io.EOF
}

type errorWriter struct {
	err error
}

func (writer errorWriter) Write([]byte) (int, error) {
	return 0, writer.err
}
