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

		opts, err := parseOptions()

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.ConfigDir).To(Equal("/tmp/hilighter-home/.hilighter"))
		Expect(opts.App).To(BeEmpty())
		Expect(opts.Mode).To(BeEmpty())
		Expect(opts.Profile).To(BeEmpty())
		Expect(opts.RulesPath).To(BeEmpty())
		Expect(opts.ThemePath).To(BeEmpty())
		Expect(opts.Command).To(BeEmpty())
	})

	It("parses explicit flag values", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")
		os.Args = []string{
			"hilighter",
			"--app", "syslog",
			"--rules", "/tmp/rules.yaml",
			"--theme", "/tmp/theme.yaml",
			"--cmd", "printf test",
			"--config-dir", "/tmp/custom",
		}
		opts, err := parseOptions()

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.App).To(Equal("syslog"))
		Expect(opts.RulesPath).To(Equal("/tmp/rules.yaml"))
		Expect(opts.ThemePath).To(Equal("/tmp/theme.yaml"))
		Expect(opts.Command).To(Equal("printf test"))
		Expect(opts.ConfigDir).To(Equal("/tmp/custom"))
	})

	It("parses the tail subcommand with profile and optional file path", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")
		os.Args = []string{"hilighter", "tail", "rails-log", "log/development.log"}

		opts, err := parseOptions()

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Mode).To(Equal("tail"))
		Expect(opts.Profile).To(Equal("rails-log"))
		Expect(opts.FilePath).To(Equal("log/development.log"))
	})

	It("parses the cat subcommand with profile and optional file path", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")
		os.Args = []string{"hilighter", "cat", "rails-log", "log/development.log"}

		opts, err := parseOptions()

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Mode).To(Equal("cat"))
		Expect(opts.Profile).To(Equal("rails-log"))
		Expect(opts.FilePath).To(Equal("log/development.log"))
	})

	It("parses the head subcommand with profile and optional file path", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")
		os.Args = []string{"hilighter", "head", "rails-log", "log/development.log"}

		opts, err := parseOptions()

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Mode).To(Equal("head"))
		Expect(opts.Profile).To(Equal("rails-log"))
		Expect(opts.FilePath).To(Equal("log/development.log"))
	})

	It("accepts a file-mode target without knowing yet whether it is a profile or path", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")
		os.Args = []string{"hilighter", "tail", "log/development.log"}

		opts, err := parseOptions()

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Mode).To(Equal("tail"))
		Expect(opts.Profile).To(Equal("log/development.log"))
		Expect(opts.FilePath).To(BeEmpty())
	})

	It("parses the validate subcommand", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")
		os.Args = []string{"hilighter", "validate"}

		opts, err := parseOptions()

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Mode).To(Equal("validate"))
	})

	It("parses the list subcommand", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")
		os.Args = []string{"hilighter", "list", "apps"}

		opts, err := parseOptions()

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Mode).To(Equal("list"))
		Expect(opts.Subject).To(Equal("apps"))
	})

	It("parses the show subcommand", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")
		os.Args = []string{"hilighter", "show", "app", "docker"}

		opts, err := parseOptions()

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Mode).To(Equal("show"))
		Expect(opts.Subject).To(Equal("app"))
		Expect(opts.Name).To(Equal("docker"))
	})

	It("parses detection flags", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")
		os.Args = []string{"hilighter", "--no-detect", "--debug-detect", "tail", "log/development.log"}

		opts, err := parseOptions()

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.NoDetect).To(BeTrue())
		Expect(opts.DebugDetect).To(BeTrue())
		Expect(opts.Mode).To(Equal("tail"))
	})

	It("parses the version flag", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")
		os.Args = []string{"hilighter", "--version"}

		opts, err := parseOptions()

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.ShowVersion).To(BeTrue())
	})

	It("parses the version subcommand", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")
		os.Args = []string{"hilighter", "version"}

		opts, err := parseOptions()

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.ShowVersion).To(BeTrue())
	})
})
