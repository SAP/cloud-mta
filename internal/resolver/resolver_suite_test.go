package resolver

import (
	"github.com/SAP/cloud-mta/internal/logs"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestResolver(t *testing.T) {
	logs.NewLogger()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Resolver Suite")
}

func getTestPath(relPath ...string) string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "testdata", filepath.Join(relPath...))
}
