package cli

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("resolveOptions", func() {
	It("uses the built-in app command when --cmd is not provided", func() {
		opts, err := resolveOptions(Options{
			App:       "syslog",
			ConfigDir: GinkgoT().TempDir(),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Command).To(Equal("/usr/bin/log stream --style syslog"))
	})

	It("prefers an explicit command over the built-in app command", func() {
		opts, err := resolveOptions(Options{
			App:       "syslog",
			Command:   "printf custom",
			ConfigDir: GinkgoT().TempDir(),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Command).To(Equal("printf custom"))
	})

	It("loads a default command from an explicit rules file", func() {
		rulesPath := filepath.Join(GinkgoT().TempDir(), "rules.yaml")
		Expect(os.WriteFile(rulesPath, []byte("cmd: printf from-rules\nrules:\n  - name: error\n    pattern: 'ERROR'\n    style: error\n"), 0o644)).To(Succeed())

		opts, err := resolveOptions(Options{
			RulesPath: rulesPath,
			ConfigDir: GinkgoT().TempDir(),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Command).To(Equal("printf from-rules"))
	})
})
