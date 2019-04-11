package version

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
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
		yaml.Unmarshal([]byte("cli_version:5.2"), &VersionConfig)
		Ω(GetVersion()).Should(Equal(Version{CliVersion: "5.2", MakeFile: "10.5.3"}))
	})
})
