package version

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Version Suite")
}

var _ = Describe("Version", func() {
	It("Sanity", func() {
		VersionConfig = []byte(`
cli_version: 5.2
makefile_version: 10.5.3
`)
		Ω(GetVersion()).Should(Equal(Version{CliVersion: "5.2", MakeFile: "10.5.3"}))
	})
})
