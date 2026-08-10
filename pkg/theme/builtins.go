package theme

import "sort"

// DefaultName is the stable selector for hilighter's default built-in theme.
const DefaultName = "monokai"

var builtins = map[string]Theme{
	DefaultName: monokai(),
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

func monokai() Theme {
	return Theme{
		Styles: map[string]Style{
			"accent":     {FG: "cyan", Bold: true},
			"bool-false": {FG: "red", Bold: true},
			"bool-true":  {FG: "green", Bold: true},
			"detail":     {FG: "white"},
			"endpoint":   {FG: "orange", Bold: true},
			"error":      {FG: "white", BG: "red", Bold: true},
			"host":       {FG: "green", Bold: true},
			"info":       {FG: "cyan", Bold: true},
			"notice":     {FG: "cyan"},
			"process":    {FG: "orange", Bold: true},
			"repeat":     {FG: "magenta"},
			"test-name":  {FG: "green"},
			"timestamp":  {FG: "gray"},
			"warning":    {FG: "yellow", Bold: true},
		},
	}
}
