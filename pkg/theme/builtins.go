package theme

import "sort"

// DefaultName is the stable selector for hilighter's default built-in theme.
const DefaultName = "monokai"

var builtins = map[string]Theme{
	"abyss": fromPalette(palette{
		foreground: "#cccccc", muted: "#596f99", red: "#ff7882", green: "#b8f171",
		yellow: "#ffe580", blue: "#80baff", magenta: "#d778ff", cyan: "#78ffff",
	}),
	"dark-2026": fromPalette(palette{
		foreground: "#bbbebf", muted: "#858889", red: "#f48771", green: "#72c892",
		yellow: "#e5ba7d", blue: "#3994bc", magenta: "#c586c0", cyan: "#48a0c7",
	}),
	"dark-modern": fromPalette(palette{
		foreground: "#cccccc", muted: "#6e7681", red: "#f85149", green: "#4ec9b0",
		yellow: "#dcdcaa", blue: "#0078d4", magenta: "#c586c0", cyan: "#9cdcfe",
	}),
	"dark-plus": fromPalette(palette{
		foreground: "#d4d4d4", muted: "#6a9955", red: "#f44747", green: "#4ec9b0",
		yellow: "#dcdcaa", blue: "#569cd6", magenta: "#c586c0", cyan: "#9cdcfe",
	}),
	"high-contrast": fromPalette(palette{
		foreground: "#ffffff", muted: "#d4d4d4", red: "#f48771", green: "#4ec9b0",
		yellow: "#dcdcaa", blue: "#569cd6", magenta: "#c586c0", cyan: "#9cdcfe",
	}),
	"kimbie-dark": fromPalette(palette{
		foreground: "#d3af86", muted: "#a57a4c", red: "#dc3958", green: "#889b4a",
		yellow: "#f79a32", blue: "#8ab1b0", magenta: "#98676a", cyan: "#8ab1b0",
	}),
	DefaultName: fromPalette(palette{
		foreground: "#f8f8f2", muted: "#88846f", red: "#f92672", green: "#a6e22e",
		yellow: "#e6db74", blue: "#819aff", magenta: "#ae81ff", cyan: "#66d9ef",
	}),
	"monokai-dimmed": fromPalette(palette{
		foreground: "#c5c8c6", muted: "#9a9b99", red: "#c7444a", green: "#9aa83a",
		yellow: "#d0b344", blue: "#6089b4", magenta: "#9872a2", cyan: "#6089b4",
	}),
	"red": fromPalette(palette{
		foreground: "#f8f8f8", muted: "#cd8d8d", red: "#ff0000", green: "#ffe862",
		yellow: "#fec758", blue: "#ff7777", magenta: "#ff6666", cyan: "#ffb454",
	}),
	"solarized-dark": fromPalette(palette{
		foreground: "#eee8d5", muted: "#586e75", red: "#dc322f", green: "#859900",
		yellow: "#b58900", blue: "#268bd2", magenta: "#d33682", cyan: "#2aa198",
	}),
	"tomorrow-night-blue": fromPalette(palette{
		foreground: "#ffffff", muted: "#7285b7", red: "#ff7882", green: "#b8f171",
		yellow: "#ffe580", blue: "#80baff", magenta: "#d778ff", cyan: "#78ffff",
	}),
	"visual-studio-dark": fromPalette(palette{
		foreground: "#d4d4d4", muted: "#6a9955", red: "#f44747", green: "#b5cea8",
		yellow: "#dcdcaa", blue: "#569cd6", magenta: "#c586c0", cyan: "#9cdcfe",
	}),
}

// Default returns the built-in Monokai-style default theme.
func Default() Theme {
	th, _ := Builtin(DefaultName)
	return th
}

// Builtin resolves a built-in theme by its stable selector.
func Builtin(name string) (Theme, bool) {
	th, ok := builtins[name]
	if !ok {
		return Theme{}, false
	}
	return clone(th), true
}

// BuiltinNames returns every stable built-in selector in lexical order.
func BuiltinNames() []string {
	names := make([]string, 0, len(builtins))
	for name := range builtins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func clone(th Theme) Theme {
	styles := make(map[string]Style, len(th.Styles))
	for name, style := range th.Styles {
		styles[name] = style
	}
	return Theme{Styles: styles}
}

type palette struct {
	foreground string
	muted      string
	red        string
	green      string
	yellow     string
	blue       string
	magenta    string
	cyan       string
}

// fromPalette adapts editor-theme colors to hilighter's semantic labels.
func fromPalette(colors palette) Theme {
	return Theme{
		Styles: map[string]Style{
			"accent":     {FG: colors.cyan, Bold: true},
			"bool-false": {FG: colors.red, Bold: true},
			"bool-true":  {FG: colors.green, Bold: true},
			"detail":     {FG: colors.foreground},
			"endpoint":   {FG: colors.blue, Bold: true},
			"error":      {FG: colors.foreground, BG: colors.red, Bold: true},
			"host":       {FG: colors.green, Bold: true},
			"info":       {FG: colors.cyan, Bold: true},
			"notice":     {FG: colors.cyan},
			"process":    {FG: colors.yellow, Bold: true},
			"repeat":     {FG: colors.magenta},
			"test-name":  {FG: colors.green},
			"timestamp":  {FG: colors.muted},
			"warning":    {FG: colors.yellow, Bold: true},
		},
	}
}
