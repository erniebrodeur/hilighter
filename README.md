# hilighter

`hilighter` adds useful ANSI color to plain or noisy terminal text while behaving like a small Unix filter.

It reads stdin or text files, highlights every non-overlapping match, and writes the original text with ANSI styling added. Existing ANSI sequences are preserved.

## Install

Requires Go 1.25 or newer.

```bash
go install github.com/erniebrodeur/hilighter/cmd/hilighter@latest
```

## Usage

```text
hilighter [-e <pattern> ...] [-t <theme>] [--] [file ...]
hilighter --help
hilighter --version
```

Highlight a command's output:

```bash
go test ./... 2>&1 | hilighter
```

Highlight one or more files in order:

```bash
hilighter application.log archive.log
```

With no file operands, `hilighter` reads stdin. A `-` operand reads stdin at that position, and `--` permits dash-prefixed filenames:

```bash
hilighter before.log - after.log
hilighter -- -server.log
```

Use repeated `-e` flags to replace the configured and built-in rules with PCRE expressions. Expressions are evaluated in order and use the `accent` style:

```bash
some-command | hilighter -e '(?i)failed' -e '\b[1-9][0-9]* errors?\b'
```

Select a built-in theme with `-t`:

```bash
some-command | hilighter -t dark-plus
```

Or select a custom theme file:

```bash
some-command | hilighter -t themes/solarized.yaml
```

Relative theme paths resolve from `~/.hilighter`.

## Built-in themes

Hilighter ships 12 dark themes adapted from the palettes bundled with Visual Studio Code:

| Theme | Selector |
| --- | --- |
| Abyss | `abyss` |
| Dark 2026 | `dark-2026` |
| Dark Modern | `dark-modern` |
| Dark+ | `dark-plus` |
| Default High Contrast | `high-contrast` |
| Kimbie Dark | `kimbie-dark` |
| Monokai, default | `monokai` |
| Monokai Dimmed | `monokai-dimmed` |
| Red | `red` |
| Solarized Dark | `solarized-dark` |
| Tomorrow Night Blue | `tomorrow-night-blue` |
| Visual Studio Dark | `visual-studio-dark` |

Use a selector with `-t` or as the `theme` value in `~/.hilighter/config.yaml`.

## Default highlighting

When no expressions or custom rules are selected, the shipped rules highlight:

- IPv4, IPv6, and MAC addresses
- extended ISO 8601 and classic syslog timestamps
- URLs containing a scheme such as `https://`
- email addresses
- `TRACE`, `DEBUG`, `INFO`, `NOTICE`, `WARN`, `WARNING`, `ERROR`, and `FATAL`, case-insensitively

Earlier rules win when matches overlap. Every other non-overlapping occurrence is highlighted.

## Configuration

On first run, `hilighter` creates:

```text
~/.hilighter/
├── config.yaml
├── rules.yaml
└── themes/
```

The initial `config.yaml` selects the compiled Monokai theme:

```yaml
theme: monokai
```

The initial `rules.yaml` is empty, which selects the shipped rules. A non-empty file replaces them. Rules use PCRE patterns and are evaluated in file order:

```yaml
rules:
  - name: failure
    pattern: '(?i)failed'
    style: error

  - name: request
    pattern: '(GET|POST)\s+(\S+)'
    groups:
      "1": info
      "2": endpoint

  - name: warning-line
    pattern: '(?i)warning'
    scope: line
    style: warning
```

`scope` may be `substring`, the default, or `line`. `groups` maps capture-group numbers to semantic style names. See [examples/rules/go-test.yaml](examples/rules/go-test.yaml) for another rule set.

Custom themes overlay the built-in theme, so only changed styles need to be specified:

```yaml
styles:
  error:
    fg: white
    bg: red
    bold: true
  endpoint:
    fg: cyan
```

Set the theme in `config.yaml` using a path relative to `~/.hilighter`:

```yaml
theme: themes/custom.yaml
```

Colors may be exact `#RRGGBB` truecolor values or the names `black`, `blue`, `cyan`, `gray`, `green`, `magenta`, `orange`, `pink`, `red`, `white`, and `yellow`. See [examples/themes/default.yaml](examples/themes/default.yaml) for a complete Monokai theme.

Rule-source precedence is repeated `-e`, then a non-empty `~/.hilighter/rules.yaml`, then shipped rules. Theme precedence is `-t`, then `config.yaml`, then compiled Monokai.

## Pipelines and errors

Preserve ANSI colors when paging:

```bash
some-command 2>&1 | hilighter | less -R
```

Input errors are written to stderr as they occur. Later file operands continue processing, and the final status is nonzero. Output errors stop processing immediately. A downstream broken pipe stops quietly without a redundant diagnostic and remains a nonzero exit.

`hilighter` processes text streams. Binary-file support is out of scope.

## Version

```bash
hilighter --version
```

Source builds report `hilighter-1.0.0`:

```bash
go build -o hilighter ./cmd/hilighter
./hilighter --version
```

Update `internal/cli/version.go` when preparing each release.

## License

Hilighter is licensed under [GPL-3.0-only](LICENSE). Palette data adapted from Visual Studio Code retains its upstream MIT notice in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
