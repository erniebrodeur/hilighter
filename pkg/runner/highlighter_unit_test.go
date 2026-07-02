package runner

import (
	"strings"
	"testing"
)

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

func TestHighlighterProcessLineForBuiltInDockerBooleans(t *testing.T) {
	highlighter, err := NewHighlighter("", "docker", "")
	if err != nil {
		t.Fatalf("new highlighter: %v", err)
	}
	defer highlighter.Close()

	output := highlighter.ProcessLine(`  Supports d_type: true`)

	if !strings.Contains(output, "\x1b[1;38;5;148mtrue\x1b[0m") {
		t.Fatalf("expected green true highlight in %q", output)
	}
	output = highlighter.ProcessLine(`  Experimental: false`)
	if !strings.Contains(output, "\x1b[1;38;5;203mfalse\x1b[0m") {
		t.Fatalf("expected red false highlight in %q", output)
	}
}

func TestHighlighterProcessLineForBuiltInDockerTableDetails(t *testing.T) {
	highlighter, err := NewHighlighter("", "docker", "")
	if err != nil {
		t.Fatalf("new highlighter: %v", err)
	}
	defer highlighter.Close()

	output := highlighter.ProcessLine(`6f3e24ceeb52   bitnami/keycloak:latest   "/opt/bitnami/script…"   12 hours ago   Up 12 hours   0.0.0.0:8080->8080/tcp   keycloak-local`)

	if !strings.Contains(output, "\x1b[1;81mbitnami/keycloak:latest\x1b[0m") {
		t.Fatalf("expected image reference highlight in %q", output)
	}
	if !strings.Contains(output, "\x1b[1;38;5;148mkeycloak-local\x1b[0m") {
		t.Fatalf("expected object name highlight in %q", output)
	}
	if !strings.Contains(output, "\x1b[1;38;5;208m0.0.0.0:8080\x1b[0m->8080/tcp") {
		t.Fatalf("expected port mapping highlight in %q", output)
	}
}

func TestHighlighterProcessLineForBuiltInDockerEndpoints(t *testing.T) {
	highlighter, err := NewHighlighter("", "docker", "")
	if err != nil {
		t.Fatalf("new highlighter: %v", err)
	}
	defer highlighter.Close()

	output := highlighter.ProcessLine(` HTTP Proxy: http.docker.internal:3128`)
	if !strings.Contains(output, "\x1b[1;38;5;208mhttp.docker.internal:3128\x1b[0m") {
		t.Fatalf("expected endpoint highlight in %q", output)
	}

	output = highlighter.ProcessLine(`No Proxy: hubproxy.docker.internal`)
	if !strings.Contains(output, "\x1b[1;38;5;208mhubproxy.docker.internal\x1b[0m") {
		t.Fatalf("expected no-proxy hostname highlight in %q", output)
	}

	output = highlighter.ProcessLine(`  hubproxy.docker.internal:5555`)
	if !strings.Contains(output, "\x1b[1;38;5;208mhubproxy.docker.internal:5555\x1b[0m") {
		t.Fatalf("expected registry highlight in %q", output)
	}

	output = highlighter.ProcessLine(` Total Memory: 7.654GiB`)
	if strings.Contains(output, "\x1b[1;38;5;208m7.654GiB\x1b[0m") {
		t.Fatalf("did not expect memory value to be treated as endpoint in %q", output)
	}

	output = highlighter.ProcessLine(` ID: 8351a675-abac-45b1-9daf-c30f4e80e2b5`)
	if strings.Contains(output, "\x1b[1;38;5;208m8351a675-abac-45b1-9daf-c30f4e80e2b5\x1b[0m") {
		t.Fatalf("did not expect docker ID to be treated as endpoint in %q", output)
	}

	output = highlighter.ProcessLine(` Name: docker-desktop`)
	if !strings.Contains(output, "\x1b[1;38;5;208mdocker-desktop\x1b[0m") {
		t.Fatalf("expected docker host name highlight in %q", output)
	}

	output = highlighter.ProcessLine(` runc version: v1.2.5-0-g59923ef`)
	if strings.Contains(output, "\x1b[1;38;5;208mv1.2.5-0-g59923ef\x1b[0m") {
		t.Fatalf("did not expect version string to be treated as endpoint in %q", output)
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
