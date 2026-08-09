package engine_test

import (
	"github.com/erniebrodeur/hilighter/pkg/engine"
	"github.com/erniebrodeur/hilighter/pkg/rules"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Engine", func() {
	It("highlights every occurrence without changing the source line", func() {
		compiled, err := rules.Compile([]rules.Spec{
			{Name: "email", Pattern: `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`, Style: "accent"},
			{Name: "ip", Pattern: `\b\d{1,3}(?:\.\d{1,3}){3}\b`, Style: "endpoint"},
			{Name: "error", Pattern: `\bERROR\b`, Style: "error"},
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { rules.Close(compiled) })

		line := "ERROR a@example.com 10.0.0.1 b@example.com ERROR 10.0.0.2"
		result := engine.New(compiled).ProcessLine(line)

		Expect(result.Line).To(Equal(line))
		Expect(result.Spans).To(HaveLen(6))
		Expect(result.Spans).To(HaveExactElements(
			engine.Span{Start: 0, End: 5, Label: "error", RuleName: "error"},
			engine.Span{Start: 6, End: 19, Label: "accent", RuleName: "email"},
			engine.Span{Start: 20, End: 28, Label: "endpoint", RuleName: "ip"},
			engine.Span{Start: 29, End: 42, Label: "accent", RuleName: "email"},
			engine.Span{Start: 43, End: 48, Label: "error", RuleName: "error"},
			engine.Span{Start: 49, End: 57, Label: "endpoint", RuleName: "ip"},
		))
	})

	It("maps capture groups for every occurrence", func() {
		compiled, err := rules.Compile([]rules.Spec{
			{
				Name:    "user",
				Pattern: `user=([a-z]+)`,
				Groups:  map[string]string{"1": "detail"},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { rules.Close(compiled) })

		result := engine.New(compiled).ProcessLine("user=alice user=bob")

		Expect(result.Spans).To(HaveExactElements(
			engine.Span{Start: 5, End: 10, Label: "detail", RuleName: "user", Group: 1},
			engine.Span{Start: 16, End: 19, Label: "detail", RuleName: "user", Group: 1},
		))
	})

	It("applies rules in declaration order when multiple patterns overlap", func() {
		compiled, err := rules.Compile([]rules.Spec{
			{Name: "full", Pattern: "ERROR", Style: "error"},
			{Name: "partial", Pattern: "ERR|WARN", Style: "warning"},
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { rules.Close(compiled) })

		result := engine.New(compiled).ProcessLine("ERROR WARN ERROR")

		Expect(result.Spans).To(HaveExactElements(
			engine.Span{Start: 0, End: 5, Label: "error", RuleName: "full"},
			engine.Span{Start: 6, End: 10, Label: "warning", RuleName: "partial"},
			engine.Span{Start: 11, End: 16, Label: "error", RuleName: "full"},
		))
	})

	It("defaults rules to substring scope when no scope is set", func() {
		compiled, err := rules.Compile([]rules.Spec{
			{Name: "error-word", Pattern: "ERROR", Style: "error"},
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { rules.Close(compiled) })

		result := engine.New(compiled).ProcessLine("prefix ERROR suffix")

		Expect(result.Spans).To(HaveLen(1))
		Expect(result.Spans[0].Start).To(Equal(7))
		Expect(result.Spans[0].End).To(Equal(12))
	})

	It("supports whole-line scope without changing the source text content", func() {
		compiled, err := rules.Compile([]rules.Spec{
			{Name: "warning-line", Pattern: "^warning:", Scope: rules.ScopeLine, Style: "warning"},
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { rules.Close(compiled) })

		line := "warning: disk low"
		result := engine.New(compiled).ProcessLine(line)

		Expect(result.Line).To(Equal(line))
		Expect(result.Spans).To(HaveLen(1))
		Expect(result.Spans[0].Start).To(Equal(0))
		Expect(result.Spans[0].End).To(Equal(len(line)))
	})

	It("maps capture groups to semantic labels for downstream rendering", func() {
		compiled, err := rules.Compile([]rules.Spec{
			{
				Name:    "panic",
				Pattern: `(panic:)(.*)$`,
				Groups: map[string]string{
					"1": "error",
					"2": "detail",
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { rules.Close(compiled) })

		result := engine.New(compiled).ProcessLine("panic: boom")

		Expect(result.Spans).To(HaveLen(2))
		Expect(result.Spans[0].Label).To(Equal("error"))
		Expect(result.Spans[0].Start).To(Equal(0))
		Expect(result.Spans[0].End).To(Equal(6))
		Expect(result.Spans[1].Label).To(Equal("detail"))
		Expect(result.Spans[1].Start).To(Equal(6))
		Expect(result.Spans[1].End).To(Equal(11))
	})

	It("processes input line by line", func() {
		compiled, err := rules.Compile([]rules.Spec{
			{Name: "error-word", Pattern: "ERROR", Style: "error"},
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { rules.Close(compiled) })

		results := engine.New(compiled).ProcessText("ok\nERROR\nfine\n")

		Expect(results).To(HaveLen(3))
		Expect(results[0].Line).To(Equal("ok"))
		Expect(results[0].Spans).To(BeEmpty())
		Expect(results[1].Line).To(Equal("ERROR"))
		Expect(results[1].Spans).To(HaveLen(1))
		Expect(results[2].Line).To(Equal("fine"))
	})
})
