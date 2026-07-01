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

		Expect(output).To(ContainSubstring(";41m"))
	})

	It("leaves unmatched text untouched", func() {
		renderer := ansi.New(theme.Default())

		output := renderer.Render(engine.Result{Line: "plain text"})

		Expect(output).To(Equal("plain text"))
	})
})
