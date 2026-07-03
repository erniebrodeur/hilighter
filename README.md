# hilighter

`hilighter` is a CLI for coloring command output with regex rules and themes.

It is meant for the normal terminal workflow:

- pipe output through it
- point it at your own rules
- or use a built-in profile like `docker`, `brew`, or `syslog`

## Install

```bash
go install github.com/erniebrodeur/hilighter/cmd/hilighter@latest
```

## Version

Print the app version marker:

```bash
hilighter --version
hilighter version
```

Release builds can stamp a version like `hilighter-1.0.0` with:

```bash
go build -ldflags "-X github.com/erniebrodeur/hilighter/internal/cli.Version=1.0.0" -o hilighter ./cmd/hilighter
```

## Configuration

The default config directory is:

```bash
~/.hilighter
```

Typical layout:

```text
~/.hilighter/
├── config.yaml
├── rules.yaml
└── themes/
    └── default.yaml
```

Example `config.yaml`:

```yaml
rules: ~/.hilighter/rules.yaml
theme: ~/.hilighter/themes/default.yaml

profiles:
  rails-log:
    rules: ~/.hilighter/rules/rails.yaml
    theme: ~/.hilighter/themes/default.yaml
    file: log/development.log

  docker-info:
    app: docker
```

Example `rules.yaml`:

```yaml
rules:
  - name: error
    pattern: '(ERROR|FATAL):\s+(.*)'
    style: error

  - name: warning
    pattern: '^warning:'
    scope: line
    style: warning

  - name: test-fail
    pattern: '^--- FAIL: (.+)$'
    groups:
      "1": test-name
```

If you want a custom theme file, start from [examples/themes/default.yaml](/Users/ebrodeur/Projects/hilighter/examples/themes/default.yaml).

## Usage

Pipe mode:

```bash
some-command 2>&1 | hilighter --rules ~/.hilighter/rules.yaml
```

Show help:

```bash
hilighter
```

Command mode:

```bash
hilighter --cmd "some-command 2>&1" --rules ~/.hilighter/rules.yaml
```

Built-in profiles:

```bash
hilighter --app docker
hilighter --app syslog
brew install wget 2>&1 | hilighter --app brew
docker info 2>&1 | hilighter --app docker
docker ps -a 2>&1 | hilighter --app docker
```

File modes:

```bash
hilighter tail rails-log
hilighter tail rails-log log/production.log
hilighter tail log/development.log
hilighter cat rails-log
hilighter cat log/development.log
hilighter head rails-log
hilighter head log/development.log
```

`hilighter tail rails-log` uses the profile's default file, resolved relative to your current working directory. For a Rails profile with `file: log/development.log`, that means:

```text
./log/development.log
```

If stdin is piped, `hilighter` reads the piped input instead of running an app profile's default command.

## Built-In Profiles

Current built-ins:

- `docker`
- `syslog`
- `brew`
- `go-test`
- `compiler`
- `logs`

Some app profiles include a default command when you run them directly:

- `docker` -> `docker ps -a`
- `syslog` -> `/usr/bin/log stream --style syslog`

## CLI Flags

```text
--app         built-in profile to use
--rules       path to a rules YAML file
--theme       path to a theme YAML file
--cmd         command to run through hilighter
--config-dir  config directory (default: ~/.hilighter)
```

## File Modes

Saved profiles can drive file-oriented commands directly, but these modes are
also meant to work as practical drop-in aliases for `tail`, `cat`, and `head`.

```bash
hilighter tail <profile> [file]
hilighter tail <file>
hilighter cat <profile> [file]
hilighter cat <file>
hilighter head <profile> [file]
hilighter head <file>
```

Examples:

```bash
hilighter tail rails-log
hilighter tail rails-log log/production.log
hilighter tail log/development.log
hilighter cat rails-log
hilighter cat log/development.log
hilighter head rails-log
hilighter head log/development.log
hilighter tail docker-info /var/log/docker.log
```

Resolution rules:

- if the first argument matches a saved profile, it is treated as the profile name
- otherwise the first argument is treated as the file target
- when using a profile, the optional third argument overrides the profile's `file` value
- if the third argument is omitted, the profile's `file` value is used
- relative file paths are resolved from `./`
- absolute file paths are used as-is
- for direct file targets, hilighter tries to auto-detect highlighting from matching saved profile file paths, `~/.hilighter/rules/*.yaml`, or obvious built-in names in the path
- if nothing matches, the command still runs as plain `tail`, `cat`, or `head`

Behavior:

- `tail` runs `tail -f`
- `cat` runs `cat`
- `head` runs `head`

## Theme Notes

The default theme is terminal-friendly and restrained:

- errors are high contrast
- warnings are easy to spot
- Docker hostnames, IPs, and ports are emphasized

You can override any of that with your own theme YAML.
