package runner_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/erniebrodeur/hilighter/pkg/runner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunStdin", func() {
	It("copies text unchanged without a highlighter", func() {
		var output bytes.Buffer

		err := runner.RunStdin(bytes.NewBufferString("hello\nworld\n"), &output, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(output.String()).To(Equal("hello\nworld\n"))
	})

	It("applies configured highlighting", func() {
		rulesPath := filepath.Join(GinkgoT().TempDir(), "rules.yaml")
		Expect(os.WriteFile(rulesPath, []byte("rules:\n  - name: error\n    pattern: 'ERROR'\n    style: error\n"), 0o644)).To(Succeed())
		highlighter, err := runner.NewHighlighter(rulesPath, "")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(highlighter.Close)
		var output bytes.Buffer

		err = runner.RunStdin(bytes.NewBufferString("ERROR\n"), &output, highlighter)

		Expect(err).NotTo(HaveOccurred())
		Expect(output.String()).To(ContainSubstring("\x1b["))
	})

	It("preserves existing ANSI sequences while adding highlights", func() {
		rulesPath := filepath.Join(GinkgoT().TempDir(), "rules.yaml")
		Expect(os.WriteFile(rulesPath, []byte("rules:\n  - name: error\n    pattern: 'ERROR'\n    style: error\n"), 0o644)).To(Succeed())
		highlighter, err := runner.NewHighlighter(rulesPath, "")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(highlighter.Close)
		var output bytes.Buffer
		input := "\x1b[32malready green\x1b[0m ERROR\n"

		err = runner.RunStdin(bytes.NewBufferString(input), &output, highlighter)

		Expect(err).NotTo(HaveOccurred())
		Expect(output.String()).To(ContainSubstring("\x1b[32malready green\x1b[0m"))
		Expect(output.String()).To(ContainSubstring(";41mERROR\x1b[0m"))
	})

	It("preserves ANSI sequences through a streaming pipeline", func() {
		rulesPath := filepath.Join(GinkgoT().TempDir(), "rules.yaml")
		Expect(os.WriteFile(rulesPath, []byte("rules:\n  - name: error\n    pattern: 'ERROR'\n    style: error\n"), 0o644)).To(Succeed())
		highlighter, err := runner.NewHighlighter(rulesPath, "")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(highlighter.Close)
		pipeReader, pipeWriter := io.Pipe()
		var output bytes.Buffer
		input := "\x1b[36mcyan\x1b[0m ERROR\n"
		writeDone := make(chan error, 1)
		go func() {
			_, err := io.WriteString(pipeWriter, input)
			if closeErr := pipeWriter.Close(); err == nil {
				err = closeErr
			}
			writeDone <- err
		}()

		err = runner.RunStdin(pipeReader, &output, highlighter)

		Expect(err).NotTo(HaveOccurred())
		Expect(<-writeDone).NotTo(HaveOccurred())
		Expect(output.String()).To(ContainSubstring("\x1b[36mcyan\x1b[0m"))
		Expect(output.String()).To(ContainSubstring(";41mERROR\x1b[0m"))
		Expect(output.String()).NotTo(Equal(input))
	})
})

var _ = Describe("RunInputs", func() {
	It("reads stdin when no file operands are provided", func() {
		var output bytes.Buffer

		err := runner.RunInputs(bytes.NewBufferString("stdin\n"), &output, io.Discard, nil, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(output.String()).To(Equal("stdin\n"))
	})

	It("reads files in order without implicitly reading stdin", func() {
		dir := GinkgoT().TempDir()
		first := writeFixture(dir, "first", "first\n")
		second := writeFixture(dir, "second", "second\n")
		var output bytes.Buffer

		err := runner.RunInputs(bytes.NewBufferString("ignored\n"), &output, io.Discard, nil, []string{first, second})

		Expect(err).NotTo(HaveOccurred())
		Expect(output.String()).To(Equal("first\nsecond\n"))
	})

	It("reads stdin where each dash appears", func() {
		dir := GinkgoT().TempDir()
		first := writeFixture(dir, "first", "first\n")
		second := writeFixture(dir, "second", "second\n")
		var output bytes.Buffer

		err := runner.RunInputs(bytes.NewBufferString("stdin\n"), &output, io.Discard, nil, []string{first, "-", "-", second})

		Expect(err).NotTo(HaveOccurred())
		Expect(output.String()).To(Equal("first\nstdin\nsecond\n"))
	})

	It("reports open errors immediately, continues, and returns ErrInput", func() {
		dir := GinkgoT().TempDir()
		missing := filepath.Join(dir, "missing")
		readable := writeFixture(dir, "readable", "still read\n")
		events := &eventLog{}

		err := runner.RunInputs(bytes.NewBuffer(nil), events.writer("stdout:"), events.writer("stderr:"), nil, []string{missing, readable})

		Expect(err).To(MatchError(runner.ErrInput))
		Expect(events.String()).To(ContainSubstring("stderr:" + missing))
		Expect(strings.Index(events.String(), "stderr:")).To(BeNumerically("<", strings.Index(events.String(), "stdout:")))
		Expect(events.String()).To(ContainSubstring("still read"))
	})

	It("reports read errors and continues with later operands", func() {
		dir := GinkgoT().TempDir()
		readable := writeFixture(dir, "readable", "after\n")
		stdin := io.MultiReader(strings.NewReader("before\n"), failingReader{})
		var output bytes.Buffer
		var diagnostics bytes.Buffer

		err := runner.RunInputs(stdin, &output, &diagnostics, nil, []string{"-", readable})

		Expect(err).To(MatchError(runner.ErrInput))
		Expect(output.String()).To(Equal("before\nafter\n"))
		Expect(diagnostics.String()).To(ContainSubstring("-: read failed"))
	})

	It("stops immediately when output fails", func() {
		dir := GinkgoT().TempDir()
		first := writeFixture(dir, "first", "first\n")
		second := writeFixture(dir, "second", "second\n")
		var diagnostics bytes.Buffer

		err := runner.RunInputs(bytes.NewBuffer(nil), failingWriter{}, &diagnostics, nil, []string{first, second})

		Expect(err).To(MatchError("write failed"))
		Expect(diagnostics.String()).To(BeEmpty())
	})

	It("stops when an input error cannot be reported", func() {
		dir := GinkgoT().TempDir()
		missing := filepath.Join(dir, "missing")
		readable := writeFixture(dir, "readable", "not reached\n")
		var output bytes.Buffer

		err := runner.RunInputs(bytes.NewBuffer(nil), &output, failingWriter{}, nil, []string{missing, readable})

		Expect(err).To(MatchError("write failed"))
		Expect(output.String()).To(BeEmpty())
	})

	It("resets highlighting at file boundaries", func() {
		dir := GinkgoT().TempDir()
		first := writeFixture(dir, "first", "ERR")
		second := writeFixture(dir, "second", "OR\n")
		rulesPath := writeFixture(dir, "rules.yaml", "rules:\n  - name: error\n    pattern: 'ERROR'\n    style: error\n")
		highlighter, err := runner.NewHighlighter(rulesPath, "")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(highlighter.Close)
		var output bytes.Buffer

		err = runner.RunInputs(bytes.NewBuffer(nil), &output, io.Discard, highlighter, []string{first, second})

		Expect(err).NotTo(HaveOccurred())
		Expect(output.String()).To(Equal("ERROR\n"))
		Expect(output.String()).NotTo(ContainSubstring("\x1b["))
	})
})

func writeFixture(dir, name, contents string) string {
	path := filepath.Join(dir, name)
	Expect(os.WriteFile(path, []byte(contents), 0o644)).To(Succeed())
	return path
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

type eventLog struct {
	strings.Builder
}

func (log *eventLog) writer(prefix string) io.Writer {
	return writerFunc(func(data []byte) (int, error) {
		_, _ = log.WriteString(prefix + string(data))
		return len(data), nil
	})
}

type writerFunc func([]byte) (int, error)

func (fn writerFunc) Write(data []byte) (int, error) { return fn(data) }
