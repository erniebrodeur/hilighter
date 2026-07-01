package theme

// Default returns the built-in Monokai-style default theme.
func Default() Theme {
	return Theme{
		Styles: map[string]Style{
			"accent":    {FG: "cyan"},
			"detail":    {FG: "white"},
			"error":     {FG: "white", BG: "red", Bold: true},
			"host":      {FG: "green"},
			"info":      {FG: "cyan"},
			"notice":    {FG: "cyan"},
			"process":   {FG: "orange"},
			"repeat":    {FG: "magenta"},
			"test-name": {FG: "green"},
			"timestamp": {FG: "gray"},
			"warning":   {FG: "yellow"},
		},
	}
}
