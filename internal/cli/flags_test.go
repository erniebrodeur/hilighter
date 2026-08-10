package cli

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("parseOptions", func() {
	It("uses stdin mode when no arguments are provided", func() {
		opts, err := parseOptions(nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Inputs).To(BeEmpty())
	})

	It("parses repeated expressions, a theme, and ordered files", func() {
		opts, err := parseOptions([]string{"-e", "ERROR", "-e", "WARN", "-t", "themes/custom.yaml", "first", "second"})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Expressions).To(Equal([]string{"ERROR", "WARN"}))
		Expect(opts.ThemePath).To(Equal("themes/custom.yaml"))
		Expect(opts.Inputs).To(Equal([]string{"first", "second"}))
	})

	It("uses -- to permit dash-prefixed filenames", func() {
		opts, err := parseOptions([]string{"--", "-events.log"})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Inputs).To(Equal([]string{"-events.log"}))
	})

	It("treats all positional words as filenames", func() {
		opts, err := parseOptions([]string{"version", "validate", "tail"})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Inputs).To(Equal([]string{"version", "validate", "tail"}))
	})

	It("parses standalone help and version operations", func() {
		help, err := parseOptions([]string{"--help"})
		Expect(err).NotTo(HaveOccurred())
		Expect(help.ShowHelp).To(BeTrue())

		version, err := parseOptions([]string{"--version"})
		Expect(err).NotTo(HaveOccurred())
		Expect(version.ShowVersion).To(BeTrue())
	})

	It("rejects removed flags", func() {
		for _, name := range []string{"--app", "--rules", "--theme", "--cmd", "--config-dir", "--no-detect", "--debug-detect"} {
			_, err := parseOptions([]string{name})
			Expect(err).To(HaveOccurred(), name)
		}
	})

	It("rejects operands or another operation with help and version", func() {
		_, err := parseOptions([]string{"--help", "file"})
		Expect(err).To(HaveOccurred())
		_, err = parseOptions([]string{"--help", "--version"})
		Expect(err).To(HaveOccurred())
		_, err = parseOptions([]string{"-e", "ERROR", "--version"})
		Expect(err).To(HaveOccurred())
	})
})
