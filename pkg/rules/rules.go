package rules

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"go.elara.ws/pcre"
	"go.yaml.in/yaml/v3"
)

// Scope controls whether a rule styles only the matched substring or a full line.
type Scope string

const (
	// ScopeSubstring applies highlighting only to the matched substring.
	ScopeSubstring Scope = "substring"
	// ScopeLine applies highlighting to the entire line when the pattern matches.
	ScopeLine Scope = "line"
)

// File is the YAML shape for a rule file.
type File struct {
	// Command is an optional default shell command associated with the rule pack.
	Command string `yaml:"cmd,omitempty"`
	// Rules is the ordered list of rule definitions to compile.
	Rules []Spec `yaml:"rules"`
}

// Spec describes one rule as written in YAML.
type Spec struct {
	// Name is a stable identifier for diagnostics and testing.
	Name string `yaml:"name"`
	// Pattern is the PCRE-compatible regular expression to evaluate.
	Pattern string `yaml:"pattern"`
	// Scope defaults to substring when omitted.
	Scope Scope `yaml:"scope,omitempty"`
	// Style is the semantic label applied to the whole match or whole line.
	Style string `yaml:"style,omitempty"`
	// Groups maps capture group indexes to semantic labels.
	Groups map[string]string `yaml:"groups,omitempty"`
}

// Compiled is a rule that has been validated and compiled to a PCRE matcher.
type Compiled struct {
	Name    string
	Pattern string
	Scope   Scope
	Style   string
	Groups  map[int]string
	Regexp  *pcre.Regexp
}

// Load reads a rule YAML file from disk.
func Load(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}

	var file File
	if err := yaml.Unmarshal(data, &file); err != nil {
		return File{}, err
	}

	return file, nil
}

// Compile validates and compiles rule specs into PCRE-backed matchers.
func Compile(specs []Spec) ([]Compiled, error) {
	compiled := make([]Compiled, 0, len(specs))
	for i, spec := range specs {
		scope := spec.Scope
		if scope == "" {
			scope = ScopeSubstring
		}
		if scope != ScopeSubstring && scope != ScopeLine {
			closeAll(compiled)
			return nil, fmt.Errorf("rule %q has invalid scope %q", spec.Name, spec.Scope)
		}

		re, err := pcre.Compile(spec.Pattern)
		if err != nil {
			closeAll(compiled)
			name := spec.Name
			if name == "" {
				name = fmt.Sprintf("index %d", i)
			}
			return nil, fmt.Errorf("compile rule %q: %w", name, err)
		}

		groups, err := compileGroups(spec.Groups)
		if err != nil {
			re.Close()
			closeAll(compiled)
			return nil, fmt.Errorf("rule %q: %w", spec.Name, err)
		}

		compiled = append(compiled, Compiled{
			Name:    spec.Name,
			Pattern: spec.Pattern,
			Scope:   scope,
			Style:   spec.Style,
			Groups:  groups,
			Regexp:  re,
		})
	}

	return compiled, nil
}

// Builtins returns the built-in rule packs shipped with hilighter.
func Builtins() map[string]File {
	return map[string]File{
		"compiler": {
			Command: "go test ./...",
			Rules: []Spec{
				{Name: "compiler-error", Pattern: `(?i)\berror\b`, Style: "error"},
				{Name: "compiler-warning", Pattern: `(?i)\bwarning\b`, Style: "warning"},
			},
		},
		"brew": {
			Rules: []Spec{
				{Name: "brew-section", Pattern: `^(==>)\s+(.*)$`, Groups: map[string]string{"1": "accent", "2": "info"}},
				{Name: "brew-warning", Pattern: `^Warning:.*$`, Scope: ScopeLine, Style: "warning"},
				{Name: "brew-error", Pattern: `^Error:.*$`, Scope: ScopeLine, Style: "error"},
				{Name: "brew-url", Pattern: `(https?://\S+)`, Style: "accent"},
				{Name: "brew-pour", Pattern: `(?i)\b([A-Za-z0-9@+_.-]+)(?:--[A-Za-z0-9+_.-]+)?(?:\.bottle(?:\.[A-Za-z0-9_.-]+)?(?:\.tar\.gz)?)\b`, Groups: map[string]string{"1": "process"}},
				{Name: "brew-success", Pattern: `^🍺\s+.*$`, Scope: ScopeLine, Style: "info"},
			},
		},
		"docker": {
			Command: "docker ps -a",
			Rules: []Spec{
				{Name: "docker-compose-prefix", Pattern: `^([A-Za-z0-9_.-]+)(\s+\|\s+)(.*)$`, Groups: map[string]string{"1": "process"}},
				{Name: "docker-warning-line", Pattern: `^WARNING:.*$`, Scope: ScopeLine, Style: "warning"},
				{Name: "docker-name-host", Pattern: `^(\s*Name:)(\s+)([A-Za-z][A-Za-z0-9-]*)$`, Groups: map[string]string{"1": "accent", "3": "endpoint"}},
				{Name: "docker-key-endpoint", Pattern: `^([A-Za-z][A-Za-z0-9 /_-]*:)(\s+)((?:https?://)?(?:localhost|[A-Za-z][A-Za-z0-9-]*(?:\.[A-Za-z0-9-]+)*\.[A-Za-z][A-Za-z-]*)(?::\d+)?|(?:https?://)?\d{1,3}(?:\.\d{1,3}){3}(?::\d+)?|::1/\d+)$`, Groups: map[string]string{"1": "accent", "3": "endpoint"}},
				{Name: "docker-section-with-value", Pattern: `^([A-Za-z][A-Za-z0-9 /_-]*:)(\s+)(.+)$`, Groups: map[string]string{"1": "accent", "3": "info"}},
				{Name: "docker-indented-key-bool-true", Pattern: `^(\s+[A-Za-z][A-Za-z0-9 /._()-]*:)(\s+)(true)\b$`, Groups: map[string]string{"1": "accent", "3": "bool-true"}},
				{Name: "docker-indented-key-bool-false", Pattern: `^(\s+[A-Za-z][A-Za-z0-9 /._()-]*:)(\s+)(false)\b$`, Groups: map[string]string{"1": "accent", "3": "bool-false"}},
				{Name: "docker-indented-key-endpoint", Pattern: `^(\s+[A-Za-z][A-Za-z0-9 /._()-]*:)(\s+)((?:localhost|[A-Za-z][A-Za-z0-9-]*(?:\.[A-Za-z0-9-]+)*\.[A-Za-z][A-Za-z-]*)(?::\d+)?|\d{1,3}(?:\.\d{1,3}){3}(?::\d+)?(?:/\d+)?|::1/\d+)$`, Groups: map[string]string{"1": "accent", "3": "endpoint"}},
				{Name: "docker-indented-key-value", Pattern: `^(\s+[A-Za-z][A-Za-z0-9 /._()-]*:)(\s+)(.+)$`, Groups: map[string]string{"1": "accent", "3": "detail"}},
				{Name: "docker-endpoint", Pattern: `\b((?:https?://)?(?:localhost|[A-Za-z][A-Za-z0-9-]*(?:\.[A-Za-z0-9-]+)*\.[A-Za-z][A-Za-z-]*)(?::\d+)?|(?:https?://)?\d{1,3}(?:\.\d{1,3}){3}(?::\d+)?|::1/\d+)\b`, Groups: map[string]string{"1": "endpoint"}},
				{Name: "docker-indented-endpoint", Pattern: `^(\s+)((?:localhost|[A-Za-z][A-Za-z0-9-]*(?:\.[A-Za-z0-9-]+)*\.[A-Za-z][A-Za-z-]*)(?::\d+)?|\d{1,3}(?:\.\d{1,3}){3}(?::\d+)?(?:/\d+)?|::1/\d+)$`, Groups: map[string]string{"2": "endpoint"}},
				{Name: "docker-table-header", Pattern: `^(CONTAINER ID|IMAGE ID|NETWORK ID|REPOSITORY|DRIVER|NAME|TYPE)(\s{2,}.*)$`, Groups: map[string]string{"1": "accent", "2": "detail"}},
				{Name: "docker-timestamp", Pattern: `^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z?)`, Groups: map[string]string{"1": "timestamp"}},
				{Name: "docker-hex-id", Pattern: `\b([a-f0-9]{12,64})\b`, Groups: map[string]string{"1": "accent"}},
				{Name: "docker-image-reference", Pattern: `\b([A-Za-z0-9._/-]+:[A-Za-z0-9._-]+)\b`, Groups: map[string]string{"1": "accent"}},
				{Name: "docker-object-name", Pattern: `\s{2,}([A-Za-z][A-Za-z0-9_.-]*(?:/[A-Za-z0-9_.-]+)?(?:-[A-Za-z0-9_.-]+)*)$`, Groups: map[string]string{"1": "host"}},
				{Name: "docker-digest", Pattern: `(sha256:[a-f0-9]{12,64})`, Style: "accent"},
				{Name: "docker-layer-status", Pattern: `^([a-f0-9]{12}):\s+(Pulling fs layer|Waiting|Downloading|Verifying Checksum|Download complete|Extracting|Pull complete|Already exists)$`, Groups: map[string]string{"1": "accent", "2": "info"}},
				{Name: "docker-bool-true", Pattern: `(?i)\b(true)\b`, Groups: map[string]string{"1": "bool-true"}},
				{Name: "docker-bool-false", Pattern: `(?i)\b(false)\b`, Groups: map[string]string{"1": "bool-false"}},
				{Name: "docker-size", Pattern: `\b(\d+(?:\.\d+)?(?:kB|MB|GB|GiB|MiB|B))\b`, Groups: map[string]string{"1": "detail"}},
				{Name: "docker-port-mapping", Pattern: `(\b\d{1,3}(?:\.\d{1,3}){3}:\d+(?:-\d+)?->\d+(?:-\d+)?/[a-z]+\b|\[::\]:\d+(?:-\d+)?->\d+(?:-\d+)?/[a-z]+)`, Style: "endpoint"},
				{Name: "docker-status", Pattern: `(?i)\b(Up|Exited|Running|Paused|Restarting|Dead|Created|Started|Healthy|inactive|local|Downloaded newer image)\b`, Style: "info"},
				{Name: "docker-none", Pattern: `<none>`, Style: "warning"},
				{Name: "docker-warning", Pattern: `(?i)\bwarning\b`, Style: "warning"},
				{Name: "docker-error", Pattern: `(?i)\b(error|denied|unauthorized|failed|exited with code)\b`, Style: "error"},
			},
		},
		"go-test": {
			Command: "go test ./...",
			Rules: []Spec{
				{Name: "fail-line", Pattern: `^--- FAIL: (.+)$`, Groups: map[string]string{"1": "test-name"}},
				{Name: "panic", Pattern: `(panic:)(.*)$`, Groups: map[string]string{"1": "error", "2": "detail"}},
			},
		},
		"logs": {
			Rules: []Spec{
				{Name: "log-error", Pattern: `(?i)\berror\b`, Style: "error"},
				{Name: "log-warn", Pattern: `(?i)\bwarn(?:ing)?\b`, Style: "warning"},
			},
		},
		"syslog": {
			Command: "/usr/bin/log stream --style syslog",
			Rules: []Spec{
				{
					Name:    "syslog-log-stream",
					Pattern: `^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}\.\d{6}[+-]\d{4})\s+(\S+)\s+([A-Za-z0-9_.-]+)\[\d+\]:\s+(.*)$`,
					Groups: map[string]string{
						"1": "timestamp",
						"2": "host",
						"3": "process",
					},
				},
				{
					Name:    "syslog-error-structured",
					Pattern: `^([A-Z][a-z]{2}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+([A-Za-z0-9_.-]+)\[\d+\]\s+(<(?i:(?:emerg(?:ency)?|alert|crit(?:ical)?|err(?:or)?))>):\s+(.*)$`,
					Groups: map[string]string{
						"1": "timestamp",
						"2": "host",
						"3": "process",
						"4": "error",
					},
				},
				{
					Name:    "syslog-warning-structured",
					Pattern: `^([A-Z][a-z]{2}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+([A-Za-z0-9_.-]+)\[\d+\]\s+(<(?i:warn(?:ing)?)>):\s+(.*)$`,
					Groups: map[string]string{
						"1": "timestamp",
						"2": "host",
						"3": "process",
						"4": "warning",
					},
				},
				{
					Name:    "syslog-notice-structured",
					Pattern: `^([A-Z][a-z]{2}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+([A-Za-z0-9_.-]+)\[\d+\]\s+(<(?i:notice)>):\s+(.*)$`,
					Groups: map[string]string{
						"1": "timestamp",
						"2": "host",
						"3": "process",
						"4": "notice",
					},
				},
				{
					Name:    "syslog-info-structured",
					Pattern: `^([A-Z][a-z]{2}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+([A-Za-z0-9_.-]+)\[\d+\]\s+(<(?i:(?:info|debug))>):\s+(.*)$`,
					Groups: map[string]string{
						"1": "timestamp",
						"2": "host",
						"3": "process",
						"4": "info",
					},
				},
				{Name: "syslog-message-error", Pattern: `(?i)\berror\b`, Style: "error"},
				{Name: "syslog-message-warning", Pattern: `(?i)\bwarn(?:ing)?\b`, Style: "warning"},
				{Name: "syslog-repeat", Pattern: `^(--- last message repeated \d+ times ---)$`, Groups: map[string]string{"1": "repeat"}},
			},
		},
	}
}

// Builtin returns one built-in rule pack by name.
func Builtin(name string) (File, bool) {
	file, ok := Builtins()[strings.ToLower(name)]
	return file, ok
}

// ErrUnknownBuiltin returns an error describing an unknown built-in rule pack.
func ErrUnknownBuiltin(name string) error {
	return fmt.Errorf("unknown built-in app profile %q", name)
}

// Close releases the PCRE resources held by compiled rules.
func Close(compiled []Compiled) {
	closeAll(compiled)
}

func closeAll(compiled []Compiled) {
	for _, rule := range compiled {
		if rule.Regexp != nil {
			runtime.SetFinalizer(rule.Regexp, nil)
			rule.Regexp.Close()
		}
	}
}

func compileGroups(groups map[string]string) (map[int]string, error) {
	if len(groups) == 0 {
		return nil, nil
	}

	keys := make([]int, 0, len(groups))
	compiled := make(map[int]string, len(groups))
	for key, label := range groups {
		index, err := strconv.Atoi(key)
		if err != nil || index < 1 {
			return nil, fmt.Errorf("invalid group index %q", key)
		}
		keys = append(keys, index)
		compiled[index] = label
	}
	sort.Ints(keys)
	return compiled, nil
}
