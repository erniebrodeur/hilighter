package rules_test

import (
	"github.com/erniebrodeur/hilighter/pkg/engine"
	"github.com/erniebrodeur/hilighter/pkg/rules"
	"github.com/erniebrodeur/hilighter/pkg/theme"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Default rules", func() {
	var (
		compiled []rules.Compiled
		matcher  *engine.Engine
	)

	BeforeEach(func() {
		var err error
		compiled, err = rules.Compile(rules.Default().Rules)
		Expect(err).NotTo(HaveOccurred())
		matcher = engine.New(compiled)
	})

	AfterEach(func() {
		rules.Close(compiled)
	})

	It("uses the approved semantic labels", func() {
		styles := map[string]string{}
		for _, rule := range rules.Default().Rules {
			styles[rule.Name] = rule.Style
			for _, style := range rule.Groups {
				styles[rule.Name] = style
			}
		}

		Expect(styles).To(Equal(map[string]string{
			"default-url":     "endpoint",
			"default-email":   "accent",
			"default-ipv6":    "endpoint",
			"default-ipv4":    "endpoint",
			"default-error":   "error",
			"default-warning": "warning",
			"default-notice":  "notice",
			"default-info":    "info",
			"default-detail":  "detail",
		}))

		monokai := theme.Default()
		for _, label := range styles {
			_, ok := monokai.Resolve(label)
			Expect(ok).To(BeTrue(), "missing Monokai style %q", label)
		}
	})

	DescribeTable("highlights representative tokens",
		func(line string, expected []highlight) {
			Expect(highlights(matcher.ProcessLine(line))).To(Equal(expected))
		},
		Entry("scheme-qualified URLs without prose punctuation",
			"See (https://example.com/a?q=1), then git+ssh://host/repo.",
			[]highlight{{"https://example.com/a?q=1", "endpoint"}, {"git+ssh://host/repo", "endpoint"}},
		),
		Entry("ASCII email addresses",
			"mail first.last+tag@example.co.uk or ops@example.com",
			[]highlight{{"first.last+tag@example.co.uk", "accent"}, {"ops@example.com", "accent"}},
		),
		Entry("valid IPv4 boundaries",
			"listen 0.0.0.0 and contact 255.255.255.255 or 192.168.1.10",
			[]highlight{{"0.0.0.0", "endpoint"}, {"255.255.255.255", "endpoint"}, {"192.168.1.10", "endpoint"}},
		),
		Entry("full, compressed, loopback, mapped, and zoned IPv6",
			"2001:db8:85a3:0:0:8a2e:370:7334 2001:db8::1 :: ::1 ::ffff:192.0.2.128 fe80::1%eth0",
			[]highlight{{"2001:db8:85a3:0:0:8a2e:370:7334", "endpoint"}, {"2001:db8::1", "endpoint"}, {"::", "endpoint"}, {"::1", "endpoint"}, {"::ffff:192.0.2.128", "endpoint"}, {"fe80::1%eth0", "endpoint"}},
		),
		Entry("case-insensitive approved log levels",
			"trace DEBUG Info notice WARN warning Error fatal",
			[]highlight{{"trace", "detail"}, {"DEBUG", "detail"}, {"Info", "info"}, {"notice", "notice"}, {"WARN", "warning"}, {"warning", "warning"}, {"Error", "error"}, {"fatal", "error"}},
		),
	)

	DescribeTable("leaves false positives plain",
		func(line string) {
			result := matcher.ProcessLine(line)
			Expect(result.Line).To(Equal(line))
			Expect(result.Spans).To(BeEmpty())
		},
		Entry("bare web addresses", "example.com www.example.com"),
		Entry("invalid IPv4", "256.1.1.1 1.2.3.999 01.2.3.4"),
		Entry("invalid IPv6", "2001:::1 2001:db8:1:2:3:4:5:6:7"),
		Entry("invalid email", "user@localhost a..b@example.com user@-example.com"),
		Entry("larger words and excluded levels", "information warningly fatalism error_code PANIC CRITICAL ALERT EMERGENCY"),
	)
})

type highlight struct {
	text  string
	label string
}

func highlights(result engine.Result) []highlight {
	out := make([]highlight, 0, len(result.Spans))
	for _, span := range result.Spans {
		out = append(out, highlight{
			text:  result.Line[span.Start:span.End],
			label: span.Label,
		})
	}
	return out
}
