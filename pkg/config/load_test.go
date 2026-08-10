package config_test

import (
	"os"
	"path/filepath"

	"github.com/erniebrodeur/hilighter/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Default paths", func() {
	It("builds the default config directory under the user home", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")

		Expect(config.DefaultDir()).To(Equal(filepath.Join("/tmp/hilighter-home", ".hilighter")))
	})

	It("builds the default config file path under the config directory", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")

		Expect(config.DefaultConfigPath()).To(Equal(filepath.Join("/tmp/hilighter-home", ".hilighter", "config.yaml")))
	})

	It("builds the conventional rules path under the customization directory", func() {
		GinkgoT().Setenv("HOME", "/tmp/hilighter-home")

		Expect(config.DefaultRulesPath("")).To(Equal(filepath.Join("/tmp/hilighter-home", ".hilighter", "rules.yaml")))
	})
})

var _ = Describe("User customization layout", func() {
	It("creates theme-only config, empty rules, and an empty themes directory", func() {
		dir := filepath.Join(GinkgoT().TempDir(), ".hilighter")

		Expect(config.EnsureLayout(dir)).To(Succeed())

		rulesInfo, err := os.Stat(filepath.Join(dir, "rules.yaml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(rulesInfo.Size()).To(BeZero())
		themesInfo, err := os.Stat(filepath.Join(dir, "themes"))
		Expect(err).NotTo(HaveOccurred())
		Expect(themesInfo.IsDir()).To(BeTrue())
		configData, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(configData)).To(Equal("theme: monokai\n"))
	})

	It("preserves existing rules", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "rules.yaml")
		Expect(os.WriteFile(path, []byte("existing\n"), 0o600)).To(Succeed())

		Expect(config.EnsureLayout(dir)).To(Succeed())
		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("existing\n"))
	})

	It("preserves existing config", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "config.yaml")
		Expect(os.WriteFile(path, []byte("theme: themes/custom.yaml\n"), 0o600)).To(Succeed())

		Expect(config.EnsureLayout(dir)).To(Succeed())
		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("theme: themes/custom.yaml\n"))
	})
})

var _ = Describe("Theme config", func() {
	It("loads the theme selector", func() {
		path := filepath.Join(GinkgoT().TempDir(), "config.yaml")
		Expect(os.WriteFile(path, []byte("theme: themes/custom.yaml\n"), 0o644)).To(Succeed())

		value, err := config.LoadTheme(path)

		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal("themes/custom.yaml"))
	})

	It("rejects settings other than theme", func() {
		path := filepath.Join(GinkgoT().TempDir(), "config.yaml")
		Expect(os.WriteFile(path, []byte("rules: rules.yaml\n"), 0o644)).To(Succeed())

		_, err := config.LoadTheme(path)

		Expect(err).To(HaveOccurred())
	})
})
