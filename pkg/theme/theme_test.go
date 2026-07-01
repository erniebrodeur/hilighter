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
		Expect(th.Styles["error"].FG).To(Equal("white"))
		Expect(th.Styles["error"].BG).To(Equal("red"))
		Expect(th.Styles).To(HaveKey("warning"))
		Expect(th.Styles["warning"].FG).To(Equal("yellow"))
		Expect(th.Styles).To(HaveKey("test-name"))
		Expect(th.Styles["test-name"].FG).To(Equal("green"))
		Expect(th.Styles).To(HaveKey("timestamp"))
		Expect(th.Styles).To(HaveKey("host"))
		Expect(th.Styles).To(HaveKey("process"))
		Expect(th.Styles).To(HaveKey("notice"))
		Expect(th.Styles).To(HaveKey("repeat"))
	})

	It("uses a red background only for the error style in the default theme", func() {
		th := theme.Default()

		for name, style := range th.Styles {
			if name == "error" {
				Expect(style.BG).To(Equal("red"))
				continue
			}
			Expect(style.BG).To(BeEmpty())
		}
	})

	It("resolves style labels used by rules and capture groups", func() {
		th := theme.Default()

		style, ok := th.Resolve("error")
		Expect(ok).To(BeTrue())
		Expect(style.FG).To(Equal("white"))
		Expect(style.BG).To(Equal("red"))

		style, ok = th.Resolve("process")
		Expect(ok).To(BeTrue())
		Expect(style.FG).To(Equal("orange"))
	})
})
