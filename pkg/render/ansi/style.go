package ansi

import (
	"fmt"
	"strings"

	"github.com/erniebrodeur/hilighter/pkg/theme"
)

const resetSequence = "\x1b[0m"

var fgColors = map[string]int{
	"black":   30,
	"blue":    33,
	"cyan":    81,
	"gray":    244,
	"green":   148,
	"magenta": 35,
	"orange":  208,
	"pink":    197,
	"red":     203,
	"white":   252,
	"yellow":  221,
}

var bgColors = map[string]int{
	"black":   40,
	"blue":    44,
	"cyan":    46,
	"gray":    100,
	"green":   42,
	"magenta": 45,
	"orange":  208,
	"pink":    197,
	"red":     41,
	"white":   47,
	"yellow":  43,
}

func sequence(style theme.Style) string {
	parts := make([]string, 0, 3)
	if style.Bold {
		parts = append(parts, "1")
	}

	if code, ok := fgColors[strings.ToLower(style.FG)]; ok {
		if code < 100 {
			parts = append(parts, fmt.Sprintf("%d", code))
		} else {
			parts = append(parts, fmt.Sprintf("38;5;%d", code))
		}
	}

	if code, ok := bgColors[strings.ToLower(style.BG)]; ok {
		if code < 100 {
			parts = append(parts, fmt.Sprintf("%d", code))
		} else {
			parts = append(parts, fmt.Sprintf("48;5;%d", code))
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return "\x1b[" + strings.Join(parts, ";") + "m"
}
