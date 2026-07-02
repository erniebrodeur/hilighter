package cli

import (
	"io/fs"
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

	It("uses the docker built-in app command when --cmd is not provided", func() {
		opts, err := resolveOptions(Options{
			App:       "docker",
			ConfigDir: GinkgoT().TempDir(),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Command).To(Equal("docker ps -a"))
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

var _ = Describe("shouldRunCommand", func() {
	It("runs an explicit command even when stdin is piped", func() {
		run := shouldRunCommand(
			Options{Command: "printf explicit"},
			Options{Command: "printf explicit"},
			0,
		)

		Expect(run).To(BeTrue())
	})

	It("uses stdin when a built-in app only contributes a default command and stdin is piped", func() {
		run := shouldRunCommand(
			Options{App: "docker"},
			Options{App: "docker", Command: "docker ps -a"},
			0,
		)

		Expect(run).To(BeFalse())
	})

	It("runs the built-in default command when stdin is interactive", func() {
		run := shouldRunCommand(
			Options{App: "docker"},
			Options{App: "docker", Command: "docker ps -a"},
			fs.ModeCharDevice,
		)

		Expect(run).To(BeTrue())
	})
})
