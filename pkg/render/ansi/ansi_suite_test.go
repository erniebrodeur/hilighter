package ansi_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestANSI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ANSI Renderer Suite")
}
