package ansi_test

import (
	"strings"

	"github.com/erniebrodeur/hilighter/pkg/engine"
	"github.com/erniebrodeur/hilighter/pkg/render/ansi"
	"github.com/erniebrodeur/hilighter/pkg/theme"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ANSI renderer", func() {
	It("renders styled substrings with resets around each highlighted span", func() {
		renderer := ansi.New(theme.Default())

		output := renderer.Render(engine.Result{
			Line: "prefix ERROR suffix",
			Spans: []engine.Span{{
				Start: 7, End: 12, Label: "error",
			}},
		})

		Expect(output).To(ContainSubstring("prefix "))
		Expect(output).To(ContainSubstring("ERROR"))
		Expect(output).To(ContainSubstring("\x1b["))
		Expect(strings.Count(output, "\x1b[0m")).To(Equal(1))
	})

	It("renders whole-line styling without changing the underlying text", func() {
		renderer := ansi.New(theme.Default())
		line := "warning: disk low"

		output := renderer.Render(engine.Result{
			Line: line,
			Spans: []engine.Span{{
				Start: 0, End: len(line), Label: "warning",
			}},
		})

		Expect(output).To(ContainSubstring(line))
		Expect(strings.Count(output, "\x1b[0m")).To(Equal(1))
	})

	It("emits a background sequence for default error styling", func() {
		renderer := ansi.New(theme.Default())

		output := renderer.Render(engine.Result{
			Line: "ERROR",
			Spans: []engine.Span{{
				Start: 0, End: 5, Label: "error",
			}},
		})

		Expect(output).To(ContainSubstring(";48;2;249;38;114m"))
	})

	It("renders exact truecolor foregrounds and backgrounds", func() {
		renderer := ansi.New(theme.Theme{Styles: map[string]theme.Style{
			"error": {FG: "#1e90ff", BG: "#A020F0", Bold: true},
		}})

		output := renderer.Render(engine.Result{
			Line: "ERROR",
			Spans: []engine.Span{{
				Start: 0, End: 5, Label: "error",
			}},
		})

		Expect(output).To(Equal("\x1b[1;38;2;30;144;255;48;2;160;32;240mERROR\x1b[0m"))
	})

	It("leaves unmatched text untouched", func() {
		renderer := ansi.New(theme.Default())

		output := renderer.Render(engine.Result{Line: "plain text"})

		Expect(output).To(Equal("plain text"))
	})
})
