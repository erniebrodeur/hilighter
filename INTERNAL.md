# Internal Notes

This file contains durable maintainer context. It is not a changelog, release status, or task checkpoint.

## Product boundary

`hilighter` is a text-stream filter. It reads stdin or ordered text-file operands and writes the same text with ANSI styling added. It does not execute commands, emulate `cat` or `tail`, detect applications, manage profiles, invoke a pager, or process binary files.

The public grammar is intentionally small:

```text
hilighter [-e <pattern> ...] [-t <theme>] [--] [file ...]
hilighter --help
hilighter --version
```

Do not restore removed profiles, application packs, command execution, detection flags, or positional control commands through a compatibility layer.

## Behavioral invariants

- No file operands means stdin, including an interactive terminal.
- File operands suppress implicit stdin and are processed in order.
- `-` reads stdin at that position. `--` permits dash-prefixed filenames.
- Highlighting state resets at each operand boundary.
- Earlier rules win overlapping spans. Every other non-overlapping occurrence is highlighted.
- Existing ANSI bytes and all other source text are preserved.
- Input failures are reported immediately, later operands continue, and the final status is nonzero.
- Ordinary output failures stop immediately and remain diagnostic.
- A downstream broken pipe stops quietly but remains nonzero.
- Input behavior never depends on TTY detection.

## Rule and theme resolution

Rule sources replace rather than merge. Precedence is repeated `-e`, a non-empty `~/.hilighter/rules.yaml`, then compiled defaults. Repeated expressions use the `accent` style.

Theme precedence is `-t`, the selector in `~/.hilighter/config.yaml`, then compiled Monokai. Relative theme paths resolve from `~/.hilighter`. Custom themes overlay Monokai.

Built-in themes are selected through stable lowercase slugs in `pkg/theme`. Registry lookups return defensive copies. Theme colors support existing named colors and exact `#RRGGBB` values rendered as ANSI truecolor.

The shipped selectors are `abyss`, `dark-2026`, `dark-modern`, `dark-plus`, `high-contrast`, `kimbie-dark`, `monokai`, `monokai-dimmed`, `red`, `solarized-dark`, `tomorrow-night-blue`, and `visual-studio-dark`. These semantic palettes are adapted from the dark themes bundled with Visual Studio Code. Keep `THIRD_PARTY_NOTICES.md` when changing or redistributing them.

First run creates `config.yaml` containing only `theme: monokai`, an empty `rules.yaml`, and an empty `themes/` directory. Compiled defaults are not copied into the user directory. Legacy config and rule metadata are rejected rather than migrated.

## Package ownership

- `internal/cli`: public grammar, option resolution, process-level error policy, and version output
- `pkg/config`: fixed user customization layout and theme selector
- `pkg/rules`: strict rule loading, PCRE compilation, and shipped defaults
- `pkg/engine`: ordered match evaluation and non-overlapping spans
- `pkg/render/ansi`: semantic-style rendering
- `pkg/theme`: built-in and overlaid theme resolution
- `pkg/runner`: ordered input streaming and operand error behavior

Keep policy at the owning layer. In particular, do not move matching into the renderer or process-level error decisions into the stream runner.

## Release procedure

`internal/cli/version.go` owns version metadata and currently declares `1.0.1`. Update that source value for every release.

Before handing the tree to Git release operations:

```bash
go test -count=1 ./...
go vet ./...
go build -o hilighter ./cmd/hilighter
./hilighter --version
```

Also exercise a real ANSI-preserving pipeline and a downstream pipe closure. The latter must produce no diagnostic and must remain nonzero. Inspect tracked release content for secrets, personal paths, stale interface references, and generated binaries.

Git staging, commits, tags, and pushes are separate maintainer actions.

`.github/workflows/test.yml` enforces formatting, vetting, and tests on pushes and pull requests. `.github/workflows/release.yml` runs only after Test succeeds for a repository-owned `v*` tag push, verifies the tag, tested commit, and `internal/cli.Version` agree, creates a GitHub Release with generated notes, then requests the tagged module through `proxy.golang.org` so pkg.go.dev can index it. It intentionally publishes no binary artifacts; installation remains the standard Go module flow.
