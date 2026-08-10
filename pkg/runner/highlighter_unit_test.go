package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHighlighterUsesShippedDefaults(t *testing.T) {
	highlighter, err := NewHighlighter("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer highlighter.Close()
	if output := highlighter.ProcessLine("ERROR"); !strings.Contains(output, "\x1b[") {
		t.Fatalf("expected ANSI highlighting, got %q", output)
	}
}

func TestHighlighterSelectsBuiltInThemeByName(t *testing.T) {
	highlighter, err := NewHighlighter("", "monokai")
	if err != nil {
		t.Fatal(err)
	}
	defer highlighter.Close()
	if output := highlighter.ProcessLine("ERROR"); !strings.Contains(output, ";48;2;249;38;114m") {
		t.Fatalf("expected named Monokai error styling, got %q", output)
	}
}

func TestHighlighterSelectsAnotherBuiltInThemeByName(t *testing.T) {
	highlighter, err := NewHighlighter("", "dark-plus")
	if err != nil {
		t.Fatal(err)
	}
	defer highlighter.Close()
	if output := highlighter.ProcessLine("ERROR"); !strings.Contains(output, ";48;2;244;71;71m") {
		t.Fatalf("expected Dark+ error styling, got %q", output)
	}
}

func TestHighlighterUsesCustomRules(t *testing.T) {
	rulesPath := filepath.Join(t.TempDir(), "rules.yaml")
	if err := os.WriteFile(rulesPath, []byte("rules:\n  - name: custom\n    pattern: CUSTOM\n    style: accent\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	highlighter, err := NewHighlighter(rulesPath, "")
	if err != nil {
		t.Fatal(err)
	}
	defer highlighter.Close()
	if output := highlighter.ProcessLine("CUSTOM"); !strings.Contains(output, "\x1b[") {
		t.Fatalf("expected custom highlighting, got %q", output)
	}
	if output := highlighter.ProcessLine("ERROR"); strings.Contains(output, "\x1b[") {
		t.Fatalf("custom rules should replace defaults, got %q", output)
	}
}

func TestExpressionsReplaceFileRules(t *testing.T) {
	highlighter, err := NewHighlighterWithExpressions([]string{"CUSTOM"}, "/not/read", "")
	if err != nil {
		t.Fatal(err)
	}
	defer highlighter.Close()
	if output := highlighter.ProcessLine("CUSTOM"); !strings.Contains(output, "\x1b[") {
		t.Fatalf("expected expression highlighting, got %q", output)
	}
}

func TestEmptyExpressionFails(t *testing.T) {
	if _, err := NewHighlighterWithExpressions([]string{""}, "", ""); err == nil {
		t.Fatal("expected empty expression to fail")
	}
}
