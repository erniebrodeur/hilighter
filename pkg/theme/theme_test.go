package theme_test

import (
	"os"
	"path/filepath"

	"github.com/erniebrodeur/hilighter/pkg/theme"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Theme loading", func() {
	It("loads semantic styles from YAML", func() {
		path := filepath.Join(GinkgoT().TempDir(), "theme.yaml")
		Expect(os.WriteFile(path, []byte("styles:\n  error:\n    fg: pink\n    bold: true\n"), 0o644)).To(Succeed())

		th, err := theme.Load(path)

		Expect(err).NotTo(HaveOccurred())
		Expect(th.Styles).To(HaveKey("error"))
		Expect(th.Styles["error"].FG).To(Equal("pink"))
		Expect(th.Styles["error"].Bold).To(BeTrue())
	})

	It("ships a Monokai-style default theme", func() {
		th := theme.Default()

		Expect(th.Styles).To(HaveKey("error"))
		Expect(th.Styles["error"].FG).To(Equal("pink"))
		Expect(th.Styles).To(HaveKey("warning"))
		Expect(th.Styles["warning"].FG).To(Equal("yellow"))
		Expect(th.Styles).To(HaveKey("test-name"))
		Expect(th.Styles["test-name"].FG).To(Equal("green"))
	})

	It("keeps the default theme foreground-only without background colors", func() {
		th := theme.Default()

		for _, style := range th.Styles {
			Expect(style.BG).To(BeEmpty())
		}
	})

	It("resolves style labels used by rules and capture groups", func() {
		th := theme.Default()

		style, ok := th.Resolve("error")
		Expect(ok).To(BeTrue())
		Expect(style.FG).To(Equal("pink"))

		style, ok = th.Resolve("test-name")
		Expect(ok).To(BeTrue())
		Expect(style.FG).To(Equal("green"))
	})
})
