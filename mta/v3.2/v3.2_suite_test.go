package v3_2

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestV3_2(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V3.2 Suite")
}