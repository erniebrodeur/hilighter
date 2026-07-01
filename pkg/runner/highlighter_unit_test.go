package runner

import "testing"

func TestHighlighterProcessLineForBuiltInSyslog(t *testing.T) {
	highlighter, err := NewHighlighter("", "syslog", "")
	if err != nil {
		t.Fatalf("new highlighter: %v", err)
	}
	defer highlighter.Close()

	output := highlighter.ProcessLine(`Jun 30 21:16:33 Ernies-MacBook-Pro syslogd[140] <Error>: something bad happened`)

	if output == `Jun 30 21:16:33 Ernies-MacBook-Pro syslogd[140] <Error>: something bad happened` {
		t.Fatalf("expected highlighted output, got passthrough")
	}
	if !containsANSI(output) {
		t.Fatalf("expected ANSI sequences in %q", output)
	}
}

func TestHighlighterProcessLineForBuiltInDocker(t *testing.T) {
	highlighter, err := NewHighlighter("", "docker", "")
	if err != nil {
		t.Fatalf("new highlighter: %v", err)
	}
	defer highlighter.Close()

	output := highlighter.ProcessLine(`api-1  | 2026-06-30T22:13:37Z Error response from daemon: pull access denied`)

	if output == `api-1  | 2026-06-30T22:13:37Z Error response from daemon: pull access denied` {
		t.Fatalf("expected highlighted output, got passthrough")
	}
	if !containsANSI(output) {
		t.Fatalf("expected ANSI sequences in %q", output)
	}
}

func containsANSI(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '\x1b' && s[i+1] == '[' {
			return true
		}
	}
	return false
}
