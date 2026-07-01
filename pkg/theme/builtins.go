package theme

// Default returns the built-in Monokai-style default theme.
func Default() Theme {
	return Theme{
		Styles: map[string]Style{
			"accent":    {FG: "cyan"},
			"detail":    {FG: "white"},
			"error":     {FG: "pink", Bold: true},
			"info":      {FG: "cyan"},
			"test-name": {FG: "green"},
			"warning":   {FG: "yellow"},
		},
	}
}
