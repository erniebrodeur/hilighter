package ansi

import (
	"strings"

	"github.com/erniebrodeur/hilighter/pkg/engine"
	"github.com/erniebrodeur/hilighter/pkg/theme"
)

// Renderer applies ANSI styling to rule-engine results.
type Renderer struct {
	theme theme.Theme
}

// New constructs a renderer for one theme.
func New(th theme.Theme) *Renderer {
	return &Renderer{theme: th}
}

// Render returns ANSI-styled text for one processed line.
func (r *Renderer) Render(result engine.Result) string {
	if len(result.Spans) == 0 {
		return result.Line
	}

	var out strings.Builder
	cursor := 0
	for _, span := range result.Spans {
		if span.Start > cursor {
			out.WriteString(result.Line[cursor:span.Start])
		}

		style, ok := r.theme.Resolve(span.Label)
		if !ok {
			out.WriteString(result.Line[span.Start:span.End])
			cursor = span.End
			continue
		}

		out.WriteString(sequence(style))
		out.WriteString(result.Line[span.Start:span.End])
		out.WriteString(resetSequence)
		cursor = span.End
	}

	if cursor < len(result.Line) {
		out.WriteString(result.Line[cursor:])
	}

	return out.String()
}
