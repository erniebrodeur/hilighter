package rules

const (
	ipv4Octet = `(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)`
	ipv4      = `(?:` + ipv4Octet + `\.){3}` + ipv4Octet
	ipv6Hex   = `[0-9A-Fa-f]{1,4}`
	ipv6      = `(?:` +
		`(?:` + ipv6Hex + `:){7}` + ipv6Hex + `|` +
		`(?:` + ipv6Hex + `:){1,7}:|` +
		`(?:` + ipv6Hex + `:){1,6}:` + ipv6Hex + `|` +
		`(?:` + ipv6Hex + `:){1,5}(?::` + ipv6Hex + `){1,2}|` +
		`(?:` + ipv6Hex + `:){1,4}(?::` + ipv6Hex + `){1,3}|` +
		`(?:` + ipv6Hex + `:){1,3}(?::` + ipv6Hex + `){1,4}|` +
		`(?:` + ipv6Hex + `:){1,2}(?::` + ipv6Hex + `){1,5}|` +
		ipv6Hex + `:(?:(?::` + ipv6Hex + `){1,6})|` +
		`:(?:(?::` + ipv6Hex + `){1,7}|:)|` +
		`(?:` + ipv6Hex + `:){6}` + ipv4 + `|` +
		`::(?:ffff(?::0{1,4})?:)?` + ipv4 + `|` +
		`(?:` + ipv6Hex + `:){1,4}:` + ipv4 +
		`)`
)

// Default returns the restrained built-in rules used when no other rule
// source is selected.
func Default() File {
	return File{Rules: []Spec{
		{
			Name:    "default-url",
			Pattern: `(?i)(?<![a-z0-9+.-])([a-z][a-z0-9+.-]*://[^\s<>"']*?[^\s<>"'])(?=[.,;!?)}\]]*(?:\s|$))`,
			Groups:  map[string]string{"1": "endpoint"},
		},
		{
			Name:    "default-email",
			Pattern: `(?i)(?<![a-z0-9._%+-])[a-z0-9_%+-]+(?:\.[a-z0-9_%+-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z]{2,63}(?![a-z0-9._%+-])`,
			Style:   "accent",
		},
		{
			Name:    "default-ipv6",
			Pattern: `(?<![0-9A-Fa-f:.])` + ipv6 + `(?:%[0-9A-Za-z._-]+)?(?![0-9A-Fa-f:.])`,
			Style:   "endpoint",
		},
		{
			Name:    "default-ipv4",
			Pattern: `(?<![0-9.])` + ipv4 + `(?![0-9.])`,
			Style:   "endpoint",
		},
		{Name: "default-error", Pattern: `(?i)\b(?:error|fatal)\b`, Style: "error"},
		{Name: "default-warning", Pattern: `(?i)\bwarn(?:ing)?\b`, Style: "warning"},
		{Name: "default-notice", Pattern: `(?i)\bnotice\b`, Style: "notice"},
		{Name: "default-info", Pattern: `(?i)\binfo\b`, Style: "info"},
		{Name: "default-detail", Pattern: `(?i)\b(?:trace|debug)\b`, Style: "detail"},
	}}
}
