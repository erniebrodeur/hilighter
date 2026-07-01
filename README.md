# hilighter

`hilighter` is a Go CLI for colorizing streamed text with user-defined regex rules. It is built for `stdout` and `stderr` style workflows where you want expressive highlighting without baking ANSI escape codes into the source text.

## What It Does

- reads text from a stream or command output
- matches lines against ordered regex rules
- applies ANSI styling to either matched substrings or whole lines
- keeps the original text unchanged apart from terminal rendering
- separates rule logic from theme styling

## Intended Usage

```bash
some-command 2>&1 | hilighter --rules examples/rules/go-test.yaml
hilighter --cmd "some-command 2>&1" --rules examples/rules/go-test.yaml
```

Pipe mode is the primary use case. Command mode exists for convenience.

## Design Direction

- implemented in Go
- YAML-based configuration
- line-by-line processing
- first-declared rule wins
- highlight-only behavior
- ANSI output in v1
- PCRE-compatible regex support
- semantic theme labels with a Monokai-style default theme

Rules can target either:

- a matched substring
- an entire line

Substring scope is the default when a rule does not specify scope. Rules may also style capture groups independently.

## Example Rule Shape

```yaml
rules:
  - name: error
    pattern: '(ERROR|FATAL):\s+(.*)'
    style: error

  - name: go-test-fail
    pattern: '^--- FAIL: (.+)$'
    groups:
      "1": test-name

  - name: full-line-warning
    pattern: '^warning:'
    scope: line
    style: warning
```

## Themes

Themes map semantic labels such as `error`, `warning`, or `test-name` to ANSI styles. The default theme direction is Monokai-style foreground colors without background changes.
