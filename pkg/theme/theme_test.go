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
		Expect(th.Styles["warning"].Bold).To(BeTrue())
		Expect(th.Styles).To(HaveKey("bool-true"))
		Expect(th.Styles["bool-true"].FG).To(Equal("green"))
		Expect(th.Styles["bool-true"].Bold).To(BeTrue())
		Expect(th.Styles).To(HaveKey("bool-false"))
		Expect(th.Styles["bool-false"].FG).To(Equal("red"))
		Expect(th.Styles["bool-false"].Bold).To(BeTrue())
		Expect(th.Styles).To(HaveKey("endpoint"))
		Expect(th.Styles["endpoint"].FG).To(Equal("orange"))
		Expect(th.Styles["endpoint"].Bold).To(BeTrue())
		Expect(th.Styles).To(HaveKey("test-name"))
		Expect(th.Styles["test-name"].FG).To(Equal("green"))
		Expect(th.Styles).To(HaveKey("timestamp"))
		Expect(th.Styles).To(HaveKey("host"))
		Expect(th.Styles["host"].Bold).To(BeTrue())
		Expect(th.Styles).To(HaveKey("process"))
		Expect(th.Styles["process"].Bold).To(BeTrue())
		Expect(th.Styles).To(HaveKey("notice"))
		Expect(th.Styles).To(HaveKey("repeat"))
		Expect(th.Styles["accent"].Bold).To(BeTrue())
		Expect(th.Styles["info"].Bold).To(BeTrue())
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

		style, ok = th.Resolve("bool-true")
		Expect(ok).To(BeTrue())
		Expect(style.FG).To(Equal("green"))

		style, ok = th.Resolve("bool-false")
		Expect(ok).To(BeTrue())
		Expect(style.FG).To(Equal("red"))

		style, ok = th.Resolve("endpoint")
		Expect(ok).To(BeTrue())
		Expect(style.FG).To(Equal("orange"))
	})
})
