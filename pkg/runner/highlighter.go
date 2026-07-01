package runner

import (
	"github.com/erniebrodeur/hilighter/pkg/engine"
	"github.com/erniebrodeur/hilighter/pkg/render/ansi"
	"github.com/erniebrodeur/hilighter/pkg/rules"
	"github.com/erniebrodeur/hilighter/pkg/theme"
)

// Highlighter binds compiled rules to one renderer for stream processing.
type Highlighter struct {
	engine   *engine.Engine
	renderer *ansi.Renderer
}

// NewHighlighter loads rules and theme data for stream processing.
//
// If themePath is empty, the built-in default theme is used. If neither a rule
// path nor built-in app profile is selected, the returned highlighter behaves
// like a passthrough.
func NewHighlighter(rulesPath, appName, themePath string) (*Highlighter, error) {
	ruleFile, err := resolveRuleFile(rulesPath, appName)
	if err != nil {
		return nil, err
	}
	if ruleFile == nil {
		return &Highlighter{}, nil
	}

	compiled, err := rules.Compile(ruleFile.Rules)
	if err != nil {
		return nil, err
	}

	th := theme.Default()
	if themePath != "" {
		th, err = theme.Load(themePath)
		if err != nil {
			rules.Close(compiled)
			return nil, err
		}
	}

	return &Highlighter{
		engine:   engine.New(compiled),
		renderer: ansi.New(th),
	}, nil
}

func resolveRuleFile(rulesPath, appName string) (*rules.File, error) {
	if rulesPath != "" {
		file, err := rules.Load(rulesPath)
		if err != nil {
			return nil, err
		}
		return &file, nil
	}

	if appName != "" {
		file, ok := rules.Builtin(appName)
		if !ok {
			return nil, rules.ErrUnknownBuiltin(appName)
		}
		return &file, nil
	}

	return nil, nil
}

// Close releases the compiled PCRE resources owned by the highlighter.
func (h *Highlighter) Close() {
	if h == nil || h.engine == nil {
		return
	}
	h.engine.Close()
}

// ProcessLine returns ANSI-styled output for a single line.
func (h *Highlighter) ProcessLine(line string) string {
	if h == nil || h.engine == nil || h.renderer == nil {
		return line
	}

	return h.renderer.Render(h.engine.ProcessLine(line))
}
