package ansi_test

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("ANSI renderer", func() {
	It("renders styled substrings with resets around each highlighted span", Pending)
	It("renders whole-line styling without changing the underlying text", Pending)
	It("emits only foreground styling for the default theme", Pending)
	It("leaves unmatched text untouched", Pending)
})
