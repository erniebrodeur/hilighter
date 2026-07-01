package runner_test

import (
	"bytes"

	"github.com/erniebrodeur/hilighter/pkg/runner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunCommand", func() {
	It("forwards stdout and stderr from the executed command", func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		err := runner.RunCommand("printf 'out'; printf 'err' >&2", &stdout, &stderr)

		Expect(err).NotTo(HaveOccurred())
		Expect(stdout.String()).To(Equal("out"))
		Expect(stderr.String()).To(Equal("err"))
	})

	It("returns an error when the command exits non-zero", func() {
		err := runner.RunCommand("exit 7", &bytes.Buffer{}, &bytes.Buffer{})

		Expect(err).To(HaveOccurred())
	})

	It("eventually streams command output through the highlighting engine", Pending)
})
