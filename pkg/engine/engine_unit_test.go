package engine

import (
	"testing"

	"github.com/erniebrodeur/hilighter/pkg/rules"
)

func TestProcessLineCombinesNonOverlappingRuleSpans(t *testing.T) {
	compiled, err := rules.Compile([]rules.Spec{
		{
			Name:    "syslog-structured",
			Pattern: `^([A-Z][a-z]{2}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+([A-Za-z0-9_.-]+)\[\d+\]\s+(<(?:Error|Notice)>):\s+(.*)$`,
			Groups: map[string]string{
				"1": "timestamp",
				"2": "host",
				"3": "process",
			},
		},
		{
			Name:    "syslog-error-severity",
			Pattern: `<Error>`,
			Style:   "error",
		},
	})
	if err != nil {
		t.Fatalf("compile rules: %v", err)
	}
	defer rules.Close(compiled)

	result := New(compiled).ProcessLine(`Jun 30 21:16:33 Ernies-MacBook-Pro syslogd[140] <Error>: something bad happened`)

	if len(result.Spans) != 4 {
		t.Fatalf("expected 4 spans, got %d: %#v", len(result.Spans), result.Spans)
	}
	labels := []string{result.Spans[0].Label, result.Spans[1].Label, result.Spans[2].Label, result.Spans[3].Label}
	want := []string{"timestamp", "host", "process", "error"}
	for i := range want {
		if labels[i] != want[i] {
			t.Fatalf("span %d label = %q, want %q", i, labels[i], want[i])
		}
	}
}

func TestProcessLineKeepsEarlierRuleOnOverlap(t *testing.T) {
	compiled, err := rules.Compile([]rules.Spec{
		{Name: "full", Pattern: `ERROR`, Style: "error"},
		{Name: "partial", Pattern: `ERR`, Style: "warning"},
	})
	if err != nil {
		t.Fatalf("compile rules: %v", err)
	}
	defer rules.Close(compiled)

	result := New(compiled).ProcessLine("ERROR")

	if len(result.Spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(result.Spans))
	}
	if result.Spans[0].Label != "error" {
		t.Fatalf("label = %q, want error", result.Spans[0].Label)
	}
}
