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
})

var _ = Describe("Config loading", func() {
	It("loads defaults from ~/.hilighter/config.yaml", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		configDir := filepath.Join(os.Getenv("HOME"), ".hilighter")
		Expect(os.MkdirAll(configDir, 0o755)).To(Succeed())
		path := filepath.Join(configDir, "config.yaml")
		Expect(os.WriteFile(path, []byte("rules: ~/.hilighter/rules.yaml\ntheme: ~/.hilighter/themes/default.yaml\n"), 0o644)).To(Succeed())

		cfg, err := config.Load("")

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.RulesPath).To(Equal(filepath.Join(os.Getenv("HOME"), ".hilighter", "rules.yaml")))
		Expect(cfg.ThemePath).To(Equal(filepath.Join(os.Getenv("HOME"), ".hilighter", "themes", "default.yaml")))
	})

	It("loads an explicit config path", func() {
		path := filepath.Join(GinkgoT().TempDir(), "config.yaml")
		Expect(os.WriteFile(path, []byte("rules: /tmp/rules.yaml\ntheme: /tmp/theme.yaml\n"), 0o644)).To(Succeed())

		cfg, err := config.Load(path)

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.RulesPath).To(Equal("/tmp/rules.yaml"))
		Expect(cfg.ThemePath).To(Equal("/tmp/theme.yaml"))
	})

	It("loads named profiles and expands home-relative paths within them", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		path := filepath.Join(GinkgoT().TempDir(), "config.yaml")
		Expect(os.WriteFile(path, []byte("profiles:\n  rails-log:\n    rules: ~/.hilighter/rules/rails.yaml\n    theme: ~/.hilighter/themes/default.yaml\n    file: ~/work/app/log/development.log\n"), 0o644)).To(Succeed())

		cfg, err := config.Load(path)

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Profiles).To(HaveKey("rails-log"))
		Expect(cfg.Profiles["rails-log"].RulesPath).To(Equal(filepath.Join(os.Getenv("HOME"), ".hilighter", "rules", "rails.yaml")))
		Expect(cfg.Profiles["rails-log"].ThemePath).To(Equal(filepath.Join(os.Getenv("HOME"), ".hilighter", "themes", "default.yaml")))
		Expect(cfg.Profiles["rails-log"].FilePath).To(Equal(filepath.Join(os.Getenv("HOME"), "work", "app", "log", "development.log")))
	})
})
