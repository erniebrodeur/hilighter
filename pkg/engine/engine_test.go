package engine_test

import (
	"github.com/erniebrodeur/hilighter/pkg/engine"
	"github.com/erniebrodeur/hilighter/pkg/rules"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Engine", func() {
	It("applies rules in declaration order when multiple patterns overlap", func() {
		compiled, err := rules.Compile([]rules.Spec{
			{Name: "full", Pattern: "ERROR", Style: "error"},
			{Name: "partial", Pattern: "ERR", Style: "warning"},
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { rules.Close(compiled) })

		result := engine.New(compiled).ProcessLine("ERROR")

		Expect(result.Spans).To(HaveLen(1))
		Expect(result.Spans[0].RuleName).To(Equal("full"))
		Expect(result.Spans[0].Label).To(Equal("error"))
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
