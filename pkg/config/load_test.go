package config_test

import (
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
