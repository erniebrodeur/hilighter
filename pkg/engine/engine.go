package engine

import (
	"bufio"
	"io"
	"sort"
	"strings"

	"github.com/erniebrodeur/hilighter/pkg/rules"
)

// Span is a highlighted segment within one processed line.
type Span struct {
	// Start is the byte offset where the styled segment begins.
	Start int
	// End is the byte offset where the styled segment ends.
	End int
	// Label is the semantic style name chosen by the matched rule.
	Label string
	// RuleName identifies which rule produced the span.
	RuleName string
	// Group is the capture group index that produced the span, or 0 for a full match.
	Group int
}

// Result is the rule-engine output for a single line.
type Result struct {
	// Line is the original unmodified line content.
	Line string
	// Spans are the semantic regions selected for rendering.
	Spans []Span
}

// Engine applies ordered rules to lines of text.
type Engine struct {
	rules []rules.Compiled
}

// New constructs an Engine from compiled rules.
func New(compiled []rules.Compiled) *Engine {
	return &Engine{rules: compiled}
}

// Close releases the underlying compiled PCRE expressions.
func (e *Engine) Close() {
	rules.Close(e.rules)
}

// ProcessLine applies the first matching rule to one line.
func (e *Engine) ProcessLine(line string) Result {
	for _, rule := range e.rules {
		indices := rule.Regexp.FindStringSubmatchIndex(line)
		if len(indices) == 0 {
			continue
		}

		return Result{
			Line:  line,
			Spans: spansForRule(rule, line, indices),
		}
	}

	return Result{Line: line}
}

// ProcessText splits text into lines and applies ProcessLine to each one.
func (e *Engine) ProcessText(text string) []Result {
	scanner := bufio.NewScanner(strings.NewReader(text))
	results := make([]Result, 0)
	for scanner.Scan() {
		results = append(results, e.ProcessLine(scanner.Text()))
	}
	return results
}

// ProcessReader streams line-oriented results from an input reader.
func (e *Engine) ProcessReader(r io.Reader) ([]Result, error) {
	scanner := bufio.NewScanner(r)
	results := make([]Result, 0)
	for scanner.Scan() {
		results = append(results, e.ProcessLine(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func spansForRule(rule rules.Compiled, line string, indices []int) []Span {
	if rule.Scope == rules.ScopeLine {
		label := rule.Style
		if label == "" {
			label = rule.Name
		}
		if label == "" {
			return nil
		}
		return []Span{{
			Start:    0,
			End:      len(line),
			Label:    label,
			RuleName: rule.Name,
		}}
	}

	if len(rule.Groups) > 0 {
		spans := make([]Span, 0, len(rule.Groups))
		for group, label := range rule.Groups {
			startIndex := group * 2
			endIndex := startIndex + 1
			if endIndex >= len(indices) {
				continue
			}
			start, end := indices[startIndex], indices[endIndex]
			if start < 0 || end < 0 {
				continue
			}
			spans = append(spans, Span{
				Start:    start,
				End:      end,
				Label:    label,
				RuleName: rule.Name,
				Group:    group,
			})
		}
		if len(spans) > 0 {
			sort.Slice(spans, func(i, j int) bool {
				return spans[i].Group < spans[j].Group
			})
			return spans
		}
	}

	label := rule.Style
	if label == "" {
		label = rule.Name
	}
	if label == "" {
		return nil
	}

	return []Span{{
		Start:    indices[0],
		End:      indices[1],
		Label:    label,
		RuleName: rule.Name,
	}}
}
