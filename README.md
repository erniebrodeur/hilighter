# hilighter

`hilighter` is a Go CLI for colorizing streamed text with user-defined regex rules. It is built for `stdout` and `stderr` style workflows where you want expressive highlighting without baking ANSI escape codes into the source text.

## Install

Install the CLI:

```bash
go install github.com/erniebrodeur/hilighter/cmd/hilighter@latest
```

If you want the module in another Go project:

```bash
go get github.com/erniebrodeur/hilighter
```

## Configuration

The default per-user config directory is:

```bash
~/.hilighter
```

The intended layout is:

```text
~/.hilighter/
├── config.yaml
├── rules.yaml
└── themes/
    └── default.yaml
```

`config.yaml` is the place for default overrides such as which rules file or theme file should be used by default.

Example:

```yaml
rules: ~/.hilighter/rules.yaml
theme: ~/.hilighter/themes/default.yaml
```

You can also point the CLI at explicit files with flags such as `--rules`, `--theme`, or override the base config directory with `--config-dir`.

Built-in app profiles can be selected with `--app` when you want shipped defaults instead of your own rule file.
Profiles may also carry a default `cmd`, so `hilighter --app syslog` can execute the associated command directly when you do not pass `--cmd`.
The `docker` profile also carries a default command, so `hilighter --app docker` defaults to `docker compose up`.
Current built-in profiles include `syslog`, `brew`, `docker`, `go-test`, `compiler`, and `logs`.

## What It Does

- reads text from a stream or command output
- matches lines against ordered regex rules
- applies ANSI styling to either matched substrings or whole lines
- keeps the original text unchanged apart from terminal rendering
- separates rule logic from theme styling

## Intended Usage

```bash
some-command 2>&1 | hilighter --rules ~/.hilighter/rules.yaml
hilighter --cmd "some-command 2>&1" --rules ~/.hilighter/rules.yaml
hilighter --app syslog
hilighter --app docker
brew install wget 2>&1 | hilighter --app brew
docker compose up 2>&1 | hilighter --app docker
```

Pipe mode is the primary use case. Command mode exists for convenience.

When both are available, an explicit `--rules` file takes precedence over `--app`.
When `--cmd` is omitted, an app profile or rules file can supply a default command through a top-level `cmd` value.

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

## Testing

The project uses Ginkgo and Gomega for BDD-style package tests.

```bash
go test ./...
```

If you want named spec output instead of only package-level detail, use:

```bash
ginkgo -v ./...
```

If you want to stay on `go test`, pass the Ginkgo verbosity flag through:

```bash
go test ./... -v -ginkgo.v
```
