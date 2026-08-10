package ansi

import (
	"fmt"
	"strconv"
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

	parts = append(parts, colorParts(style.FG, fgColors, "38")...)
	parts = append(parts, colorParts(style.BG, bgColors, "48")...)

	if len(parts) == 0 {
		return ""
	}

	return "\x1b[" + strings.Join(parts, ";") + "m"
}

func colorParts(value string, named map[string]int, extendedPrefix string) []string {
	if red, green, blue, ok := parseHexColor(value); ok {
		return []string{extendedPrefix, "2", strconv.Itoa(red), strconv.Itoa(green), strconv.Itoa(blue)}
	}

	code, ok := named[strings.ToLower(value)]
	if !ok {
		return nil
	}
	if code < 100 {
		return []string{fmt.Sprintf("%d", code)}
	}
	return []string{extendedPrefix, "5", fmt.Sprintf("%d", code)}
}

func parseHexColor(value string) (red, green, blue int, ok bool) {
	if len(value) != 7 || value[0] != '#' {
		return 0, 0, 0, false
	}
	parsed, err := strconv.ParseUint(value[1:], 16, 24)
	if err != nil {
		return 0, 0, 0, false
	}
	return int(parsed >> 16), int(parsed >> 8 & 0xff), int(parsed & 0xff), true
}
