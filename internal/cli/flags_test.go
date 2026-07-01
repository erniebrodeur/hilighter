package cli

import (
	"flag"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("parseOptions", func() {
	var originalArgs []string
	var originalCommandLine *flag.FlagSet

	BeforeEach(func() {
		originalArgs = append([]string(nil), os.Args...)
		originalCommandLine = flag.CommandLine
		flag.CommandLine = flag.NewFlagSet("hilighter-test", flag.ContinueOnError)
	})

	AfterEach(func() {
		os.Args = originalArgs
		flag.CommandLine = originalCommandLine
	})

	It("uses the default config directory when no flags are provided", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")
		os.Args = []string{"hilighter"}

		opts := parseOptions()

		Expect(opts.ConfigDir).To(Equal("/tmp/hilighter-home/.hilighter"))
		Expect(opts.RulesPath).To(BeEmpty())
		Expect(opts.ThemePath).To(BeEmpty())
		Expect(opts.Command).To(BeEmpty())
	})

	It("parses explicit flag values", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")
		os.Args = []string{
			"hilighter",
			"--rules", "/tmp/rules.yaml",
			"--theme", "/tmp/theme.yaml",
			"--cmd", "printf test",
			"--config-dir", "/tmp/custom",
		}
		opts := parseOptions()

		Expect(opts.RulesPath).To(Equal("/tmp/rules.yaml"))
		Expect(opts.ThemePath).To(Equal("/tmp/theme.yaml"))
		Expect(opts.Command).To(Equal("printf test"))
		Expect(opts.ConfigDir).To(Equal("/tmp/custom"))
	})
})
