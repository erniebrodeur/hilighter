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

	It("ships a truecolor Monokai default theme", func() {
		th := theme.Default()

		Expect(th.Styles).To(HaveKey("error"))
		Expect(th.Styles["error"].FG).To(Equal("#f8f8f2"))
		Expect(th.Styles["error"].BG).To(Equal("#f92672"))
		Expect(th.Styles).To(HaveKey("warning"))
		Expect(th.Styles["warning"].FG).To(Equal("#e6db74"))
		Expect(th.Styles["warning"].Bold).To(BeTrue())
		Expect(th.Styles).To(HaveKey("bool-true"))
		Expect(th.Styles["bool-true"].FG).To(Equal("#a6e22e"))
		Expect(th.Styles["bool-true"].Bold).To(BeTrue())
		Expect(th.Styles).To(HaveKey("bool-false"))
		Expect(th.Styles["bool-false"].FG).To(Equal("#f92672"))
		Expect(th.Styles["bool-false"].Bold).To(BeTrue())
		Expect(th.Styles).To(HaveKey("endpoint"))
		Expect(th.Styles["endpoint"].FG).To(Equal("#819aff"))
		Expect(th.Styles["endpoint"].Bold).To(BeTrue())
		Expect(th.Styles).To(HaveKey("test-name"))
		Expect(th.Styles["test-name"].FG).To(Equal("#a6e22e"))
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

	It("registers every dark theme as an isolated built-in", func() {
		Expect(theme.BuiltinNames()).To(Equal([]string{
			"abyss",
			"dark-2026",
			"dark-modern",
			"dark-plus",
			"high-contrast",
			"kimbie-dark",
			"monokai",
			"monokai-dimmed",
			"red",
			"solarized-dark",
			"tomorrow-night-blue",
			"visual-studio-dark",
		}))

		for _, name := range theme.BuiltinNames() {
			builtIn, ok := theme.Builtin(name)
			Expect(ok).To(BeTrue(), name)
			Expect(builtIn.Styles).To(HaveLen(14), name)
			for _, label := range []string{"accent", "bool-false", "bool-true", "detail", "endpoint", "error", "host", "info", "notice", "process", "repeat", "test-name", "timestamp", "warning"} {
				Expect(builtIn.Styles).To(HaveKey(label), "%s: %s", name, label)
				Expect(builtIn.Styles[label].FG).To(MatchRegexp(`^#[0-9a-f]{6}$`), "%s: %s", name, label)
			}
			Expect(builtIn.Styles["error"].BG).To(MatchRegexp(`^#[0-9a-f]{6}$`), name)
		}

		first, ok := theme.Builtin("monokai")
		Expect(ok).To(BeTrue())
		first.Styles["error"] = theme.Style{FG: "green"}

		second, ok := theme.Builtin("monokai")
		Expect(ok).To(BeTrue())
		Expect(second.Styles["error"]).To(Equal(theme.Default().Styles["error"]))
	})

	It("rejects unknown built-in theme names", func() {
		_, ok := theme.Builtin("missing")
		Expect(ok).To(BeFalse())
	})

	It("uses an error background only for the error style in the default theme", func() {
		th := theme.Default()

		for name, style := range th.Styles {
			if name == "error" {
				Expect(style.BG).To(Equal("#f92672"))
				continue
			}
			Expect(style.BG).To(BeEmpty())
		}
	})

	It("resolves style labels used by rules and capture groups", func() {
		th := theme.Default()

		style, ok := th.Resolve("error")
		Expect(ok).To(BeTrue())
		Expect(style.FG).To(Equal("#f8f8f2"))
		Expect(style.BG).To(Equal("#f92672"))

		style, ok = th.Resolve("process")
		Expect(ok).To(BeTrue())
		Expect(style.FG).To(Equal("#e6db74"))

		style, ok = th.Resolve("bool-true")
		Expect(ok).To(BeTrue())
		Expect(style.FG).To(Equal("#a6e22e"))

		style, ok = th.Resolve("bool-false")
		Expect(ok).To(BeTrue())
		Expect(style.FG).To(Equal("#f92672"))

		style, ok = th.Resolve("endpoint")
		Expect(ok).To(BeTrue())
		Expect(style.FG).To(Equal("#819aff"))
	})

	It("overlays custom styles while retaining Monokai fallbacks", func() {
		custom := theme.Theme{Styles: map[string]theme.Style{
			"error": {FG: "green"},
		}}

		merged := theme.Overlay(theme.Default(), custom)

		Expect(merged.Styles["error"]).To(Equal(theme.Style{FG: "green"}))
		Expect(merged.Styles["warning"]).To(Equal(theme.Default().Styles["warning"]))
	})
})
