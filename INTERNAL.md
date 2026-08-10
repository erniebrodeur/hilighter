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

`internal/cli/version.go` owns version metadata. Module-aware builds use Go's embedded main-module version, which may be a release tag or VCS pseudo-version. An explicit `-ldflags` value takes precedence. Builds without usable module metadata use `dev`.

Before handing the tree to Git release operations:

```bash
go test -count=1 ./...
go vet ./...
go build -ldflags "-X github.com/erniebrodeur/hilighter/internal/cli.Version=1.0.0" -o hilighter ./cmd/hilighter
./hilighter --version
```

Also exercise a real ANSI-preserving pipeline and a downstream pipe closure. The latter must produce no diagnostic and must remain nonzero. Inspect tracked release content for secrets, personal paths, stale interface references, and generated binaries.

Git staging, commits, tags, pushes, and GitHub releases are separate maintainer actions.
