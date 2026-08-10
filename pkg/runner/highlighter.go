package runner

import (
	"fmt"

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
// path nor built-in app profile is selected, the built-in default rules are
// used.
func NewHighlighter(rulesPath, appName, themePath string) (*Highlighter, error) {
	return NewHighlighterWithExpressions(nil, rulesPath, appName, themePath)
}

// NewHighlighterWithExpressions resolves expression, file, and built-in rule
// sources in precedence order. Repeated expressions replace every other rule
// source and use the accent style.
func NewHighlighterWithExpressions(expressions []string, rulesPath, appName, themePath string) (*Highlighter, error) {
	ruleFile, err := resolveRuleFile(expressions, rulesPath, appName)
	if err != nil {
		return nil, err
	}

	compiled, err := rules.Compile(ruleFile.Rules)
	if err != nil {
		return nil, err
	}

	th := theme.Default()
	if themePath != "" && themePath != "monokai" {
		custom, loadErr := theme.Load(themePath)
		err = loadErr
		if err != nil {
			rules.Close(compiled)
			return nil, err
		}
		th = theme.Overlay(th, custom)
	}

	return &Highlighter{
		engine:   engine.New(compiled),
		renderer: ansi.New(th),
	}, nil
}

func resolveRuleFile(expressions []string, rulesPath, appName string) (rules.File, error) {
	if len(expressions) > 0 {
		file := rules.File{Rules: make([]rules.Spec, 0, len(expressions))}
		for i, pattern := range expressions {
			if pattern == "" {
				return rules.File{}, fmt.Errorf("expression %d is empty", i+1)
			}
			file.Rules = append(file.Rules, rules.Spec{
				Name:    fmt.Sprintf("expression-%d", i+1),
				Pattern: pattern,
				Style:   "accent",
			})
		}
		return file, nil
	}

	if rulesPath != "" {
		file, err := rules.Load(rulesPath)
		if err != nil {
			return rules.File{}, err
		}
		return file, nil
	}

	if appName != "" {
		file, ok := rules.Builtin(appName)
		if !ok {
			return rules.File{}, rules.ErrUnknownBuiltin(appName)
		}
		return file, nil
	}

	return rules.Default(), nil
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
