package theme

// Default returns the built-in Monokai-style default theme.
func Default() Theme {
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
