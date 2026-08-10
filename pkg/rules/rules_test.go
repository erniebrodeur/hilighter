package rules_test

import (
	"os"
	"path/filepath"

	"github.com/erniebrodeur/hilighter/pkg/rules"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rules", func() {
	It("loads and compiles PCRE-compatible patterns", func() {
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

	It("rejects invalid patterns with useful errors", func() {
		_, err := rules.Compile([]rules.Spec{{Name: "broken", Pattern: "(", Style: "error"}})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("broken"))
	})

	It("rejects removed command metadata", func() {
		path := filepath.Join(GinkgoT().TempDir(), "rules.yaml")
		Expect(os.WriteFile(path, []byte("cmd: go test ./...\nrules: []\n"), 0o644)).To(Succeed())

		_, err := rules.Load(path)

		Expect(err).To(HaveOccurred())
	})
})
