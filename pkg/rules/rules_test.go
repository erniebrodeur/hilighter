package rules_test

import (
	"os"
	"path/filepath"

	"github.com/erniebrodeur/hilighter/pkg/rules"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rules", func() {
	It("compiles PCRE-compatible patterns from YAML config", func() {
		path := filepath.Join(GinkgoT().TempDir(), "rules.yaml")
		Expect(os.WriteFile(path, []byte("rules:\n  - name: lookahead\n    pattern: '\\d+(?= USD)'\n    style: money\n"), 0o644)).To(Succeed())

		file, err := rules.Load(path)
		Expect(err).NotTo(HaveOccurred())

		compiled, err := rules.Compile(file.Rules)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { rules.Close(compiled) })

		Expect(compiled).To(HaveLen(1))
		Expect(compiled[0].Regexp.MatchString("9000 USD")).To(BeTrue())
		Expect(compiled[0].Regexp.MatchString("9000 EUR")).To(BeFalse())
	})

	It("rejects invalid regex patterns with useful errors", func() {
		_, err := rules.Compile([]rules.Spec{{
			Name:    "broken",
			Pattern: "(",
			Style:   "error",
		}})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("broken"))
	})

	It("ships built-in rule packs for common command output", func() {
		builtins := rules.Builtins()

		Expect(builtins).To(HaveKey("go-test"))
		Expect(builtins).To(HaveKey("logs"))
		Expect(builtins).To(HaveKey("compiler"))
		Expect(builtins).To(HaveKey("syslog"))
		Expect(builtins).To(HaveKey("brew"))
		Expect(builtins).To(HaveKey("docker"))
	})

	It("resolves built-in packs by app name", func() {
		file, ok := rules.Builtin("syslog")

		Expect(ok).To(BeTrue())
		Expect(file.Command).To(Equal("/usr/bin/log stream --style syslog"))
		Expect(file.Rules).NotTo(BeEmpty())
	})

	It("ships a structured syslog profile for timestamp, host, process, and repeat lines", func() {
		file, ok := rules.Builtin("syslog")
		Expect(ok).To(BeTrue())

		compiled, err := rules.Compile(file.Rules)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { rules.Close(compiled) })

		var foundLogStream bool
		var foundNotice bool
		var foundError bool
		for _, rule := range compiled {
			switch rule.Name {
			case "syslog-log-stream":
				foundLogStream = true
				indices := rule.Regexp.FindStringSubmatchIndex(`2026-06-30 22:13:37.018014-0700  localhost WindowServer[171]: (IOHIDNXEventTranslatorSessionFilter) [com.apple.iohid:activity] displayStateFilter policy:0`)
				Expect(indices).NotTo(BeEmpty())
				Expect(rule.Groups[1]).To(Equal("timestamp"))
				Expect(rule.Groups[2]).To(Equal("host"))
				Expect(rule.Groups[3]).To(Equal("process"))
			case "syslog-notice-structured":
				foundNotice = true
				indices := rule.Regexp.FindStringSubmatchIndex(`Jun 30 19:18:52 Ernies-MacBook-Pro login[99495] <Notice>: USER_PROCESS: 99495 ttys004`)
				Expect(indices).NotTo(BeEmpty())
				Expect(rule.Groups[1]).To(Equal("timestamp"))
				Expect(rule.Groups[2]).To(Equal("host"))
				Expect(rule.Groups[3]).To(Equal("process"))
				Expect(rule.Groups[4]).To(Equal("notice"))
			case "syslog-error-structured":
				foundError = true
				indices := rule.Regexp.FindStringSubmatchIndex(`Jun 30 21:16:33 Ernies-MacBook-Pro syslogd[140] <Error>: something bad happened`)
				Expect(indices).NotTo(BeEmpty())
				Expect(rule.Groups[4]).To(Equal("error"))
			}
		}

		Expect(foundLogStream).To(BeTrue())
		Expect(foundNotice).To(BeTrue())
		Expect(foundError).To(BeTrue())
	})

	It("ships a brew profile for sections, warnings, errors, and package artifacts", func() {
		file, ok := rules.Builtin("brew")
		Expect(ok).To(BeTrue())

		compiled, err := rules.Compile(file.Rules)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { rules.Close(compiled) })

		var foundSection bool
		var foundWarning bool
		var foundError bool
		var foundPour bool
		for _, rule := range compiled {
			switch rule.Name {
			case "brew-section":
				foundSection = true
				indices := rule.Regexp.FindStringSubmatchIndex("==> Downloading https://ghcr.io/v2/homebrew/core/wget/manifests/1.25.0")
				Expect(indices).NotTo(BeEmpty())
				Expect(rule.Groups[1]).To(Equal("accent"))
				Expect(rule.Groups[2]).To(Equal("info"))
			case "brew-warning":
				foundWarning = true
				Expect(rule.Regexp.MatchString("Warning: Treating foo as a formula.")).To(BeTrue())
			case "brew-error":
				foundError = true
				Expect(rule.Regexp.MatchString("Error: No available formula with the name \"foo\"")).To(BeTrue())
			case "brew-pour":
				foundPour = true
				indices := rule.Regexp.FindStringSubmatchIndex("wget--1.25.0.arm64_sequoia.bottle.tar.gz")
				Expect(indices).NotTo(BeEmpty())
				Expect(rule.Groups[1]).To(Equal("process"))
			}
		}

		Expect(foundSection).To(BeTrue())
		Expect(foundWarning).To(BeTrue())
		Expect(foundError).To(BeTrue())
		Expect(foundPour).To(BeTrue())
	})

	It("ships a docker profile for compose prefixes, layer progress, digests, and failures", func() {
		file, ok := rules.Builtin("docker")
		Expect(ok).To(BeTrue())

		compiled, err := rules.Compile(file.Rules)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { rules.Close(compiled) })

		var foundPrefix bool
		var foundLayer bool
		var foundDigest bool
		var foundError bool
		for _, rule := range compiled {
			switch rule.Name {
			case "docker-compose-prefix":
				foundPrefix = true
				indices := rule.Regexp.FindStringSubmatchIndex("api-1  | 2026-06-30T22:13:37Z server started")
				Expect(indices).NotTo(BeEmpty())
				Expect(rule.Groups[1]).To(Equal("process"))
			case "docker-layer-status":
				foundLayer = true
				indices := rule.Regexp.FindStringSubmatchIndex("4f4fb700ef54: Pull complete")
				Expect(indices).NotTo(BeEmpty())
				Expect(rule.Groups[1]).To(Equal("accent"))
				Expect(rule.Groups[2]).To(Equal("info"))
			case "docker-digest":
				foundDigest = true
				Expect(rule.Regexp.MatchString("Digest: sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00f7e0c6d1c7d2b57")).To(BeTrue())
			case "docker-error":
				foundError = true
				Expect(rule.Regexp.MatchString("Error response from daemon: pull access denied")).To(BeTrue())
			}
		}

		Expect(foundPrefix).To(BeTrue())
		Expect(foundLayer).To(BeTrue())
		Expect(foundDigest).To(BeTrue())
		Expect(foundError).To(BeTrue())
	})
})
