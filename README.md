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

Tail mode with a named profile:

```bash
hilighter tail rails-log
hilighter tail rails-log log/production.log
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

## Tail Mode

`tail` is intended for saved log profiles.

```bash
hilighter tail <profile> [file]
```

Examples:

```bash
hilighter tail rails-log
hilighter tail rails-log log/production.log
hilighter tail docker-info /var/log/docker.log
```

Resolution rules:

- the second argument is the profile name
- the optional third argument is the file to follow
- if the third argument is omitted, the profile's `file` value is used
- relative file paths are resolved from `./`
- absolute file paths are used as-is

## Theme Notes

The default theme is terminal-friendly and restrained:

- errors are high contrast
- warnings are easy to spot
- Docker hostnames, IPs, and ports are emphasized

You can override any of that with your own theme YAML.
